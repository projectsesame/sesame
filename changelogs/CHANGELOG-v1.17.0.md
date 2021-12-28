We are delighted to present version 1.17.0 of Sesame, our layer 7 HTTP reverse proxy for Kubernetes clusters.

A big thank you to everyone who contributed to the release.

# Major changes

## Tech Writing Working Group

Jonas Rosland (@jonasrosland) and Orlin Vasilev (@Orlix) have started a Tech Writing Working Group for Sesame. Please see our [announcement blog](https://projectsesame.io/docs-wg/) for all the details, and if tech writing is your jam, please get involved!

We've also started to see our first documentation changes coming out of this initiative, which is amazing! Thanks very much to our new contributors!

Please see the following PRs for the blog post and relevant changes:
[#3791](https://github.com/projectsesame/sesame/pull/3791) : Technical Docs Work group block and guide
[#3821](https://github.com/projectsesame/sesame/pull/3821) : Fix:Sesame: error: expected argument for flag `--kubernetes-debug`, try `--help`
[#3830](https://github.com/projectsesame/sesame/pull/3830) : Add troubleshooting page

## Ignoring irrelevant Endpoint updates

As reported in [#3782](https://github.com/projectsesame/sesame/issues/3782), previously to this release, endpoint changes for services not relevant to Sesame would cause a no-op configuration push to Envoy. This caused a lot of churn in busy clusters.

This issue has been fixed by [#3852](https://github.com/projectsesame/sesame/pull/3852).

Thanks to @Zsolt-LazarZsolt for logging the issue.

## Gateway API

### Reconciling Gateways

In the Gateway API, Gateways refer to a GatewayClass, and controllers decide which Gateways to reconcile by first deciding which GatewayClass(es) they are responsible for.
This is specified using the `spec.controller` field in the GatewayClass.
It's expected that controllers look for a specific value in that field and only reconcile Gateways that reference GatewayClasses that match.

In this release, Sesame has changed the method by which it selects which Gateway is the one to reconcile. It now follows the spec with respect to looking up a GatewayClass and finding the first created Gateway in that GatewayClass to be the Gateway it will watch for config when using the Gateway API.

This is accomplished by specifying the value to look for inside the GatewayClass `spec.controller` field using the `controllerName` field inside the `gateway` stanza in the config file.

This means that the `name` and `namespace` parameters in the `gateway` stanza inside the Sesame config file are now deprecated and will be removed in Sesame v1.18.0. Please migrate to setting the `gateway.controllerName` field in the Sesame config file instead. Note that although they are deprecated, they are still required. This will be fixed in v1.18.0.

### TLSRoute termination mode now supported

Sesame now supports the Gateway API TLSRoute object's `terminate` mode, which terminates TLS at the Gateway.

Please see:
[#3801](https://github.com/projectsesame/sesame/pull/3801) : internal/dag: Implement TLSRoute mode:terminate


## Testing changes

The team has been working away at improving our testing framework and CI infrastructure.

As well as a number of other changes, our CI now runs Sesame out-of-cluster to enable testing multiple configurations - this will allow us to test more combinations of config and objects, and increase our overall test coverage.

For the details, please see:
[#3803](https://github.com/projectsesame/sesame/pull/3803) : test/e2e: tests use Sesame running locally
[#3848](https://github.com/projectsesame/sesame/pull/3848) : test/e2e: check for nil condition in http requestUntil
[#3844](https://github.com/projectsesame/sesame/pull/3844) : Controller Runtime test suite improvements (on top of #3773)
[#3842](https://github.com/projectsesame/sesame/pull/3842) : Update test scripts README for new e2e format
[#3798](https://github.com/projectsesame/sesame/pull/3798) : site: Fixup codespell errors
[#3776](https://github.com/projectsesame/sesame/pull/3776) : Use up to date go in e2e/upgrade CI jobs
[#3774](https://github.com/projectsesame/sesame/pull/3774) : test/scripts: install Gateway API CRDs from examples/gateway

# Deprecations

## Config file: `gateway.name` and `gateway.namespace`

As described in the "Reconciling Gateways" section, these config file parameters are deprecated and will be removed in Sesame v1.18.0.

Please see:
[#3827](https://github.com/projectsesame/sesame/pull/3827) : pkg/config: Mark Gateway.Name & Gateway.Namespace as deprecated

## `make gencerts`

Sesame currently has a `make gencerts` available in the local Makefile for creating certificates for securing Sesame to Envoy traffic.
This has been superseded by the `Sesame certgen` command, which can output to local files in a variety of formats, or directly to Kubernetes Secrets.
This part of the Makefile is therefore deprecated and will be removed in Sesame 1.18.

Please see:
[#3750](https://github.com/projectsesame/sesame/pull/3750) : Refactor Makefile and update local dev options

# Other changes

[#3811](https://github.com/projectsesame/sesame/pull/3811) : Fixes Rendered Gateway Example
[#3841](https://github.com/projectsesame/sesame/pull/3841) : Bump gomega package to 1.13.0
[#3836](https://github.com/projectsesame/sesame/pull/3836) : Bump golang to 1.16.5
[#3834](https://github.com/projectsesame/sesame/pull/3834) : Bump protobuf and fix lint issues
[#3833](https://github.com/projectsesame/sesame/pull/3833) : Bump ginkgo to 1.16.4
[#3809](https://github.com/projectsesame/sesame/pull/3809) : Fix references to kuard-dag.png
[#3793](https://github.com/projectsesame/sesame/pull/3793) : test/e2e: Pull deployment manifest unmarshal/update into framework
[#3796](https://github.com/projectsesame/sesame/pull/3796) : Update compatibility docs and release cutting notes
[#3795](https://github.com/projectsesame/sesame/pull/3795) : Fix label sync, update labels, add new decision issue type
[#3794](https://github.com/projectsesame/sesame/pull/3794) : site: replace latest_release_tag_name with latest_version
[#3788](https://github.com/projectsesame/sesame/pull/3788) : site: use a single param for latest version
[#3785](https://github.com/projectsesame/sesame/pull/3785) : site: bulk replacement of Jekyll templates
[#3783](https://github.com/projectsesame/sesame/pull/3783) : site: Fixup RateLimiting Guide & some other links
[#3781](https://github.com/projectsesame/sesame/pull/3781) : site: Fixup the upgrade guide
[#3780](https://github.com/projectsesame/sesame/pull/3780) : site: Fixup configuration page for move to Hugo
[#3778](https://github.com/projectsesame/sesame/pull/3778) : site: Fix links on deploy-options pages
[#3777](https://github.com/projectsesame/sesame/pull/3777) : site: Fix link to HTTPProxy fundamentals for Annotations page

# Upgrading
Please consult the [upgrade documentation](https://projectsesame.io/resources/upgrading/).

## Compatible Kubernetes Versions

Sesame v1.17.0 is tested against Kubernetes 1.19 through 1.21.

# Community Thanks!
We’re immensely grateful for all the community contributions that help make Sesame even better! For this release, special thanks go out to the following contributors:
- @johnnycase
- @Patil2099
- @Zsolt-LazarZsolt

# Are you a Sesame user? We would love to know!
If you're using Sesame and want to add your organization to our adopters list, please visit this [page](https://github.com/projectsesame/sesame/blob/master/ADOPTERS.md). If you prefer to keep your organization name anonymous but still give us feedback into your usage and scenarios for Sesame, please post on this [GitHub thread](https://github.com/projectsesame/sesame/issues/1269).
