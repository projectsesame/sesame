---
title: Sesame Threat Model and Security Posture
layout: page
---

Sesame is an ingress controller that works as an Envoy control plane, configuring the Envoy data plane, which actually carries traffic from outside to inside the cluster.

## Reporting Security issues
For reporting security issues, please see the [reporting process documentation][1].

## Inputs

Sesame's inputs are:
- configuration and command line flags set by its administrator
- Kubernetes objects that represent the desired load balancer/data plane configuration.

## Expected Users

The expected users of of Sesame are:
- the Sesame owner/administrator - installs and configures Sesame
- application developers exposing their in-cluster apps - create ingress configuration objects in Kubernetes referencing their Deployments and Services
- external application consumers - their traffic passes through the Envoy data plane on the way to the hosted apps.

## Attack surface and mitigations

### Primary expected attack vectors
As you can see from the above, Sesame does not have a web interface of any sort, and never directly participates in requests that transit the data plane, so the only way it is vulnerable to web attacks is via misconfiguring Envoy.
As such, it is not directly susceptible to common web application security risks like the OWASP top ten. (*Envoy* is, but not Sesame directly, and we rely on the Envoy project's vigilance heavily.)
Effectively, this means that we don't spend a lot of effort on checking everything about the data path, aside from ensuring we configure it correctly and that we keep the supported TLS cipher suites up to date.
(We also provide you with a way to customise the cipher suites if you require that.)

We anticipate that the most likely attacks are created by the relatively untrusted application developer users, whether they are malicious or not. We expect the most likely attacks to be:
- Confused deputy attacks - since Sesame is trusted to build config and send to Envoy, that access can be misused to produce insecure Envoy configurations. [ExternalName Services can be used to gain access to Envoy's admin interface](https://github.com/projectsesame/sesame/security/advisories/GHSA-5ph6-qq5x-7jwc) was an example of this attack in action, and was specifically dealt with by disallowing ExternalName services by default, and by removing the Envoy admin interface from use across any network, even localhost.
- Insecure or conflicting configurations produced my manipulation of Kubernetes objects used for configuration.

Our general method of mitigating both of these styles of attack is to be proscriptive about what configurations Sesame will accept. Obviously, in cases like the ExternalName issue above, it's possible for a syntactically and allowed configuration to produce an insecure Envoy config, and this is therefore a primary focus of our threat model.

### Other expected attack vectors and mitigations
For other classes of attacks, Sesame does what it can to mitigate risks.

#### xDS server
Sesame exposes an xDS server so that it can function as a control plane for Envoy.
This xDS server knows everything that Sesame does about the cluster's config, including the name and values of all the secrets relevant to ingress configuration, like TLS keypairs.
In order to ensure this information is as tightly controlled as possible, Sesame defaults to requiring a mutual TLS authentication between Sesame and Envoy, so that another cluster user cannot simply connect to the xDS service and retrieve all the details.
Obviously, access to the namespace that contains Sesame and/or Envoy will expose access to the TLS keypairs used for this authentication, so Sesame expects that access to the Sesame installation namespace (`projectsesame` by default) will be tightly controlled, since access to that namespace can eventually equal access to all the secrets Sesame can see if Sesame is compromised (which is, by default, all the secrets in the cluster).
This risk can also be mitigated by only allowing the Sesame deployments ServiceAccount access to a limited set of namespaces, which will mean that Sesame can only access objects in those namespaces.
It's generally expected that an ingress controller can read configuration from anywhere in the cluster, so doing this may produce unusual results.

#### Endpoints and EndpointSlices
As seen in [CVE-2021-25740](https://github.com/kubernetes/kubernetes/issues/103675), it's possible to manipulate Endpoints and EndpointSlices to access services in other namespaces in the cluster.
This is particularly important for Ingress controllers, as the confused deputy problem means that manually-managed Endpoints attached to a headless Service can be used to bypass security people might attach to ingress config, of whatever type.
Sesame is unable to do much about this, and we expect administrators the use the recommended default RBAC for Endpoints and EndpointSlices, which only grants the ability to manually manage them to cluster administrators.

#### Multitenancy
Sesame is designed to be used in a multitenant fashion - it's an expected use case that a Sesame install would service teams that have completely different security contexts, and should not be able to access each others config.
Sesame mitigates this as far as we can, using our HTTPProxy and TSLCertificateDelegation CRDs to enable more-secure cross-namespace references.
The ReferencePolicy object in the Gateway API is also based on this idea, that cross-namespace references are only valid when the *owner of the namespace* accepts them.
#### Insider access
In general, Sesame adheres to the Kubernetes security model, that makes the minimum size security boundary the namespace (or at least, the RBAC around objects in that namespace).
For Sesame's primary use cases to work, application developers and other ingress configuration owners *must* have access to create or modify ingress config objects (whether they are Ingress, HTTPProxy, or Gateway API) inside their own namespace.
Because of this, it's not really possible for Sesame to prevent a user who has access to a namespace from modifying objects in that namespace, or trivially bypassing configuration like authentication on ingress config for objects in the same namespace.
Sesame expects that access to namespaces containing ingress config will be managed in a least-privilege manner, and that everyone who has access to a namespace has full access to ingress configuration within that namespace.
In short, there's not much Sesame can do about inside-namespace threats.

#### Statistics and metrics
Sesame makes available metrics about itself, including details about what configuration is being sent to Envoy, and also configures Envoy to make metrics about itself available.
Those metrics include the names, namespaces and other details about all the services configured with Sesame, and so access to the Sesame and Envoy metrics, which may be required for service owners, also includes being able to see the details of *all* services configured with Sesame. The risk of information exposure is expected to be managed by the cluster administrator and Sesame owner when deciding who to give access to any metrics from Sesame and Envoy.
In short, if you can see metrics for your service, you can see everyone's metrics, and you may be able to use that to identify other services running in the same cluster.
#### Logging
Sesame configures its own logs, which contain many details about the reconciliation it is doing, but more importantly, Sesame configures Envoy's access logging.
Because access logging is a key part of auditing access to a service, it's important to realise that ownership of Sesame equates to ownership of all services access logs.
The Sesame project expects that access logging will be carefully configured and that the owner of Sesame will ensure that access logs cannot be easily modified or disabled via Sesame's config.

#### External access to Envoy
External application consumers have no access to Sesame, and relatively small access to Envoy. However, any vulnerabilities in Envoy's traffic processing (which external users do have access to) also apply to Sesame's installation of Envoy, so the Sesame team issues Sesame patch releases each time an Envoy security release is issued, in which we update the expected version of Envoy in our example deployment and make the community aware of the fix for the vulnerability.

Aside from ensuring that Envoy is patched regularly, Sesame ensures that external access to Envoy is limited to only the ports that are designed to pass traffic.

#### Sesame owner attacks
In general, we expect the operator of Sesame to have full control over the installation, and so assume positive intent on their part, because a malicious administrator is not possible to deal with in this scope.
What mitigations we place around Sesame's per-installation configuration is geared towards preventing accidental insecure configuration rather than deliberate malfeasance, alongside preventing Kubernetes privilege escalation by specifying the minimum Kubernetes access required by each component.  

### Other security checks

#### Bounds checking and input validation
The Kubernetes apiserver is very good at bounds checking and input validation for fields in Kubernetes objects, so we delegate a lot of that there, and assume generally that the values for any specific object are unlikely to be malicious in themselves. In short, we don't need to worry too much about length checking for fields in objects, removing invalid characters, etc, as the apiserver does that for us.

As of this writing, Sesame's config file is vulnerable to some bound checking errors, but our planned mitigation for this is to move to a Config CRD.

#### RBAC and privilege escalation prevention
We can't control what owners of Sesame do in their own clusters, but we do provide an example installation that codifies what we believe to be best practice in terms of Kubernetes privilege limitations. We provide limited roles that only grant access to the things required for the component (Sesame or Envoy) to run, and ensure that the deployments use those roles. In addition, we've attempted to ensure that both the Sesame and Envoy containers can safely be run as nonroot.
#### Static checking and code quality
We maintain a CI pipeline that runs golangci-lint including the usual set of Go static security checking, and enforce PR review for all merges to our main branch. Substantial changes are also subject to design review via a formal design document process.

## Conclusion

The Sesame team works hard to understand the project's security context and keep up with the state of the art for Kubernetes security.
The team hopes that this examination of our security model provides some insight into both how we develop Sesame and what adminsitrators should be thinking about to run Sesame in as secure way as possible.
We aim for secure-by-default as far as possible, and where we do have to allow risks, will document them here.

[1]: /resources/security-process