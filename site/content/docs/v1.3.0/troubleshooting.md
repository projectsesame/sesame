This document contains suggestions for debugging issues with your Sesame installation.

## Envoy container not listening on port 8080 or 8443

Sesame does not configure Envoy to listen on a port unless there is traffic to be served.
For example, if you have not configured any TLS ingress objects then Sesame does not command Envoy to open port 8443 (443 in the service object).
Because the HTTP and HTTPS listeners both use the same code, if you have no ingress objects deployed in your cluster, or if no ingress objects are permitted to talk on HTTP, then Envoy does not listen on port 8080 (80 in the service object).

To test whether Sesame is correctly deployed you can deploy the kuard example service:

```sh
$ kubectl apply -f https://projectsesame.io/examples/kuard.yaml
```

## Access the Envoy admin interface remotely

Getting access to the Envoy admin interface can be useful for diagnosing issues with routing or cluster health.

The Envoy admin interface is bound by default to `http://127.0.0.1:9001`.
To access it from your workstation use `kubectl port-forward` like so,

```sh
# Get one of the pods that matches the Envoy daemonset
ENVOY_POD=$(kubectl -n projectsesame get pod -l app=envoy -o name | head -1)
# Do the port forward to that pod
kubectl -n projectsesame port-forward $ENVOY_POD 9001
```

Then navigate to `http://127.0.0.1:9001/` to access the admin interface for the Envoy container running on that pod.

## Accessing Sesame's /debug/pprof service

Sesame exposes the [net/http/pprof][1]handlers for `go tool pprof` and `go tool trace` by default on `127.0.0.1:6060`.
This service is useful for profiling Sesame.
To access it from your workstation use `kubectl port-forward` like so,

```sh
# Get one of the pods that matches the Sesame deployment
Sesame_POD=$(kubectl -n projectsesame get pod -l app=sesame -o name | head -1)
# Do the port forward to that pod
kubectl -n projectsesame port-forward $Sesame_POD 6060
```

## Visualizing Sesame's internal directed acyclic graph (DAG)

Sesame models its configuration using a DAG, which can be visualized through a debug endpoint that outputs the DAG in [DOT][2] format.
To visualize the graph, you must have [`graphviz`][3] installed on your system.

To download the graph and save it as a PNG:

```sh
# Port forward into the sesame pod
Sesame_POD=$(kubectl -n projectsesame get pod -l app=sesame -o name | head -1)
# Do the port forward to that pod
kubectl -n projectsesame port-forward $Sesame_POD 6060
# Download and store the DAG in png format
curl localhost:6060/debug/dag | dot -T png > sesame-dag.png
```

The following is an example of a DAG that maps `http://kuard.local:80/` to the
`kuard` service in the `default` namespace:

![Sample DAG][4]

## Interrogate Sesame's gRPC API

Sometimes it's helpful to be able to interrogate Sesame to find out exactly the data it is sending to Envoy.
Sesame ships with a `Sesame cli` subcommand which can be used for this purpose.

Because Sesame secures its communications with Envoy using Secrets in the cluster, the easiest way is to run `Sesame cli` commands _inside_ the pod.
Do this is via `kubectl exec`:

```sh
# Get one of the pods that matches the examples/daemonset
Sesame_POD=$(kubectl -n projectsesame get pod -l app=sesame -o jsonpath='{.items[0].metadata.name}')
# Do the port forward to that pod
kubectl -n projectsesame exec $Sesame_POD -c sesame -- sesame cli lds --cafile=/ca/cacert.pem --cert-file=/certs/tls.crt --key-file=/certs/tls.key
```

Which will stream changes to the LDS api endpoint to your terminal.
Replace `Sesame cli lds` with `Sesame cli rds` for RDS, `Sesame cli cds` for CDS, and `Sesame cli eds` for EDS.

## I've deployed on Minikube or kind and nothing seems to work

See [the deployment documentation][5] for some tips on using these two deployment options successfully.

[1]: https://golang.org/pkg/net/http/pprof
[2]: https://en.wikipedia.org/wiki/DOT
[3]: https://graphviz.gitlab.io/
[4]: /img/kuard-dag.png
[5]: {% link docs/main/deploy-options.md %}