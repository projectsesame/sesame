# Deployment Options

The [Getting Started][8] guide shows you a simple way to get started with Sesame on your cluster.
This topic explains the details and shows you additional options.
Most of this covers running Sesame using a Kubernetes Service of `Type: LoadBalancer`.
If you don't have a cluster with that capability see the [Running without a Kubernetes LoadBalancer][1] section.

## Installation

### Recommended installation details

The recommended installation is for Sesame to run as a Deployment and Envoy to run as a Daemonset. A secret containing
TLS certificates should be used to secure the gRPC communication between them. A Service of `type: LoadBalancer` should
also be created to forward traffic to the Envoy instances. The [example manifest][2] or [Sesame Operator][12] will
create an installation based on these recommendations.

__Note:__ Sesame Operator is alpha and therefore follows the Sesame [deprecation policy][13].

If you wish to use Host Networking, please see the [appropriate section][3] for the details.

## Testing your installation

### Get your hostname or IP address

To retrieve the IP address or DNS name assigned to your Sesame deployment, run:

```bash
$ kubectl get -n projectsesame service envoy -o wide
```

On AWS, for example, the response looks like:

```
NAME      CLUSTER-IP     EXTERNAL-IP                                                                    PORT(S)        AGE       SELECTOR
Sesame   10.106.53.14   a47761ccbb9ce11e7b27f023b7e83d33-2036788482.ap-southeast-2.elb.amazonaws.com   80:30274/TCP   3h        app=Sesame
```

Depending on your cloud provider, the `EXTERNAL-IP` value is an IP address, or, in the case of Amazon AWS, the DNS name of the ELB created for Sesame. Keep a record of this value.

Note that if you are running an Elastic Load Balancer (ELB) on AWS, you must add more details to your configuration to get the remote address of your incoming connections.
See the [instructions for enabling the PROXY protocol.][4]

#### Minikube

On Minikube, to get the IP address of the Sesame service run:

```bash
$ minikube service -n projectsesame envoy --url
```

The response is always an IP address, for example `http://192.168.99.100:30588`. This is used as Sesame_IP in the rest of the documentation.

#### kind

When creating the cluster on Kind, pass a custom configuration to allow Kind to expose port 80/443 to your local host:

```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    listenAddress: "0.0.0.0"  
  - containerPort: 443
    hostPort: 443
    listenAddress: "0.0.0.0"
```

Then run the create cluster command passing the config file as a parameter.
This file is in the `examples/kind` directory:

```bash
$ kind create cluster --config examples/kind/kind-expose-port.yaml
```

Then, your Sesame_IP (as used below) will just be `localhost:80`.

_Note: We've created a public DNS record (`local.projectsesame.io`) which is configured to resolve to `127.0.0.1``. This allows you to use a real domain name in your kind cluster._

### Test with Ingress

The Sesame repository contains an example deployment of the Kubernetes Up and Running demo application, [kuard][5].
To test your Sesame deployment, deploy `kuard` with the following command:

```bash
$ kubectl apply -f https://projectsesame.io/examples/kuard.yaml
```

Then monitor the progress of the deployment with:

```bash
$ kubectl get po,svc,ing -l app=kuard
```

You should see something like:

```
NAME                       READY     STATUS    RESTARTS   AGE
po/kuard-370091993-ps2gf   1/1       Running   0          4m
po/kuard-370091993-r63cm   1/1       Running   0          4m
po/kuard-370091993-t4dqk   1/1       Running   0          4m

NAME        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
svc/kuard   10.110.67.121   <none>        80/TCP    4m

NAME        HOSTS     ADDRESS     PORTS     AGE
ing/kuard   *         10.0.0.47   80        4m
```

... showing that there are three Pods, one Service, and one Ingress that is bound to all virtual hosts (`*`).

In your browser, navigate your browser to the IP or DNS address of the Sesame Service to interact with the demo application.

### Test with HTTPProxy

To test your Sesame deployment with [HTTPProxy][9], run the following command:

```sh
$ kubectl apply -f https://projectsesame.io/examples/kuard-httpproxy.yaml
```

Then monitor the progress of the deployment with:

```sh
$ kubectl get po,svc,httpproxy -l app=kuard
```

