/*
Copyright 2022 The KCP Authors.

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

package options

import (
	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	allowedFlags = sets.NewString(
		// auditing flags
		"audit-log-batch-buffer-size",           // The size of the buffer to store events before batching and writing. Only used in batch mode.
		"audit-log-batch-max-size",              // The maximum size of a batch. Only used in batch mode.
		"audit-log-batch-max-wait",              // The amount of time to wait before force writing the batch that hadn't reached the max size. Only used in batch mode.
		"audit-log-batch-throttle-burst",        // Maximum number of requests sent at the same moment if ThrottleQPS was not utilized before. Only used in batch mode.
		"audit-log-batch-throttle-enable",       // Whether batching throttling is enabled. Only used in batch mode.
		"audit-log-batch-throttle-qps",          // Maximum average number of batches per second. Only used in batch mode.
		"audit-log-compress",                    // If set, the rotated log files will be compressed using gzip.
		"audit-log-format",                      // Format of saved audits. "legacy" indicates 1-line text format for each event. "json" indicates structured json format. Known formats are legacy,json.
		"audit-log-maxage",                      // The maximum number of days to retain old audit log files based on the timestamp encoded in their filename.
		"audit-log-maxbackup",                   // The maximum number of old audit log files to retain.
		"audit-log-maxsize",                     // The maximum size in megabytes of the audit log file before it gets rotated.
		"audit-log-mode",                        // Strategy for sending audit events. Blocking indicates sending events should block server responses. Batch causes the backend to buffer and write events asynchronously. Known modes are batch,blocking,blocking-strict.
		"audit-log-path",                        // If set, all requests coming to the apiserver will be logged to this file.  '-' means standard out.
		"audit-log-truncate-enabled",            // Whether event and batch truncating is enabled.
		"audit-log-truncate-max-batch-size",     // Maximum size of the batch sent to the underlying backend. Actual serialized size can be several hundreds of bytes greater. If a batch exceeds this limit, it is split into several batches of smaller size.
		"audit-log-truncate-max-event-size",     // Maximum size of the audit event sent to the underlying backend. If the size of an event is greater than this number, first request and response are removed, and if this doesn't reduce the size enough, event is discarded.
		"audit-log-version",                     // API group and version used for serializing audit events written to log.
		"audit-policy-file",                     // Path to the file that defines the audit policy configuration.
		"audit-webhook-batch-buffer-size",       // The size of the buffer to store events before batching and writing. Only used in batch mode.
		"audit-webhook-batch-initial-backoff",   // The amount of time to wait before retrying the first failed request.
		"audit-webhook-batch-max-size",          // The maximum size of a batch. Only used in batch mode.
		"audit-webhook-batch-max-wait",          // The amount of time to wait before force writing the batch that hadn't reached the max size. Only used in batch mode.
		"audit-webhook-batch-throttle-burst",    // Maximum number of requests sent at the same moment if ThrottleQPS was not utilized before. Only used in batch mode.
		"audit-webhook-batch-throttle-enable",   // Whether batching throttling is enabled. Only used in batch mode.
		"audit-webhook-batch-throttle-qps",      // Maximum average number of batches per second. Only used in batch mode.
		"audit-webhook-config-file",             // Path to a kubeconfig formatted file that defines the audit webhook configuration.
		"audit-webhook-initial-backoff",         // The amount of time to wait before retrying the first failed request.
		"audit-webhook-mode",                    // Strategy for sending audit events. Blocking indicates sending events should block server responses. Batch causes the backend to buffer and write events asynchronously. Known modes are batch,blocking,blocking-strict.
		"audit-webhook-truncate-enabled",        // Whether event and batch truncating is enabled.
		"audit-webhook-truncate-max-batch-size", // Maximum size of the batch sent to the underlying backend. Actual serialized size can be several hundreds of bytes greater. If a batch exceeds this limit, it is split into several batches of smaller size.
		"audit-webhook-truncate-max-event-size", // Maximum size of the audit event sent to the underlying backend. If the size of an event is greater than this number, first request and response are removed, and if this doesn't reduce the size enough, event is discarded.
		"audit-webhook-version",                 // API group and version used for serializing audit events written to webhook.

		// authentication flags
		"anonymous-auth",                     // Enables anonymous requests to the secure port of the API server. Requests that are not rejected by another authentication method are treated as anonymous requests. Anonymous requests have a username of system:anonymous, and a group name of system:unauthenticated.
		"api-audiences",                      // Identifiers of the API. The service account token authenticator will validate that tokens used against the API are bound to at least one of these audiences. If the --service-account-issuer flag is configured and this flag is not, this field defaults to a single element list containing the issuer URL.
		"client-ca-file",                     // If set, any request presenting a client certificate signed by one of the authorities in the client-ca-file is authenticated with an identity corresponding to the CommonName of the client certificate.
		"enable-bootstrap-token-auth",        // Enable to allow secrets of type 'bootstrap.kubernetes.io/token' in the 'kube-system' namespace to be used for TLS bootstrapping authentication.
		"oidc-ca-file",                       // If set, the OpenID server's certificate will be verified by one of the authorities in the oidc-ca-file, otherwise the host's root CA set will be used.
		"oidc-client-id",                     // The client ID for the OpenID Connect client, must be set if oidc-issuer-url is set.
		"oidc-groups-claim",                  // If provided, the name of a custom OpenID Connect claim for specifying user groups. The claim value is expected to be a string or array of strings. This flag is experimental, please see the authentication documentation for further details.
		"oidc-groups-prefix",                 // If provided, all groups will be prefixed with this value to prevent conflicts with other authentication strategies.
		"oidc-issuer-url",                    // The URL of the OpenID issuer, only HTTPS scheme will be accepted. If set, it will be used to verify the OIDC JSON Web Token (JWT).
		"oidc-required-claim",                // A key=value pair that describes a required claim in the ID Token. If set, the claim is verified to be present in the ID Token with a matching value. Repeat this flag to specify multiple claims.
		"oidc-signing-algs",                  // Comma-separated list of allowed JOSE asymmetric signing algorithms. JWTs with a 'alg' header value not in this list will be rejected. Values are defined by RFC 7518 https://tools.ietf.org/html/rfc7518#section-3.1.
		"oidc-username-claim",                // The OpenID claim to use as the user name. Note that claims other than the default ('sub') is not guaranteed to be unique and immutable. This flag is experimental, please see the authentication documentation for further details.
		"oidc-username-prefix",               // If provided, all usernames will be prefixed with this value. If not provided, username claims other than 'email' are prefixed by the issuer URL to avoid clashes. To skip any prefixing, provide the value '-'.
		"requestheader-allowed-names",        // List of client certificate common names to allow to provide usernames in headers specified by --requestheader-username-headers. If empty, any client certificate validated by the authorities in --requestheader-client-ca-file is allowed.
		"requestheader-client-ca-file",       // Root certificate bundle to use to verify client certificates on incoming requests before trusting usernames in headers specified by --requestheader-username-headers. WARNING: generally do not depend on authorization being already done for incoming requests.
		"requestheader-extra-headers-prefix", // List of request header prefixes to inspect. X-Remote-Extra- is suggested.
		"requestheader-group-headers",        // List of request headers to inspect for groups. X-Remote-Group is suggested.
		"requestheader-username-headers",     // List of request headers to inspect for usernames. X-Remote-User is common.
		"token-auth-file",                    // If set, the file that will be used to secure the secure port of the API server via token authentication.

		// KCP Authorization flags
		"authorization-always-allow-paths", // A list of HTTP paths to skip during authorization, i.e. these are authorized without contacting the 'core' kubernetes server.

		// logs flags
		"logging-format", // Sets the log format. Permitted formats: "text".

		// traces flags
		"tracing-config-file", // File with apiserver tracing configuration.

		// KCP flags
		"discovery-poll-interval", // Polling interval for dynamic discovery informers.
		"enable-sharding",         // Enable delegating to peer kcp shards.
		"kubeconfig-path",         // Path to which the administrative kubeconfig should be written at startup.
		"profiler-address",        // [Address]:port to bind the profiler to
		"root-directory",          // Root directory.
		"shard-kubeconfig-file",   // Kubeconfig holding admin(!) credentials to peer kcp shards.

		// secure serving flags
		"bind-address",                     // The IP address on which to listen for the --secure-port port. The associated interface(s) must be reachable by the rest of the cluster, and by CLI/web clients. If blank or an unspecified address (0.0.0.0 or ::), all interfaces will be used.
		"cert-dir",                         // The directory where the TLS certs are located. If --tls-cert-file and --tls-private-key-file are provided, this flag will be ignored.
		"http2-max-streams-per-connection", // The limit that the server gives to clients for the maximum number of streams in an HTTP/2 connection. Zero means to use golang's default.
		"permit-address-sharing",           // If true, SO_REUSEADDR will be used when binding the port. This allows binding to wildcard IPs like 0.0.0.0 and specific IPs in parallel, and it avoids waiting for the kernel to release sockets in TIME_WAIT state. [default=false]
		"permit-port-sharing",              // If true, SO_REUSEPORT will be used when binding the port, which allows more than one instance to bind on the same address and port. [default=false]
		"secure-port",                      // The port on which to serve HTTPS with authentication and authorization. It cannot be switched off with 0.
		"tls-cert-file",                    // File containing the default x509 Certificate for HTTPS. (CA cert, if any, concatenated after server cert). If HTTPS serving is enabled, and --tls-cert-file and --tls-private-key-file are not provided, a self-signed certificate and key are generated for the public address and saved to the directory specified by --cert-dir.
		"tls-cipher-suites",                // Comma-separated list of cipher suites for the server. If omitted, the default Go cipher suites will be used.
		"tls-min-version",                  // Minimum TLS version supported. Possible values: VersionTLS10, VersionTLS11, VersionTLS12, VersionTLS13
		"tls-private-key-file",             // File containing the default x509 private key matching --tls-cert-file.
		"tls-sni-cert-key",                 // A pair of x509 certificate and private key file paths, optionally suffixed with a list of domain patterns which are fully qualified domain names, possibly with prefixed wildcard segments. The domain patterns also allow IP addresses, but IPs should only be used if the apiserver has visibility to the IP address requested by a client. If no domain patterns are provided, the names of the certificate are extracted. Non-wildcard matches trump over wildcard matches, explicit domain patterns trump over extracted names. For multiple key/certificate pairs, use the --tls-sni-cert-key multiple times. Examples: "example.crt,example.key" or "foo.crt,foo.key:*.foo.com,foo.com".

		// Embedded etcd flags
		"embedded-etcd-client-port",    // Port for embedded etcd client
		"embedded-etcd-directory",      // Directory for embedded etcd
		"embedded-etcd-peer-port",      // Port for embedded etcd peer
		"embedded-etcd-wal-size-bytes", // Size of embedded etcd WAL

		// KCP Controllers flags
		"auto-publish-apis",                      // If true, the APIs imported from physical clusters will be published automatically as CRDs
		"cluster-controller-threads",             // Number of threads to use for the cluster controller.
		"pull-mode",                              // Deploy the syncer in registered physical clusters in POD, and have it sync resources from KCP
		"push-mode",                              // If true, run syncer for each cluster from inside cluster controller
		"resources-to-sync",                      // Provides the list of resources that should be synced from KCP logical cluster to underlying physical clusters
		"run-controllers",                        // Run the controllers in-process
		"syncer-image",                           // Syncer image to install on clusters
		"unsupported-run-individual-controllers", // Run individual controllers in-process. The controller names can change at any time.

		// generic flags
		"cors-allowed-origins",                 // List of allowed origins for CORS, comma separated.  An allowed origin can be a regular expression to support subdomain matching. If this list is empty CORS will not be enabled.
		"goaway-chance",                        // To prevent HTTP/2 clients from getting stuck on a single apiserver, randomly close a connection (GOAWAY). The client's other in-flight requests won't be affected, and the client will reconnect, likely landing on a different apiserver after going through the load balancer again. This argument sets the fraction of requests that will be sent a GOAWAY. Clusters with single apiservers, or which don't use a load balancer, should NOT enable this. Min is 0 (off), Max is .02 (1/50 requests); .001 (1/1000) is a recommended starting point.
		"livez-grace-period",                   // This option represents the maximum amount of time it should take for apiserver to complete its startup sequence and become live. From apiserver's start time to when this amount of time has elapsed, /livez will assume that unfinished post-start hooks will complete successfully and therefore return true.
		"shutdown-delay-duration",              // Time to delay the termination. During that time the server keeps serving requests normally. The endpoints /healthz and /livez will return success, but /readyz immediately returns failure. Graceful termination starts after this delay has elapsed. This can be used to allow load balancer to stop sending traffic to this server.
		"strict-transport-security-directives", // List of directives for HSTS, comma separated. If this list is empty, then HSTS directives will not be added. Example: 'max-age=31536000,includeSubDomains,preload'

		// etcd flags
		"etcd-cafile",                   // SSL Certificate Authority file used to secure etcd communication.
		"etcd-certfile",                 // SSL certification file used to secure etcd communication.
		"etcd-compaction-interval",      // The interval of compaction requests. If 0, the compaction request from apiserver is disabled.
		"etcd-count-metric-poll-period", // Frequency of polling etcd for number of resources per type. 0 disables the metric collection.
		"etcd-db-metric-poll-interval",  // The interval of requests to poll etcd and update metric. 0 disables the metric collection
		"etcd-healthcheck-timeout",      // The timeout to use when checking etcd health.
		"etcd-keyfile",                  // SSL key file used to secure etcd communication.
		"etcd-prefix",                   // The prefix to prepend to all resource paths in etcd.
		"etcd-servers",                  // List of etcd servers to connect with (scheme://ip:port), comma separated.
		"etcd-servers-overrides",        // Per-resource etcd servers overrides, comma separated. The individual override format: group/resource#servers, where servers are URLs, semicolon separated. Note that this applies only to resources compiled into this server binary.
		"lease-reuse-duration-seconds",  // The time in seconds that each lease is reused. A lower value could avoid large number of objects reusing the same lease. Notice that a too small value may cause performance problems at storage layer.
		"storage-backend",               // The storage backend for persistence. Options: 'etcd3' (default).
		"storage-media-type",            // The media type to use to store objects in storage. Some resources or storage backends may only support a specific media type and will ignore this setting.
		"watch-cache",                   // Enable watch caching in the apiserver
		"watch-cache-sizes",             // Watch cache size settings for some resources (pods, nodes, etc.), comma separated. The individual setting format: resource[.group]#size, where resource is lowercase plural (no version), group is omitted for resources of apiVersion v1 (the legacy core API) and included for others, and size is a number. It takes effect when watch-cache is enabled. Some resources (replicationcontrollers, endpoints, nodes, pods, services, apiservices.apiregistration.k8s.io) have system defaults set by heuristics, others default to default-watch-cache-size

		// features flags
		"contention-profiling", // Enable lock contention profiling, if profiling is enabled
		"profiling",            // Enable profiling via web interface host:port/debug/pprof/

		// metrics flags
		"allow-metric-labels",             // The map from metric-label to value allow-list of this label. The key's format is <MetricName>,<LabelName>. The value's format is <allowed_value>,<allowed_value>...e.g. metric1,label1='v1,v2,v3', metric1,label2='v1,v2,v3' metric2,label1='v1,v2,v3'.
		"disabled-metrics",                // This flag provides an escape hatch for misbehaving metrics. You must provide the fully qualified metric name in order to disable it. Disclaimer: disabling metrics is higher in precedence than showing hidden metrics.
		"show-hidden-metrics-for-version", // The previous version for which you want to show hidden metrics. Only the previous minor version is meaningful, other values will not be allowed. The format is <major>.<minor>, e.g.: '1.16'. The purpose of this format is make sure you have the opportunity to notice if the next release hides additional metrics, rather than being surprised when they are permanently removed in the release after that.

		// misc flags
		"enable-logs-handler",                   // If true, install a /logs handler for the apiserver logs.
		"event-ttl",                             // Amount of time to retain events.
		"identity-lease-duration-seconds",       // The duration of kube-apiserver lease in seconds, must be a positive number. (In use when the APIServerIdentity feature gate is enabled.)
		"identity-lease-renew-interval-seconds", // The interval of kube-apiserver renewing its lease in seconds, must be a positive number. (In use when the APIServerIdentity feature gate is enabled.)
		"max-connection-bytes-per-sec",          // If non-zero, throttle each user connection to this number of bytes/sec. Currently only applies to long-running requests.
		"proxy-client-cert-file",                // Client certificate used to prove the identity of the aggregator or kube-apiserver when it must call out during a request. This includes proxying requests to a user api-server and calling out to webhook admission plugins. It is expected that this cert includes a signature from the CA in the --requestheader-client-ca-file flag. That CA is published in the 'extension-apiserver-authentication' configmap in the kube-system namespace. Components receiving calls from kube-aggregator should use that CA to perform their half of the mutual TLS verification.
		"proxy-client-key-file",                 // Private key for the client certificate used to prove the identity of the aggregator or kube-apiserver when it must call out during a request. This includes proxying requests to a user api-server and calling out to webhook admission plugins.
	)

	disallowedFlags = sets.NewString(
		// logs flags
		"experimental-logging-sanitization", // [Experimental] When enabled prevents logging of fields tagged as sensitive (passwords, keys, tokens).

		// generic flags
		"advertise-address",              // The IP address on which to advertise the apiserver to members of the cluster. This address must be reachable by the rest of the cluster. If blank, the --bind-address will be used. If --bind-address is unspecified, the host's default interface will be used.
		"enable-priority-and-fairness",   // If true and the APIPriorityAndFairness feature gate is enabled, replace the max-in-flight handler with an enhanced one that queues and dispatches with priority and fairness
		"external-hostname",              // The hostname to use when generating externalized URLs for this master (e.g. Swagger API Docs or OpenID Discovery).
		"feature-gates",                  // A set of key=value pairs that describe feature gates for alpha/experimental features. Options are:
		"master-service-namespace",       // DEPRECATED: the namespace from which the Kubernetes master services should be injected into pods.
		"max-mutating-requests-inflight", // This and --max-requests-inflight are summed to determine the server's total concurrency limit (which must be positive) if --enable-priority-and-fairness is true. Otherwise, this flag limits the maximum number of mutating requests in flight, or a zero value disables the limit completely.
		"max-requests-inflight",          // This and --max-mutating-requests-inflight are summed to determine the server's total concurrency limit (which must be positive) if --enable-priority-and-fairness is true. Otherwise, this flag limits the maximum number of non-mutating requests in flight, or a zero value disables the limit completely.
		"min-request-timeout",            // An optional field indicating the minimum number of seconds a handler must keep a request open before timing it out. Currently only honored by the watch request handler, which picks a randomized value above this number as the connection timeout, to spread out load.
		"request-timeout",                // An optional field indicating the duration a handler must keep a request open before timing it out. This is the default request timeout for requests but may be overridden by flags such as --min-request-timeout for specific types of requests.
		"target-ram-mb",                  // DEPRECATED: Memory limit for apiserver in MB (used to configure sizes of caches, etc.)

		// etcd flags
		"default-watch-cache-size",                // Default watch cache size. If zero, watch cache will be disabled for resources that do not have a default watch size set.
		"delete-collection-workers",               // Number of workers spawned for DeleteCollection call. These are used to speed up namespace cleanup.
		"deserialization-cache-size",              // Number of deserialized json objects to cache in memory.
		"enable-garbage-collector",                // Enables the generic garbage collector. MUST be synced with the corresponding flag of the kube-controller-manager.
		"encryption-provider-config",              // The file containing configuration for encryption providers to be used for storing secrets in etcd
		"experimental-encryption-provider-config", // The file containing configuration for encryption providers to be used for storing secrets in etcd

		// features flags
		"enable-swagger-ui", // Enables swagger ui on the apiserver at /swagger-ui

		// admission flags
		"admission-control-config-file", // File with admission control configuration.
		"disable-admission-plugins",     // admission plugins that should be disabled although they are in the default enabled plugins list (NamespaceLifecycle). Comma-delimited list of admission plugins: MutatingAdmissionWebhook, NamespaceLifecycle, ValidatingAdmissionWebhook. The order of plugins in this flag does not matter.
		"enable-admission-plugins",      // admission plugins that should be enabled in addition to default enabled ones (NamespaceLifecycle). Comma-delimited list of admission plugins: MutatingAdmissionWebhook, NamespaceLifecycle, ValidatingAdmissionWebhook. The order of plugins in this flag does not matter.

		// egress selector flags
		"egress-selector-config-file", // File with apiserver egress selector configuration.
	)
)