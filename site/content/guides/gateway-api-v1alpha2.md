---
title: Using Gateway API v1alpha2 with Sesame
layout: page
---

## Introduction

[Gateway API][1] is an open source project managed by the Kubernetes SIG-NETWORK community. The project's goal is to
evolve service networking APIs within the Kubernetes ecosystem. Gateway API consists of multiple resources that provide
user interfaces to expose Kubernetes applications- Services, Ingress, and more.

This guide covers using version **v1alpha2** of the Gateway API, with Sesame `v1.20.0-beta.1`.
Please note that this is pre-release software, and we don't recommend installing it in a production environment.

### Background

Gateway API targets three personas:

- __Platform Provider__: The Platform Provider is responsible for the overall environment that the cluster runs in, i.e.
  the cloud provider. The Platform Provider will interact with GatewayClass and Sesame resources.
- __Platform Operator__: The Platform Operator is responsible for overall cluster administration. They manage policies,
  network access, application permissions and will interact with the Gateway resource.
- __Service Operator__: The Service Operator is responsible for defining application configuration and service
  composition. They will interact with xRoute resources and other typical Kubernetes resources.

Gateway API contains three primary resources:

- __GatewayClass__: Defines a set of gateways with a common configuration and behavior.
- __Gateway__: Requests a point where traffic can be translated to a Service within the cluster.
- __xRoutes__: Describes how traffic coming via the Gateway maps to the Services.

Resources are meant to align with personas. For example, a platform operator will create a Gateway, so a developer can
expose an HTTP application using an HTTPRoute resource.

### Prerequisites
The following prerequisites must be met before using Gateway API with Sesame:

- A working [Kubernetes][2] cluster. Refer to the [compatibility matrix][3] for cluster version requirements.
- The [kubectl][4] command-line tool, installed and configured to access your cluster.

## Deploying Sesame with Gateway API

### Option #1: Gateway API with Sesame

Refer to the [Sesame][6] design for additional background on the Gateway API implementation.

Deploy Sesame:
```shell
$ kubectl apply -f https://raw.githubusercontent.com/projectsesame/Sesame/v1.20.0-beta.1/examples/render/Sesame-gateway.yaml
```
This command creates:

- Namespace `projectsesame` to run Sesame.
- Sesame CRDs
- Gateway API CRDs
- Sesame RBAC resources
- Sesame Deployment / Service
- Envoy Daemonset / Service
- Sesame ConfigMap enabling Gateway API support
- Gateway API GatewayClass and Gateway

See the last section ([Testing the Gateway API](#testing-the-gateway-api)) on how to test it all out!

### Option #2: Using Gateway API with Sesame Operator

Refer to the [Sesame][6] and [operator][7] designs for additional background on the gateway API implementation.

Run the operator:
```shell
$ kubectl apply -f https://raw.githubusercontent.com/projectsesame/sesame-operator/v1.20.0-beta.1/examples/operator/operator.yaml
```
This command creates:

- Namespace `Sesame-operator` to run the operator.
- Operator and Sesame CRDs.
- Operator RBAC resources for the operator.
- A Deployment to manage the operator.
- A Service to front-end the operator’s metrics endpoint.

Create the Gateway API resources:

Option 1: Using a LoadBalancer Service:
```shell
$ kubectl apply -f https://raw.githubusercontent.com/projectsesame/sesame-operator/v1.20.0-beta.1/examples/gateway/gateway.yaml
```

Option 2: Using a NodePort Service:
```shell
$ kubectl apply -f https://raw.githubusercontent.com/projectsesame/sesame-operator/v1.20.0-beta.1/examples/gateway/gateway-nodeport.yaml
```

Either of the above options create:

- Namespace `projectsesame` to run the Gateway and child resources, i.e. Envoy DaemonSet.
- A Sesame custom resource named `Sesame-gateway-sample` in the operator's namespace. This resource exposes infrastructure-specific configuration and is referenced by the GatewayClass.
- A GatewayClass named `sample-gatewayclass` that abstracts the infrastructure-specific configuration from Gateways.
- A Gateway named `Sesame` in namespace `projectsesame`. This gateway will serve the test application through routing rules deployed in the next step.

See the next section ([Testing the Gateway API](#testing-the-gateway-api)) on how to test it all out!

## Testing the Gateway API

Run the test application:
```shell
$ kubectl apply -f https://raw.githubusercontent.com/projectsesame/Sesame/v1.20.0-beta.1/examples/gateway/kuard/kuard.yaml
```
This command creates:

- A Deployment named `kuard` in namespace `projectsesame` to run kuard as the test application.
- A Service named `kuard` in namespace `projectsesame` to expose the kuard application on TCP port 80.
- An HTTPRoute named `kuard` in namespace `projectsesame` to route requests for `local.projectsesame.io` to the kuard
  service.

Verify the kuard resources are available:
```shell
$ kubectl get po,svc,httproute -n projectsesame -l app=kuard
NAME                         READY   STATUS    RESTARTS   AGE
pod/kuard-798585497b-78x6x   1/1     Running   0          21s
pod/kuard-798585497b-7gktg   1/1     Running   0          21s
pod/kuard-798585497b-zw42m   1/1     Running   0          21s

NAME            TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)   AGE
service/kuard   ClusterIP   172.30.168.168   <none>        80/TCP    21s

NAME                                        HOSTNAMES
httproute.gateway.networking.k8s.io/kuard   ["local.projectsesame.io"]
```

Test access to the kuard application:

Get the Envoy Service IP address:
```shell
export GATEWAY=$(kubectl -n projectsesame get svc/envoy -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
```
__Note:__ Replace `hostname` with `ip` in the above command if your cloud provide uses an IP address for a
load-balancer.

Use `curl` to test access to the application:
```shell
$ curl -H "Host: local.projectsesame.io" -s -o /dev/null -w "%{http_code}" "http://$GATEWAY/"
```
A 200 HTTP status code should be returned.

[1]: https://gateway-api.sigs.k8s.io/
[2]: https://kubernetes.io/
[3]: https://projectsesame.io/resources/compatibility-matrix/
[4]: https://kubernetes.io/docs/tasks/tools/install-kubectl/
[5]: https://github.com/projectsesame/sesame-operator
[6]: https://github.com/projectsesame/sesame/blob/main/design/gateway-apis-implementation.md
[7]: https://github.com/projectsesame/sesame-operator/blob/main/design/gateway-api.md
