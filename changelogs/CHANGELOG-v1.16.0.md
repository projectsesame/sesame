We are delighted to present version 1.16.0 of Sesame, our layer 7 HTTP reverse proxy for Kubernetes clusters.

A big thank you to everyone who contributed to the release.

# Major Changes

## Gateway API

### Added Initial Support for TLSRoute
PR https://github.com/projectsesame/sesame/pull/3627 added support for TLSRoute to enable Passthrough TCP Proxying to pods via SNI. See https://github.com/projectsesame/sesame/issues/3440 for additional information on TLSRoute feature requirements.

### GatewayClass Support
PR https://github.com/projectsesame/sesame/pull/3659 added GatewayClass support through the gateway.controllerName configuration parameter. The gateway.namespace and gateway.name parameters are required when setting gateway.controllerName. When the cluster contains Gateway API CRDs and this parameter is set, Sesame will reconcile GatewayClass resources with the spec.controller field that matches gateway.controllerName. ControllerName should take the form of `projectsesame.io/<namespace>/Sesame`, where `<namespace>` is the namespace that Sesame runs in.

## CA is No Longer Ignored when Downstream "Skip Verify" is True
With PR https://github.com/projectsesame/sesame/pull/3661, Sesame will no longer ignore a certificate under the following conditions:
  - If no CA is set and "skip verify" is true, client certs are not required by Envoy.
  - If CA set and "skip verify" is true, client certs are required by Envoy.
  - CA is still required if "skip verify" is false.
  - `caCert` is now optional since skipClientCertValidation can be true. PR https://github.com/projectsesame/sesame/pull/3658 added an `omitempty` JSON tag to omit the `caCert` field when serializing to JSON and it hasn't been specified.  

## Website Update
PR https://github.com/projectsesame/sesame/pull/3704 revamps the Sesame website based on [Hugo](https://gohugo.io/). Check out the fresh new look and tell us what you think.

# Deprecation & Removal Notices
- PR https://github.com/projectsesame/sesame/pull/3642 removed the `experimental-service-apis` flag has been removed. The gateway.name & gateway.namespace in the Sesame configuration file should be used for configuring Gateway API (formerly Service APIs).
- PR https://github.com/projectsesame/sesame/pull/3645 removed support for Ingress v1beta1. Ingress v1 resources should be used with Sesame.

# Upgrading
Please consult the [upgrade documentation](https://projectsesame.io/resources/upgrading/).

## Community Thanks!
We’re immensely grateful for all the community contributions that help make Sesame even better! For this release, special thanks go out to the following contributors:
- @geoffcline 
- @pandeykartikey
- @pyaillet

# Are you a Sesame user? We would love to know!
If you're using Sesame and want to add your organization to our adopters list, please visit this [page](https://github.com/projectsesame/sesame/blob/master/ADOPTERS.md). If you prefer to keep your organization name anonymous but still give us feedback into your usage and scenarios for Sesame, please post on this [GitHub thread](https://github.com/projectsesame/sesame/issues/1269).
