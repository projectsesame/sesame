# Sesame Installation

This is an installation guide to configure Sesame in a Deployment separate from Envoy which allows for easier scaling of each component.

This configuration has several advantages:

1. Envoy runs as a DaemonSet which allows for distributed scaling across workers in the cluster
2. Communication between Sesame and Envoy is secured by mutually-checked self-signed certificates.

## Moving parts

- Sesame is run as Deployment and Envoy as a DaemonSet
- Envoy runs on host networking
- Envoy runs on ports 80 & 443

The TLS secrets used to secure the gRPC session between Sesame and Envoy are generated using a Job that runs `sesame certgen`.
For detailed instructions on how to configure the required secrets manually, see the [step-by-step TLS HOWTO](https://projectsesame.io/docs/main/grpc-tls-howto).

## Deploy Sesame

Either:

1. Run `kubectl apply -f https://projectsesame.io/quickstart/sesame.yaml`

or:
Clone or fork the repository, then run:

```bash
kubectl apply -f examples/sesame
```

This will:

- set up RBAC and Sesame's CRDs (CRDs include HTTPProxy, TLSCertificateDelegation)
- run a Kubernetes Job that will generate one-year validity certs and put them into `projectsesame`
- Install Sesame and Envoy in a Deployment and DaemonSet respectively.

**NOTE**: The current configuration exposes the `/stats` path from the Envoy Admin UI so that Prometheus can scrape for metrics.

## Test

1. Install a workload (see the kuard example in the [main deployment guide](https://projectsesame.io/docs/main/deploy-options/#test-with-httpproxy)).

## Deploying with Host Networking enabled for Envoy

In order to deploy the Envoy DaemonSet with host networking enabled, you need to make two changes.

In the Envoy daemonset definition, at the Pod spec level, change:

```yaml
dnsPolicy: ClusterFirst
```

to

```yaml
dnsPolicy: ClusterFirstWithHostNet
```

and add

```yaml
hostNetwork: true
```

Then, in the Envoy Service definition, change the annotation from:

```yaml
  # This annotation puts the AWS ELB into "TCP" mode so that it does not
  # do HTTP negotiation for HTTPS connections at the ELB edge.
  # The downside of this is the remote IP address of all connections will
  # appear to be the internal address of the ELB. See docs/proxy-proto.md
  # for information about enabling the PROXY protocol on the ELB to recover
  # the original remote IP address.
  service.beta.kubernetes.io/aws-load-balancer-backend-protocol: tcp
```

to

```yaml
   service.beta.kubernetes.io/aws-load-balancer-type: nlb
```

Then, apply the example as normal. This will still deploy a LoadBalancer Service, but it will be an NLB instead of an ELB.
