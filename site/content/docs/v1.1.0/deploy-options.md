# Deployment Options

The [Getting Started][8] guide shows you a simple way to get started with Sesame on your cluster.
This topic explains the details and shows you additional options.
Most of this covers running Sesame using a Kubernetes Service of `Type: LoadBalancer`.
If you don't have a cluster with that capability see the [Running without a Kubernetes LoadBalancer][1] section.

## Installation

### Recommended installation details

The recommended installation of Sesame is Sesame running in a Deployment and Envoy in a Daemonset with TLS securing the gRPC communication between them.
The [`Sesame` example][2] will install this for you.
A Service of `type: LoadBalancer` is also set up to forward traffic to the Envoy instances.

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

### Test with IngressRoute

To test your Sesame deployment with [IngressRoutes][6], run the following command:

```sh
$ kubectl apply -f https://projectsesame.io/examples/kuard-ingressroute.yaml
```

Then monitor the progress of the deployment with:

```sh
$ kubectl get po,svc,ingressroute -l app=kuard
```

You should see something like:

```sh
NAME                        READY     STATUS    RESTARTS   AGE
pod/kuard-bcc7bf7df-9hj8d   1/1       Running   0          1h
pod/kuard-bcc7bf7df-bkbr5   1/1       Running   0          1h
pod/kuard-bcc7bf7df-vkbtl   1/1       Running   0          1h

NAME            TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)   AGE
service/kuard   ClusterIP   10.102.239.168   <none>        80/TCP    1h

NAME                                    CREATED AT
ingressroute.sesame.heptio.com/kuard   1h
```

... showing that there are three Pods, one Service, and one IngressRoute.

In your terminal, use curl with the IP or DNS address of the Sesame Service to send a request to the demo application:

```sh
$ curl -H 'Host: kuard.local' ${Sesame_IP}
```
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
This is done by having the Sesame pod run with host networking.
Do this with `hostNetwork: true` and `dnsPolicy: ClusterFirstWithHostNet` on your pod definition.
Envoy will listen directly on port 8080 on each host that it is running.
This is best paired with a DaemonSet (perhaps paired with Node affinity) to ensure that a single instance of Sesame runs on each Node.
See the [AWS NLB tutorial][10] as an example.

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

[1]: #running-without-a-kubernetes-loadbalancer
[2]: {{< param github_url >}}/tree/{{page.version}}/examples/Sesame/README.md
[3]: #host-networking
[4]: {% link _guides/proxy-proto.md %}
[5]: https://github.com/kubernetes-up-and-running/kuard
[6]: /docs/{{page.version}}/ingressroute
[7]: {{< param github_url >}}/tree/{{page.version}}/examples/Sesame/02-service-envoy.yaml
[8]: {% link getting-started.md %}
[9]: httpproxy.md
[10]: {% link _guides/deploy-aws-nlb.md %}
