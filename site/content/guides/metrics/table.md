| Name | Type | Labels | Description |
| ---- | ---- | ------ | ----------- |
| Sesame_build_info | [GAUGE](https://prometheus.io/docs/concepts/metric_types/#gauge) | branch, revision, version | Build information for Sesame. Labels include the branch and git SHA that Sesame was built from, and the Sesame version. |
| Sesame_cachehandler_onupdate_duration_seconds | [SUMMARY](https://prometheus.io/docs/concepts/metric_types/#summary) |  | Histogram for the runtime of xDS cache regeneration. |
| Sesame_dagrebuild_timestamp | [GAUGE](https://prometheus.io/docs/concepts/metric_types/#gauge) |  | Timestamp of the last DAG rebuild. |
| Sesame_dagrebuild_total | [COUNTER](https://prometheus.io/docs/concepts/metric_types/#counter) |  | Total number of times DAG has been rebuilt since startup |
| Sesame_eventhandler_operation_total | [COUNTER](https://prometheus.io/docs/concepts/metric_types/#counter) | kind, op | Total number of Kubernetes object changes Sesame has received by operation and object kind. |
| Sesame_httpproxy | [GAUGE](https://prometheus.io/docs/concepts/metric_types/#gauge) | namespace | Total number of HTTPProxies that exist regardless of status. |
| Sesame_httpproxy_invalid | [GAUGE](https://prometheus.io/docs/concepts/metric_types/#gauge) | namespace, vhost | Total number of invalid HTTPProxies. |
| Sesame_httpproxy_orphaned | [GAUGE](https://prometheus.io/docs/concepts/metric_types/#gauge) | namespace | Total number of orphaned HTTPProxies which have no root delegating to them. |
| Sesame_httpproxy_root | [GAUGE](https://prometheus.io/docs/concepts/metric_types/#gauge) | namespace | Total number of root HTTPProxies. Note there will only be a single root HTTPProxy per vhost. |
| Sesame_httpproxy_valid | [GAUGE](https://prometheus.io/docs/concepts/metric_types/#gauge) | namespace, vhost | Total number of valid HTTPProxies. |
