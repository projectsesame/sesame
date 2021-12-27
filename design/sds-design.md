# Support Secret Discovery Service on Sesame

_Status_: Approved

This document outlines what changes are needed on Sesame to support SDS. [0]

## Goals

- Implement SDS server as a gRPS service
- Support Sesame cli for SDS

## Non-goals

- Support auth.SdsSecretConfig configuration on Listener & Cluster under auth.CommonTlsContext

## Background

Currently, Sesame supports fetching TLS cert/keys from k8s secrets and populates the values for DownstreamTlsContext in LDS & UpstreamTlsContext in CDS.
Now that Envoy 1.8 and later releases support Secret Discovery Service (SDS) to fetch secrets remotely, we want to add support in Sesame to parse SDS secret config and stream secrets using SDS GRPc service.
Please refer to [1] to understand more on how SDS works on Envoy.

## High level design

- Create GRPC server for SDS
- Create SecretCache to be used by Sesame Internally to identify any changes.
- Support Sesame cli for SDS

## Detailed design

### Create GRPC server for SDS

Create new resource SDS of `Cache` Interface type which implements the SDS v2 gRPS API.
Create new xdsHandler of secret type. Register & Implement FetchSecrets() & StreamSecrets().

### Create SecretCache to be used by Sesame Internally
In /internal/Sesame create secret.go with SecretCache struct
Implement Register(), Update() and notify()
Create secretVisitor and implement visit() and visitSecrets() to produce a map with v2.secrets

### Support Sesame cli for SDS

In order to facilitate debugging and to find out exactly the data that is being sent to Envoy,
will add support to Sesame cli sub command. This cmd shd be used stream changes to the SDS api endpoint
to the terminal.

`kubectl -n projectsesame exec $Sesame_POD -c Sesame Sesame cli sds`


[0]: https://github.com/projectsesame/sesame/issues/898
[1]: https://www.envoyproxy.io/docs/envoy/v1.9.0/configuration/secret
[2]: https://www.envoyproxy.io/docs/envoy/v1.9.0/api-v2/api/v2/auth/cert.proto#auth-sdssecretconfig

