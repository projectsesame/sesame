## xDS management connection between Sesame and Envoy set to TLSv1.3

The minimum accepted TLS version for Sesame xDS server is changed from TLSv1.2 to TLSv1.3.
Previously in Sesame 1.19, the maximum accepted TLS version for Envoy xDS client was increased to TLSv1.3 which allows it to connect to Sesame xDS server using TLSv1.3.

If upgrading from a version **prior to Sesame 1.19**, the old Envoys will be unable to connect to new Sesame until also Envoys are upgraded.
Until that, old Envoys are unable to receive new configuration data.

For further information, see [Sesame architecture](https://projectsesame.io/docs/main/architecture/) and [xDS API](https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol) in Envoy documentation.
