# Sesame Configuration Reference

## Configuration File

A configuration file can be passed to the `--config-path` argument of the `Sesame serve` command to specify additional configuration to Sesame.
In most deployments, this file is passed to Sesame via a ConfigMap which is mounted as a volume to the Sesame pod.

The Sesame configuration file is optional.
In its absence, Sesame will operate with reasonable defaults.
Where Sesame settings can also be specified with command-line flags, the command-line value takes precedence over the configuration file.

| Field Name | Type | Default | Description |
|------------|------|---------|-------------|
| accesslog-format | string | `envoy` | This key sets the global [access log format][2] for Envoy. Valid options are `envoy` or `json`. |
| debug | boolean | `false` | Enables debug logging. |
| disablePermitInsecure | boolean | `false` | If this field is true, Sesame will ignore `PermitInsecure` field in HTTPProxy documents. |
| envoy-service-name | string | `envoy` | This sets the service name that will be inspected for address details to be applied to Ingress objects. |
| envoy-service-namespace | string | `projectsesame` | This sets the namespace of the service that will be inspected for address details to be applied to Ingress objects. |
| ingress-status-address | string | None | If present, this specifies the address that will be copied into the Ingress status for each Ingress that Sesame manages. It is exclusive with `envoy-service-name` and `envoy-service-namespace`.|
| incluster | boolean | `false` | This field specifies that Sesame is running in a Kubernetes cluster and should use the in-cluster client access configuration.  |
| json-fields | string array | [fields][5]| This is the list the field names to include in the JSON [access log format][2]. |
| kubeconfig | string | `$HOME/.kube/config` | Path to a Kubernetes [kubeconfig file][3] for when Sesame is executed outside a cluster. |
| leaderelection | leaderelection | | The [leader election configuration](#leader-election-configuration). |
| request-timeout | [duration][4] | `0s` | This field specifies the default request timeout as a Go duration string. Zero means there is no timeout. |
| tls | TLS | | The default [TLS configuration](#tls-configuration). |
{: class="table thead-dark table-bordered"}

### TLS Configuration

The TLS configuration block can be used to configure default values for how
Sesame should provision TLS hosts.

| Field Name | Type| Default  | Description |
|------------|-----|----------|-------------|
| minimum-protocol-version| string | `""` | This field specifies the minimum TLS protocol version that is allowed. Valid options are `1.2` and `1.3`. Any other value defaults to TLS 1.1. |
{: class="table thead-dark table-bordered"}

### Leader Election Configuration

The leader election configuration block configures how a deployment with more than one Sesame pod elects a leader.
The Sesame leader is responsible for updating the status field on Ingress and HTTPProxy documents.
In the vast majority of deployments, only the `configmap-name` and `configmap-namespace` fields should require any configuration.

| Field Name | Type | Default | Description |
|------------|------|---------|-------------|
| configmap-name | string | `leader-elect` | The name of the ConfigMap that Sesame leader election will lease. |
| configmap-namespace | string | `projectsesame` | The namespace of the ConfigMap that Sesame leader election will lease. If the `Sesame_NAMESPACE` environment variable is present, Sesame will populate this field with its value. |
| lease-duration | [duration][4] | `15s` | The duration of the leadership lease. |
| renew-deadline | [duration][4] | `10s` | The length of time that the leader will retry refreshing leadership before giving up. |
| retry-period | [duration][4] | `2s` | The interval at which Sesame will attempt to the acquire leadership lease. |
{: class="table thead-dark table-bordered"}

### Configuration Example

The following is an example ConfigMap with configuration file included:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: sesame
  namespace: projectsesame
data:
  Sesame.yaml: |
    # should Sesame expect to be running inside a k8s cluster
    # incluster: true
    #
    # path to kubeconfig (if not running inside a k8s cluster)
    # kubeconfig: /path/to/.kube/config
    #
    # disable ingressroute permitInsecure field
    # disablePermitInsecure: false
    tls:
      # minimum TLS version that Sesame will negotiate
      # minimumProtocolVersion: "1.1"
    # The following config shows the defaults for the leader election.
    # leaderelection:
      # configmap-name: leader-elect
      # configmap-namespace: projectsesame
```

_Note:_ The default example `Sesame` includes this [file][1] for easy deployment of Sesame.

[1]: {{< param github_url >}}/tree/{{page.version}}/examples/Sesame/01-Sesame-config.yaml
[2]: /guides/structured-logs
[3]: https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/
[4]: https://golang.org/pkg/time/#ParseDuration
[5]: https://godoc.org/github.com/projectsesame/sesame/internal/envoy#DefaultFields
