# Sesame ![Build and Test Pull Request](https://github.com/projectsesame/sesame/workflows/Build%20and%20Test%20Pull%20Request/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/projectsesame/sesame)](https://goreportcard.com/report/github.com/projectsesame/sesame) ![GitHub release](https://img.shields.io/github/release/projectsesame/sesame.svg) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![Slack](https://img.shields.io/badge/slack-join%20chat-e01563.svg?logo=slack)](https://kubernetes.slack.com/messages/sesame) [![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/4141/badge)](https://bestpractices.coreinfrastructure.org/projects/4141)



## Overview

Sesame is an Ingress controller for Kubernetes that works by deploying the [Envoy proxy](https://www.envoyproxy.io/) as a reverse proxy and load balancer.
Sesame supports dynamic configuration updates out of the box while maintaining a lightweight profile.

Sesame also introduces a new ingress API ([HTTPProxy](https://projectsesame.io/docs/main/config/fundamentals/)) which is implemented via a Custom Resource Definition (CRD).
Its goal is to expand upon the functionality of the Ingress API to allow for a richer user experience as well as solve shortcomings in the original design.

## Prerequisites

See the [compatibility matrix](https://projectsesame.io/resources/compatibility-matrix/) for the Kubernetes versions Sesame is supported with.

RBAC must be enabled on your cluster.

## Get started

Getting started with Sesame is as simple as one command.
See the [Getting Started](https://projectsesame.io/getting-started) document.

## Troubleshooting

If you encounter issues, review the Troubleshooting section of [the docs](https://projectsesame.io/docs), [file an issue](https://github.com/projectsesame/sesame/issue), or talk to us on the [#sesame channel](https://kubernetes.slack.com/messages/sesame) on the Kubernetes Slack server.

## Contributing

Thanks for taking the time to join our community and start contributing!

- Please familiarize yourself with the [Code of Conduct](/CODE_OF_CONDUCT.md) before contributing.
- See [CONTRIBUTING.md](/CONTRIBUTING.md) for information about setting up your environment, the workflow that we expect, and instructions on the developer certificate of origin that we require.
- Check out the [open issues](https://github.com/projectsesame/sesame/issues).
- Join our Kubernetes Slack channel: [#sesame](https://kubernetes.slack.com/messages/sesame/)
- Join the **Sesame Community Meetings** - [schedule, notes, and recordings can be found here](https://projectsesame.io/community)

## Roadmap

See [Sesame's roadmap](https://github.com/projectsesame/community/blob/main/ROADMAP.md) to learn more about where we are headed.

## Security

### Security Audit

A third party security audit was performed by Cure53 in December of 2020. You can see the full report [here](Sesame_Security_Audit_Dec2020.pdf).

### Reporting security vulnerabilities

If you've found a security related issue, a vulnerability, or a potential vulnerability in Sesame please let the [Sesame Security Team](mailto:cncf-sesame-maintainers@lists.cncf.io) know with the details of the vulnerability. We'll send a confirmation email to acknowledge your report, and we'll send an additional email when we've identified the issue positively or negatively.

For further details please see our [security policy](SECURITY.md).

## Changelog

See [the list of releases](https://github.com/projectsesame/sesame/releases) to find out about feature changes.
