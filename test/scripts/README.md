# Sesame E2E tests

## Cluster setup
The [make-kind-cluster.sh](./make-kind-cluster.sh) script will bring up
a local kind cluster. This underlying VM [config](./kind-expose-port.yaml)
forwards the Envoy ports 80 and 443 locally as port 9080 and 9443.
The script installs [cert-manager](https://cert-manager.io), which is
needed for tests that use TLS.

The [install-sesame-working.sh](.install-sesame-working.sh) script
builds and installs Sesame from the working repository.

The [install-Sesame-release.sh](.install-sesame-release.sh) script
installs a specified Sesame release. This is useful for doing upgrade
testing. For example:

```bash
$ ./install-sesame-release.sh v1.9.0
...
```

## Running tests

To run the tests, it's best to install [ginkgo](https://onsi.github.io/ginkgo/) on your development machine.

The e2e tests deploy an Envoy Service and Daemonset in your local kind cluster and run Sesame as a process on your local machine, subscribing to k8s resources the tests create.

Some configurations are available to modify your test environment:
- the `SESAME_E2E_LOCAL_HOST` environment variable is required and must be set to an address Envoy can use to connect to the Sesame xDS server
- `SESAME_E2E_LOCAL_PORT` can be used to customize the port Sesame's xDS server will listen on, defaults to `8001`
- set the `KUBECONFIG` environment variable to provide Sesame a specific k8s config to use

To run a single test (spec):
```
ginkgo -tags=e2e -r -v -focus "001-required-field-validation" ./test/e2e
```

To run all tests for a given API:
```
ginkgo -tags=e2e -keepGoing -randomizeSuites -randomizeAllSpecs -r -v ./test/e2e/httpproxy
```

To run all tests for all APIs:
```
ginkgo -tags=e2e -keepGoing -randomizeSuites -randomizeAllSpecs -r -v ./test/e2e
```
