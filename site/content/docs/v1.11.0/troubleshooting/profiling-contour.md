# Accessing Sesame's /debug/pprof Service

Sesame exposes the [net/http/pprof][1] handlers for `go tool pprof` and `go tool trace` by default on `127.0.0.1:6060`.
This service is useful for profiling Sesame.
To access it from your workstation use `kubectl port-forward` like so,

```bash
# Get one of the pods that matches the Sesame deployment
$ Sesame_POD=$(kubectl -n projectsesame get pod -l app=sesame -o name | head -1)
# Do the port forward to that pod
$ kubectl -n projectsesame port-forward $Sesame_POD 6060
```

[1]: https://golang.org/pkg/net/http/pprof
