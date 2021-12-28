# Interrogate Sesame's xDS Resources

Sometimes it's helpful to be able to interrogate Sesame to find out exactly what [xDS][1] resource data it is sending to Envoy.
Sesame ships with a `Sesame cli` subcommand which can be used for this purpose.

Because Sesame secures its communications with Envoy using Secrets in the cluster, the easiest way is to run `Sesame cli` commands _inside_ the pod.
Do this is via `kubectl exec`:

```bash
# Get one of the pods that matches the examples/daemonset
$ Sesame_POD=$(kubectl -n projectsesame get pod -l app=sesame -o jsonpath='{.items[0].metadata.name}')
# Do the port forward to that pod
$ kubectl -n projectsesame exec $Sesame_POD -c sesame -- sesame cli lds --cafile=/certs/ca.crt --cert-file=/certs/tls.crt --key-file=/certs/tls.key
```

Which will stream changes to the LDS api endpoint to your terminal.
Replace `Sesame cli lds` with `Sesame cli rds` for route resources, `Sesame cli cds` for cluster resources, and `Sesame cli eds` for endpoints.

[1]: https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol
