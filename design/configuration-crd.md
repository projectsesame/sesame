# Sesame Configuration CRD

Currently, Sesame gets its configuration from two different places, one is the configuration file represented as a Kubernetes configmap.
The other are flags which are passed to Sesame.

Sesame's configmap configuration file has grown to the point where moving to a CRD will enable a better user experience as well as allowing Sesame to react to changes in its configuration faster.

This design proposes two new CRDs, one that represents a `SesameConfiguration` (Short name `SesameConfig`) and another which represents a `SesameDeployment`, both which are namespaced.
The Sesame configuration mirrors how the configmap functions today. 
A Sesame Deployment is managed by a controller (aka Operator) which uses the details of the CRD spec to deploy a fully managed instance of Sesame inside a cluster.

## Benefits
- Eliminates the need to translate from a CRD to a configmap (Like the Operator does today)
- Allows for a place to surface information about configuration errors - the CRD status, in addition to the Sesame log files
- Allows the Operator workflow to match a non-operator workflow (i.e. you start with a Sesame configuration CRD)
- Provides better validation to avoid errors in the configuration file
- Dynamic restarting of Sesame when configuration changes

## New CRD Spec

The contents of current configuration file has grown over time and some fields need to be better categorized.
- Any field that was not previously `camelCased` will be renamed to match.
- All the non-startup required command flags have been moved to the configuration CRD.
- New groupings of similar fields have been created to make the file flow better.

```yaml
apiVersion: projectsesame.io/v1alpha1
kind: SesameConfiguration
metadata:
  name: sesame
spec:
  xdsServer:
    type: sesame
    address: 0.0.0.0
    port: 8001
    insecure: false
    tls:
      caFile: 
      certFile:
      keyFile:
  ingress:
    className: sesame
    statusAddress: local.projectsesame.io
  debug:
    address: 127.0.0.1
    port: 6060
    logLevel: Info 
    kubernetesLogLevel: 0
  health:
    address: 0.0.0.0
    port: 8000
  metrics:
    address: 0.0.0.0
    port: 8002    
  envoy:
    listener:
      useProxyProtocol: false
      disableAllowChunkedLength: false
      connectionBalancer: exact
      tls:
        minimumProtocolVersion: "1.2"
        cipherSuites:
          - '[ECDHE-ECDSA-AES128-GCM-SHA256|ECDHE-ECDSA-CHACHA20-POLY1305]'
          - '[ECDHE-RSA-AES128-GCM-SHA256|ECDHE-RSA-CHACHA20-POLY1305]'
          - 'ECDHE-ECDSA-AES256-GCM-SHA384'
          - 'ECDHE-RSA-AES256-GCM-SHA384'
    service:
      name: sesame
      namespace: projectsesame
    http: 
      address: 0.0.0.0
      port: 80
      accessLog: /dev/STDOUT
    https:
      address: 0.0.0.0
      port: 443
      accessLog: /dev/STDOUT
    metrics:
      address: 0.0.0.0
      port: 8002
    clientCertificate:
      name: envoy-client-cert-secret-name
      namespace: projectsesame
    network:
      numTrustedHops: 0
      adminPort: 9001
    managed:
      networkPublishing:
      nodePlacement:
        nodeSelector:
        tolerations:
    logging:
      accessLogFormat: envoy
      accessLogFormatString: "...\n"
      jsonFields:
        - <Fields Omitted)
    defaultHTTPVersions:
      - "HTTP/2"
      - "HTTP/1.1"
    timeouts:
      requestTimeout: infinity
      connectionIdleTimeout: 60s
      streamIdleTimeout: 5m
      maxConnectionDuration: infinity
      delayedCloseTimeout: 1s
      connectionShutdownGracePeriod: 5s
    cluster:
      dnsLookupFamily: auto
  gateway:
      controllerName: projectsesame.io/projectsesame/sesame
  httpproxy:
    disablePermitInsecure: false
    rootNamespaces: 
      - foo
      - bar
    fallbackCertificate:
      name: fallback-secret-name
      namespace: projectsesame
  leaderElection:
    configmap:
      name: leader-elect
      namespace: projectsesame
    disableLeaderElection: false
  enableExternalNameService: false
  rateLimitService:
    extensionService: projectsesame/ratelimit
    domain: sesame
    failOpen: false
    enableXRateLimitHeaders: false
  policy:
    requestHeaders:
      set:
    responseHeaders:
      set:
status:
```

## Converting from Configmap

Sesame will provide a way internally to move to the new CRD and not require users to manually migrate to the new CRD format.
Sesame will provide a new command or external tool (similar to ir2proxy) which will migrate between the specs accordingly. 

## Sesame Deployment

A managed version of Sesame was made available with the `Sesame Operator`.
Since Sesame will manage Envoy instances, the Operator will now manage instances of Sesame.
The details of how an instance of Sesame should be deployed within a cluster will be defined in the second CRD named `SesameDeployment`. 
The `spec.confguration` of this object will be the same struct defined in the `SesameConfiguration`. 

A controller will watch for these objects to be created and take action on them accordingly to make desired state in the cluster match the configuration on the spec. 

```yaml
apiVersion: projectsesame.io/v1alpha1
kind: SesameDeployment
metadata:
  name: sesame
spec:
  replicas: 2
  nodePlacement:
    nodeSelector:
    tolerations:
  configuration:
    <same config as above>
status:
```

## Processing Logic

Sesame will require a new flag (`--Sesame-config`), which will allow for customizing the name of the `SesameConfiguration` CRD that is it to use.
It will default to one named `Sesame`, but could be overridden if desired.
The SesameConfiguration referenced must also be in the same namespace as Sesame is running, it's not valid to reference a configuration from another namespace.
The current flag `--config-path`/`-c` will continue to point to the Configmap file, but over time could eventually be deprecated and the short code `-c` be used for the CRD location (i.e. `--Sesame-config`) for simplicity.
The Sesame configuration CRD will still remain optional.
In its absence, Sesame will operate with reasonable defaults.
Where Sesame settings can also be specified with command-line flags, the command-line value takes precedence over the configuration file.

On startup, Sesame will look for a `SesameConfiguration` CRD named `Sesame` or using the value of the flag `--Sesame-config` in the same namespace which Sesame is running in, Sesame won't support referencing from a different namespace.
If the `SesameConfiguration` CRD is not found Sesame will start up with reasonable defaults.

Sesame will set status on the object informing the user if there are any errors or issues with the config file or that is all fine and processing correctly.
If the configuration file is not valid, Sesame will not start up its controllers, and will fail its readiness probe.
Once the configuration is valid again, Sesame will start its controllers with the valid configuration.

Once Sesame begins using a Configuration CRD, it will add a finalizer to it such that if that resource is going to get deleted, Sesame is aware of it.
Should the Configuration CRD be deleted while it is in use, Sesame will default back to reasonable defaults and log the issue.

When config in the CRD changes we will gracefully stop the dependent ingress/gateway controllers and restart them with new config, or dynamically update some in-memory data that the controllers use.
Sesame will first validate the new Configuration, ff that new change set results in the object being invalid, Sesame will stop its Controller and will become not ready, and not serve any xDS traffic.
As soon as the configuration does become valid, Sesame will start up its controllers and begin processing as normal.

### Initial Implementation

Sesame will initially start implementation by restarting the Sesame pod and allowing Kubernetes to restart itself when the config file changes.
Should the configuration be invalid, Sesame will start up, set status on the SesameConfig CRD and then crash.
Kubernetes will crash-loop until the configuration is valid, however, due to the nature of the exponential backoff, updates to the Configuration CRD will not be realized until the next restart loop, or Sesame is restarted manually. 

## Versioning

Initially, the CRDs will use the `v1alpha1` api group to allow for changes to the specs before they are released as a full `v1` spec. 
It's possible that we find issues along the way developing the new controllers and migrating to this new configuration method. 
Having a way to change the spec should we need to will be helpful on the path to a full v1 version. 

Once we get to `v1` we have hard compatibility requirements - no more breaking changes without a major version rev.
This should result in increased review scrutiny for proposed changes to the CRD spec.