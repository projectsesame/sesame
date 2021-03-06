# Annotations Reference

<div id="toc" class="navigation"></div>

Annotations are used in Ingress Controllers to configure features that are not covered by the Kubernetes Ingress API.

Some of the features that have been historically configured via annotations are supported as first-class features in Sesame's [IngressRoute API][15], which provides a more robust configuration interface over
annotations.

However, Sesame still supports a number of annotations on the Ingress resources.

<p class="alert-deprecation">
<b>Deprecation Notice</b><br>
The <code>Sesame.heptio.com</code> annotations are deprecated, please use the <code>projectsesame.io</code> form going forward.
</p>

## Standard Kubernetes Ingress annotations

The following Kubernetes annotions are supported on [`Ingress`] objects:

 - `kubernetes.io/ingress.class`: The Ingress class that should interpret and serve the Ingress. If not set, then all Ingress controllers serve the Ingress. If specified as `kubernetes.io/ingress.class: Sesame`, then Sesame serves the Ingress. If any other value, Sesame ignores the Ingress definition. You can override the default class `Sesame` with the `--ingress-class-name` flag at runtime. This can be useful while you are migrating from another controller, or if you need multiple instances of Sesame.
 - `ingress.kubernetes.io/force-ssl-redirect`: Requires TLS/SSL for the Ingress to Envoy by setting the [Envoy virtual host option require_tls][16].
 - `kubernetes.io/ingress.allow-http`: Instructs Sesame to not create an Envoy HTTP route for the virtual host. The Ingress exists only for HTTPS requests. Specify `"false"` for Envoy to mark the endpoint as HTTPS only. All other values are ignored.

The `ingress.kubernetes.io/force-ssl-redirect` annotation takes precedence over `kubernetes.io/ingress.allow-http`. If they are set to `"true"` and `"false"` respectively, Sesame *will* create an Envoy HTTP route for the Virtual host, and set the `require_tls` virtual host option.

## Sesame specific Ingress annotations

 - `projectsesame.io/ingress.class`: The Ingress class that should interpret and serve the Ingress. If not set, then all Ingress controllers serve the Ingress. If specified as `projectsesame.io/ingress.class: Sesame`, then Sesame serves the Ingress. If any other value, Sesame ignores the Ingress definition. You can override the default class `Sesame` with the `--ingress-class-name` flag at runtime. This can be useful while you are migrating from another controller, or if you need multiple instances of Sesame.
 - `projectsesame.io/num-retries`: [The maximum number of retries][1] Envoy should make before abandoning and returning an error to the client. Applies only if `projectsesame.io/retry-on` is specified.
 - `projectsesame.io/per-try-timeout`: [The timeout per retry attempt][2], if there should be one. Applies only if `projectsesame.io/retry-on` is specified.
 - `projectsesame.io/response-timeout`: [The Envoy HTTP route timeout][3], specified as a [golang duration][4]. By default, Envoy has a 15 second timeout for a backend service to respond. Set this to `infinity` to specify that Envoy should never timeout the connection to the backend. Note that the value `0s` / zero has special semantics for Envoy.
 - `projectsesame.io/retry-on`: [The conditions for Envoy to retry a request][5]. See also [possible values and their meanings for `retry-on`][6].
 - `projectsesame.io/tls-minimum-protocol-version`: [The minimum TLS protocol version][7] the TLS listener should support.
 - `projectsesame.io/websocket-routes`: [The routes supporting websocket protocol][8], the annotation value contains a list of route paths separated by a comma that must match with the ones defined in the `Ingress` definition. Defaults to Envoy's default behavior which is `use_websocket` to `false`.
 - `Sesame.heptio.com/ingress.class`: deprecated form of `projectsesame.io/ingress.class`.
 - `Sesame.heptio.com/num-retries`: deprecated form of `projectsesame.io/num-retries`.
 - `Sesame.heptio.com/per-try-timeout`: deprecated form of `projectsesame.io/per-try-timeout`.
 - `Sesame.heptio.com/request-timeout`: deprecated form of `projectsesame.io/response-timeout`. _Note_ this is **response-timeout**.
 - `Sesame.heptio.com/retry-on`:  deprecated form of `projectsesame.io/retry-on`.
 - `Sesame.heptio.com/tls-minimum-protocol-version`: deprecated form of `projectsesame.io/tls-minimum-protocol-version`.
 - `Sesame.heptio.com/websocket-routes`: deprecated form of `projectsesame.io/websocket-routes`.

## Sesame specific Service annotations

A [Kubernetes Service][9] maps to an [Envoy Cluster][10]. Envoy clusters have many settings to control specific behaviors. These annotations allow access to some of those settings.

