---

cascade:
  layout: docs
  gh: https://github.com/projectsesame/sesame/tree/release-1.13
  version: v1.13.1
---

## Overview
Sesame is an Ingress controller for Kubernetes that works by deploying the [Envoy proxy][1] as a reverse proxy and load balancer.
Sesame supports dynamic configuration updates out of the box while maintaining a lightweight profile.

Sesame also introduces a new ingress API [HTTPProxy][2] which is implemented via a Custom Resource Definition (CRD).
Its goal is to expand upon the functionality of the Ingress API to allow for a richer user experience as well as solve shortcomings in the original design.

## Prerequisites
Sesame is tested with Kubernetes clusters running version [1.16 and later][4].

RBAC must be enabled on your cluster.

## Get started
Getting started with Sesame is as simple as one command.
See the [Getting Started][3] document.

[1]: https://www.envoyproxy.io/
[2]: /docs/{{< param version >}}/config/fundamentals
[3]: /getting-started
[4]: /_resources/kubernetes