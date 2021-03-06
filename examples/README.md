# examples

This directory contains example code for installing Sesame and Envoy.

Most subdirectories contain a complete set of Kubernetes YAML that can be applied to a cluster.
This section describes the purpose of each subdirectory.

## [`Sesame`](sesame/README.md)

This is the recommended example installation of Sesame.
It will deploy Sesame into a Deployment, and Envoy into a Daemonset.
The gRPC communication is secured with certificates.
A `LoadBalancer` Service is created to expose Envoy to your cloud provider's load balancer.

## `render`

Single file renderings of other examples suitable for `kubectl apply`ing via a URL.

## `example-workload`

HTTPProxy examples under the `example-workload/httpproxy` directory. See the [README](./example-workload/httpproxy/README.md) for more details on each example.

## `grafana`, `prometheus`

Grafana and Prometheus examples, including the apps themselves, which can show the metrics that Sesame exposes.

If you have your own Grafana and Prometheus deployment already, the supplied [ConfigMap](./grafana/02-grafana-configmap.yaml) contains a sample dashboard with Sesame's metrics.

## `kind`, `root-rbac`

Both of these examples are fragments used in other documentation ([deploy-options](https://projectSesame.io/docs/main/deploy-options))