- `projectsesame.io/max-connections`: [The maximum number of connections][11] that a single Envoy instance allows to the Kubernetes Service; defaults to 1024.
- `projectsesame.io/max-pending-requests`: [The maximum number of pending requests][13] that a single Envoy instance allows to the Kubernetes Service; defaults to 1024.
- `projectsesame.io/max-requests`: [The maximum parallel requests][13] a single Envoy instance allows to the Kubernetes Service; defaults to 1024
- `projectsesame.io/max-retries`: [The maximum number of parallel retries][14] a single Envoy instance allows to the Kubernetes Service; defaults to 1024. This is independent of the per-Kubernetes Ingress number of retries (`projectsesame.io/num-retries`) and retry-on (`projectsesame.io/retry-on`), which control whether retries are attempted and how many times a single request can retry.
- `projectsesame.io/upstream-protocol.{protocol}` : The protocol used in the upstream. The annotation value contains a list of port names and/or numbers separated by a comma that must match with the ones defined in the `Service` definition. For now, just `h2`, `h2c`, and `tls` are supported: `Sesame.heptio.com/upstream-protocol.h2: "443,https"`. Defaults to Envoy's default behavior which is `http1` in the upstream.
  - The `tls` protocol allows for requests which terminate at Envoy to proxy via tls to the upstream. _Note: This does not validate the upstream certificate._
- `Sesame.heptio.com/max-connections`:  deprecated form of `projectsesame.io/max-connections`
- `Sesame.heptio.com/max-pending-requests`: deprecated form of `projectsesame.io/max-pending-requests`.
- `Sesame.heptio.com/max-requests`: deprecated form of `projectsesame.io/max-requests`.
- `Sesame.heptio.com/max-retries`: deprecated form of `projectsesame.io/max-retries`.
- `Sesame.heptio.com/upstream-protocol.{protocol}` : deprecated form of `projectsesame.io/upstream-protocol.{protocol}`.

## Sesame specific IngressRoute annotations
- `Sesame.heptio.com/ingress.class`: The Ingress class that should interpret and serve the IngressRoute. If not set, then all all Sesame instances serve the IngressRoute. If specified as `Sesame.heptio.com/ingress.class: Sesame`, then Sesame serves the IngressRoute. If any other value, Sesame ignores the IngressRoute definition. You can override the default class `Sesame` with the `--ingress-class-name` flag at runtime.

[1]: https://www.envoyproxy.io/docs/envoy/v1.11.2/configuration/http_filters/router_filter.html#config-http-filters-router-x-envoy-max-retries
[2]: https://www.envoyproxy.io/docs/envoy/v1.11.2/api-v2/api/v2/route/route.proto#envoy-api-field-route-routeaction-retrypolicy-retry-on
[3]: https://www.envoyproxy.io/docs/envoy/v1.11.2/api-v2/api/v2/route/route.proto.html#envoy-api-field-route-routeaction-timeout
[4]: https://golang.org/pkg/time/#ParseDuration
[5]: https://www.envoyproxy.io/docs/envoy/v1.11.2/api-v2/api/v2/route/route.proto#envoy-api-field-route-routeaction-retrypolicy-retry-on
[6]: https://www.envoyproxy.io/docs/envoy/v1.11.2/configuration/http_filters/router_filter.html#config-http-filters-router-x-envoy-retry-on
[7]: https://www.envoyproxy.io/docs/envoy/v1.11.2/api-v2/api/v2/auth/cert.proto#envoy-api-msg-auth-tlsparameters
[8]: https://www.envoyproxy.io/docs/envoy/v1.11.2/api-v2/api/v2/route/route.proto#envoy-api-field-route-routeaction-use-websocket
[9]: https://kubernetes.io/docs/concepts/services-networking/service/
[10]: https://www.envoyproxy.io/docs/envoy/v1.11.2/intro/arch_overview/intro/terminology.html
[11]: https://www.envoyproxy.io/docs/envoy/v1.11.2/api-v2/api/v2/cluster/circuit_breaker.proto#envoy-api-field-cluster-circuitbreakers-thresholds-max-connections
[12]: https://www.envoyproxy.io/docs/envoy/v1.11.2/api-v2/api/v2/cluster/circuit_breaker.proto#envoy-api-field-cluster-circuitbreakers-thresholds-max-pending-requests
[13]: https://www.envoyproxy.io/docs/envoy/v1.11.2/api-v2/api/v2/cluster/circuit_breaker.proto#envoy-api-field-cluster-circuitbreakers-thresholds-max-requests
[14]: https://www.envoyproxy.io/docs/envoy/v1.11.2/api-v2/api/v2/cluster/circuit_breaker.proto#envoy-api-field-cluster-circuitbreakers-thresholds-max-retries
[15]: ingressroute.md
[16]: https://www.envoyproxy.io/docs/envoy/v1.11.2/api-v2/api/v2/route/route.proto.html#envoy-api-field-route-virtualhost-require-tls