You should see something like:

```sh
NAME                        READY     STATUS    RESTARTS   AGE
pod/kuard-bcc7bf7df-9hj8d   1/1       Running   0          1h
pod/kuard-bcc7bf7df-bkbr5   1/1       Running   0          1h
pod/kuard-bcc7bf7df-vkbtl   1/1       Running   0          1h

NAME            TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)   AGE
service/kuard   ClusterIP   10.102.239.168   <none>        80/TCP    1h

NAME                                    FQDN                TLS SECRET                  FIRST ROUTE  STATUS  STATUS DESCRIPT
httpproxy.projectsesame.io/kuard      kuard.local         <SECRET NAME IF TLS USED>                valid   valid HTTPProxy
```

... showing that there are three Pods, one Service, and one HTTPProxy .

In your terminal, use curl with the IP or DNS address of the Sesame Service to send a request to the demo application:

```sh
$ curl -H 'Host: kuard.local' ${Sesame_IP}
```

## Running without a Kubernetes LoadBalancer

If you can't or don't want to use a Service of `type: LoadBalancer` there are other ways to run Sesame.

### NodePort Service

If your cluster doesn't have the capability to configure a Kubernetes LoadBalancer,
or if you want to configure the load balancer outside Kubernetes,
you can change the Envoy Service in the [`02-service-envoy.yaml`][7] file and set `type` to `NodePort`.

This will have every node in your cluster listen on the resultant port and forward traffic to Sesame.
That port can be discovered by taking the second number listed in the `PORT` column when listing the service, for example `30274` in `80:30274/TCP`.

Now you can point your browser at the specified port on any node in your cluster to communicate with Sesame.

### Host Networking

You can run Sesame without a Kubernetes Service at all.
This is done by having the Envoy pod run with host networking.
Sesame's examples utilize this model in the `/examples` directory.
To configure, set: `hostNetwork: true` and `dnsPolicy: ClusterFirstWithHostNet` on your Envoy pod definition.
Next, pass `--envoy-service-http-port=80 --envoy-service-https-port=443` to the Sesame `serve` command which instructs Envoy to listen directly on port 80/443 on each host that it is running.
This is best paired with a DaemonSet (perhaps paired with Node affinity) to ensure that a single instance of Sesame runs on each Node.
See the [AWS NLB tutorial][10] as an example.

### Upgrading Sesame/Envoy

At times it's needed to upgrade Sesame, the version of Envoy, or both.
The included `shutdown-manager` can assist with watching Envoy for open connections while draining and give signal back to Kubernetes as to when it's fine to delete Envoy pods during this process.

See the [redeploy envoy][11] docs for more information.

## Running Sesame in tandem with another ingress controller

If you're running multiple ingress controllers, or running on a cloudprovider that natively handles ingress,
you can specify the annotation `kubernetes.io/ingress.class: "Sesame"` on all ingresses that you would like Sesame to claim.
You can customize the class name with the `--ingress-class-name` flag at runtime.
If the `kubernetes.io/ingress.class` annotation is present with a value other than `"Sesame"`, Sesame will ignore that ingress.

## Uninstall Sesame

To remove Sesame from your cluster, delete the namespace:

```bash
$ kubectl delete ns projectsesame
```
**Note**: The namespace may differ from above if [Sesame Operator][12] was used to
deploy Sesame.

## Uninstall Sesame Operator

To remove Sesame Operator from your cluster, delete the operator's namespace:

```bash
$ kubectl delete ns sesame-operator
```

[1]: #running-without-a-kubernetes-loadbalancer
[2]: {{< param github_url>}}/tree/{{< param version >}}/examples/Sesame
[3]: #host-networking
[4]: /guides/proxy-proto.md
[5]: https://github.com/kubernetes-up-and-running/kuard
[7]: {{< param github_url>}}/tree/{{< param version >}}/examples/Sesame/02-service-envoy.yaml
[8]: /getting-started.md
[9]: config/fundamentals.md
[10]: /guides/deploy-aws-nlb.md
[11]: redeploy-envoy.md
[12]: https://github.com/projectsesame/sesame-operator
[13]: https://projectsesame.io/resources/deprecation-policy/
