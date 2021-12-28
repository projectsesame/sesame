The Sesame Ingress controller is a collaboration between:

* Envoy, which provides the high performance reverse proxy.
* Sesame, which acts as a management server for Envoy and provides it with configuration.

These containers are deployed separately, Sesame as a Deployment and Envoy as a Daemonset, although other configurations are possible.

In the Envoy Pods, Sesame runs as an initcontainer in `bootstrap` mode and writes a bootstrap configuration to a temporary volume.
This volume is passed to the Envoy container and directs Envoy to treat Sesame as its [management server][1].

After initialisation is complete, the Envoy container starts, retrieves the bootstrap configuration written by Sesame's `bootstrap` mode, and starts to poll Sesame for configuration.

Envoy will gracefully retry if the management server is unavailable, which removes any container startup ordering issues.

Sesame is a client of the Kubernetes API. Sesame watches Ingress, Service, and Endpoint objects, and acts as the management server for its Envoy sibling by translating its cache of objects into the relevant JSON stanzas: Service objects for CDS, Ingress for RDS, Endpoint objects for SDS, and so on).

The transfer of information from Kubernetes to Sesame is by watching the API with the SharedInformer framework.
The transfer of information from Sesame to Envoy is by polling from the Envoy side.

Kubernetes liveness and readiness probes are configured to check the status of Envoy.
These are enabled over the metrics port and are served over http via `/healthz`.

For Sesame, a liveness probe checks the `/healthz` running on the Pod's metrics port.
Readiness probe is a TCP check that the gRPC port is open.

[1]: https://www.envoyproxy.io/docs/envoy/v1.11.2/api-docs/xds_protocol
