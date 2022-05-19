/*
Copyright 2021 The KCP Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/kcp-dev/logicalcluster"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/tools/clusters"
	"k8s.io/klog/v2"

	tenancyv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/tenancy/v1alpha1"
	workloadv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/workload/v1alpha1"
	"github.com/kcp-dev/kcp/pkg/syncer/shared"
	conditionsapi "github.com/kcp-dev/kcp/third_party/conditions/apis/conditions/v1alpha1"
	"github.com/kcp-dev/kcp/third_party/conditions/util/conditions"
)

const (
	SchedulingDisabledLabel = "experimental.workloads.kcp.dev/scheduling-disabled"

	// The presence of `workloads.kcp.dev/schedulable: true` on a workspace
	// enables scheduling for the contents of the workspace. It is applied by
	// default to workspaces of type `Universal`.
	WorkspaceSchedulableLabel = "workloads.kcp.dev/schedulable"
)

// reconcileResource is responsible for setting the cluster for a resource of
// any type, to match the cluster where its namespace is assigned.
func (c *Controller) reconcileResource(ctx context.Context, lclusterName logicalcluster.Name, unstr *unstructured.Unstructured, gvr *schema.GroupVersionResource) error {
	if gvr.Group == "networking.k8s.io" && gvr.Resource == "ingresses" {
		klog.V(4).Infof("Skipping reconciliation of ingress %s/%s", unstr.GetNamespace(), unstr.GetName())
		return nil
	}

	klog.V(2).Infof("Reconciling GVR %q %s|%s/%s", gvr.String(), lclusterName, unstr.GetNamespace(), unstr.GetName())

	// If the resource is not namespaced (incl if the resource is itself a
	// namespace), ignore it.
	if unstr.GetNamespace() == "" {
		klog.V(4).Infof("GVR %q %s|%s had no namespace; ignoring", gvr.String(), unstr.GetClusterName(), unstr.GetName())
		return nil
	}

	// Align the resource's assigned cluster with the namespace's assigned
	// cluster.
	// First, get the namespace object (from the cached lister).
	ns, err := c.namespaceLister.Get(clusters.ToClusterAwareKey(lclusterName, unstr.GetNamespace()))
	if apierrors.IsNotFound(err) {
		// Namespace was deleted; this resource will eventually get deleted too, so ignore
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reconciling resource %s|%s/%s: error getting namespace: %w", lclusterName, unstr.GetNamespace(), unstr.GetName(), err)
	}

	if !scheduleRequirement.Matches(labels.Set(ns.Labels)) {
		// Do not schedule the resource transitively, and let external controllers
		// or users be responsible for it, consistently with the scheduling of the
		// namespace.
		return nil
	}

	lbls := unstr.GetLabels()
	if lbls == nil {
		lbls = map[string]string{}
	}
	//nolint:staticcheck
	previousCluster, newCluster := shared.DeprecatedGetAssignedWorkloadCluster(lbls), shared.DeprecatedGetAssignedWorkloadCluster(ns.Labels)
	if previousCluster == newCluster {
		// Already assigned to the right cluster.
		return nil
	}

	// Update the resource's assignment.
	patchType, patchBytes, err := clusterLabelPatchBytes(previousCluster, newCluster)
	if err != nil {
		klog.Errorf("error creating patch for %s %s|%s: %v", gvr.String(), unstr.GetClusterName(), unstr.GetName(), err)
		return err
	}

	if updated, err := c.dynClient.Cluster(lclusterName).Resource(*gvr).Namespace(ns.Name).
		Patch(ctx, unstr.GetName(), patchType, patchBytes, metav1.PatchOptions{}); err != nil {
		return err
	} else {
		klog.V(2).Infof("Patched cluster assignment for %q %s|%s/%s: %q -> %q. Labels=%v",
			gvr, lclusterName, ns.Name, unstr.GetName(), previousCluster, newCluster, updated.GetLabels())
	}
	return nil
}

func (c *Controller) reconcileGVR(ctx context.Context, gvr schema.GroupVersionResource) error {
	// Update all resources in the namespace with the cluster assignment.
	listers, _ := c.ddsif.Listers()
	lister, found := listers[gvr]
	if !found {
		return fmt.Errorf("informer for %q is not synced; re-enqueueing", gvr)
	}

	// Enqueue workqueue items to reconcile every resource of this type, in
	// all namespaces.
	objs, err := lister.List(labels.Everything())
	if err != nil {
		return err
	}
	for _, obj := range objs {
		c.enqueueResource(gvr, obj)
	}
	return nil
}

// clusterLabelPatchBytes returns a patch expressing an operation
// to add, replace to the given value, or delete the cluster assignment label on
// a resource.
func clusterLabelPatchBytes(old, new string) (types.PatchType, []byte, error) {
	patches := make(map[string]interface{})

	if new == "" && old != "" {
		patches[workloadv1alpha1.InternalClusterResourceStateLabelPrefix+old] = nil
	} else if new != "" && old == "" {
		patches[workloadv1alpha1.InternalClusterResourceStateLabelPrefix+new] = string(workloadv1alpha1.ResourceStateSync)
	} else {
		patches[workloadv1alpha1.InternalClusterResourceStateLabelPrefix+old] = nil
		patches[workloadv1alpha1.InternalClusterResourceStateLabelPrefix+new] = string(workloadv1alpha1.ResourceStateSync)
	}

	bs, err := json.Marshal(map[string]interface{}{"metadata": map[string]interface{}{"labels": patches}})
	if err != nil {
		return "", nil, err
	}
	return types.MergePatchType, bs, nil
}

// schedulingClusterLabelPatchBytes returns a patch expressing an operation
// to add, replace to the given value, or delete the cluster assignment label on a
// namespace.
func schedulingClusterLabelPatchBytes(oldClusterName, newClusterName string) (types.PatchType, []byte, error) {
	patches := make(map[string]interface{})

	if newClusterName == "" && oldClusterName != "" {
		patches[DeprecatedScheduledClusterNamespaceLabel] = nil
		patches[workloadv1alpha1.InternalClusterResourceStateLabelPrefix+oldClusterName] = nil
	} else if newClusterName != "" && oldClusterName == "" {
		patches[DeprecatedScheduledClusterNamespaceLabel] = newClusterName
		patches[workloadv1alpha1.InternalClusterResourceStateLabelPrefix+newClusterName] = string(workloadv1alpha1.ResourceStateSync)
	} else {
		patches[DeprecatedScheduledClusterNamespaceLabel] = newClusterName
		patches[workloadv1alpha1.InternalClusterResourceStateLabelPrefix+oldClusterName] = nil
		patches[workloadv1alpha1.InternalClusterResourceStateLabelPrefix+newClusterName] = string(workloadv1alpha1.ResourceStateSync)
	}

	bs, err := json.Marshal(map[string]interface{}{"metadata": map[string]interface{}{"labels": patches}})
	if err != nil {
		return "", nil, err
	}
	return types.MergePatchType, bs, nil
}

// observeCluster is responsible for watching to see if the Cluster is happy;
// if it's not, any namespace assigned to that cluster with automatic scheduling
// will be unassigned.
//
// After the namespace is unassigned, it will be picked up by
// reconcileNamespace above and assigned to another happy cluster if one can be
// found.
func (c *Controller) observeCluster(ctx context.Context, cluster *workloadv1alpha1.WorkloadCluster) error {
	klog.V(2).Infof("Observing WorkloadCluster %s|%s", logicalcluster.From(cluster), cluster.Name)

	strategy, pendingCordon := enqueueStrategyForCluster(cluster)

	if pendingCordon {
		dur := time.Until(cluster.Spec.EvictAfter.Time)
		c.enqueueClusterAfter(cluster, dur)
	}

	clusterName := logicalcluster.From(cluster)

	switch strategy {
	case enqueueUnscheduled:
		var errs []error
		errs = append(errs, c.enqueueNamespaces(clusterName, labels.NewSelector().Add(unscheduledRequirement).Add(scheduleRequirement)))
		errs = append(errs, c.enqueueNamespaces(clusterName, labels.NewSelector().Add(scheduleEmptyLabelRequirement).Add(scheduleRequirement)))
		return errors.NewAggregate(errs)

	case enqueueScheduled:
		scheduledToCluster, err := labels.NewRequirement(DeprecatedScheduledClusterNamespaceLabel, selection.Equals, []string{cluster.Name})
		if err != nil {
			return err
		}
		return c.enqueueNamespaces(clusterName, labels.NewSelector().Add(*scheduledToCluster))

	case enqueueNothing:
		break

	default:
		return fmt.Errorf("unexpected enqueue strategy: %d", strategy)
	}

	return nil
}

// enqueueNamespaces adds all namespaces matching selector to the queue to allow for scheduling.
func (c *Controller) enqueueNamespaces(clusterName logicalcluster.Name, selector labels.Selector) error {
	// TODO(ncdc): use cluster scoped generated lister when available
	namespaces, err := c.namespaceLister.List(selector)
	if err != nil {
		return err
	}

	for _, namespace := range namespaces {
		if logicalcluster.From(namespace) != clusterName {
			continue
		}

		if namespaceBlocklist.Has(namespace.Name) {
			klog.V(2).Infof("Skipping syncing namespace %q", namespace.Name)
			continue
		}

		c.enqueueNamespace(namespace)
	}

	return nil
}

type clusterEnqueueStrategy int

const (
	enqueueScheduled clusterEnqueueStrategy = iota
	enqueueUnscheduled
	enqueueNothing
)

// enqueueStrategyForCluster determines what namespace enqueueing strategy
// should be used in response to a given cluster state. Also returns a boolean
// indication of whether to enqueue the cluster in the future to respond to a
// impending cordon event.
func enqueueStrategyForCluster(cl *workloadv1alpha1.WorkloadCluster) (strategy clusterEnqueueStrategy, pendingCordon bool) {
	ready := conditions.IsTrue(cl, conditionsapi.ReadyCondition)
	cordoned := cl.Spec.EvictAfter != nil && cl.Spec.EvictAfter.Time.Before(time.Now())
	if !ready || cordoned {
		// An unready or cordoned cluster requires revisiting the scheduling
		// for the namespaces currently scheduled to the cluster to ensure
		// rescheduling is performed.
		return enqueueScheduled, false
	}

	// For ready clusters, a future cordon event requires enqueueing the
	// cluster for processing at the time of the event.
	pendingCordon = cl.Spec.EvictAfter != nil && cl.Spec.EvictAfter.After(time.Now())

	if cl.Spec.Unschedulable {
		// A ready cluster marked unschedulable doesn't allow new
		// assignments and doesn't need rescheduling of existing
		// assignments.
		return enqueueNothing, pendingCordon
	}

	// The cluster is schedulable and not cordoned. Enqueue unscheduled
	// namespaces to allow them to schedule to the cluster.
	return enqueueUnscheduled, pendingCordon
}

// statusPatchBytes returns the bytes required to patch status for the
// provided namespace from its old to new state.
func statusPatchBytes(old, new *corev1.Namespace) ([]byte, error) {
	oldData, err := json.Marshal(corev1.Namespace{
		Status: old.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal existing status for namespace %s|%s: %w", logicalcluster.From(new), new.Name, err)
	}

	newData, err := json.Marshal(corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			UID:             new.UID,
			ResourceVersion: new.ResourceVersion,
		}, // to ensure they appear in the patch as preconditions
		Status: new.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal new status for namespace %s|%s: %w", logicalcluster.From(new), new.Name, err)
	}

	patchBytes, err := jsonpatch.CreateMergePatch(oldData, newData)
	if err != nil {
		return nil, fmt.Errorf("failed to create status patch for namespace %s|%s: %w", logicalcluster.From(new), new.Name, err)
	}
	return patchBytes, nil
}

type getWorkspaceFunc func(name string) (*tenancyv1alpha1.ClusterWorkspace, error)

// isWorkspaceSchedulable indicates whether the contents of the workspace
// identified by the logical cluster name are schedulable.
func isWorkspaceSchedulable(getWorkspace getWorkspaceFunc, logicalClusterName logicalcluster.Name) (bool, error) {
	org, hasParent := logicalClusterName.Parent()
	if !hasParent {
		return false, nil
	}

	workspaceKey := clusters.ToClusterAwareKey(org, logicalClusterName.Base())
	workspace, err := getWorkspace(workspaceKey)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve workspace with key %s", workspaceKey)
	}

	return workspaceSchedulableRequirement.Matches(labels.Set(workspace.Labels)), nil
}
