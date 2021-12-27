We are delighted to present version 1.18.1 of Sesame, our layer 7 HTTP reverse proxy for Kubernetes clusters.

# Fixes

## Envoy Admin Interface Changes to Mitigate CVE-2021-32783

As an additional mitigation for [CVE-2021-32783](https://github.com/projectsesame/sesame/security/advisories/GHSA-5ph6-qq5x-7jwc), [PR #3934](https://github.com/projectsesame/sesame/pull/3934) has been backported to this release. Previously an ExternalName service with an address of `localhost` and port matching the Envoy admin interface listener port could expose admin endpoints that would allow an attacker to shut down Envoy remotely. Sesame's bootstrap command now configures Envoy to listen on a Unix domain socket to ensure an ExternalName service cannot expose the "writable" endpoints of the admin interface. For backwards compatibility Sesame now configures an additional Envoy listener to expose endpoints useful for debugging. This listener does not expose any endpoints that can be used to modify data, set Envoy healthcheck status, or shut down Envoy. See [this documentation page](https://projectsesame.io/docs/v1.18.1/troubleshooting/envoy-admin-interface/) for some more detail.

*Note: While this fix does mitigate some aspects of CVE-2021-32783, ExternalName service usage is still discouraged as they can still be used to expose services across namespaces.*

If you have been using the Sesame example YAMLs to deploy Sesame and Envoy, no changes are required except to deploy the updated YAMLs for release 1.18.1.

If you are managing your Envoy DaemonSet using another method, be sure to inspect the [updated YAML for required changes](https://github.com/projectsesame/sesame/blob/v1.18.1/examples/Sesame/03-envoy.yaml). An [additional volume](https://github.com/projectsesame/sesame/blob/v1.18.1/examples/Sesame/03-envoy.yaml#L134-L135) is required and it must be mounted in the [`shutdown-manager`](https://github.com/projectsesame/sesame/blob/v1.18.1/examples/Sesame/03-envoy.yaml#L48-L50) and [`envoy`](https://github.com/projectsesame/sesame/blob/v1.18.1/examples/Sesame/03-envoy.yaml#L95-L96) containers to ensure both have access to the Unix domain socket Envoy is now configured to listen on.

## Envoy Updated to 1.19.1

Upgrades the default Envoy version to 1.19.1 for security and bug fixes. See the [Envoy 1.19.1 changelogs](https://www.envoyproxy.io/docs/envoy/v1.19.1/version_history/current) for more details.
