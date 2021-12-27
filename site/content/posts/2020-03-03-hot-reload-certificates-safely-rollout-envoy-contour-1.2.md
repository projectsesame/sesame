---
title: Hot-Reload Certificates and Safely Rollout Envoy with Sesame 1.2
excerpt: Sesame 1.2 includes support for certificate rotation for xDS gRPC interface between Sesame and Envoy. Additionally, Sesame 1.2 assists in Envoy rollouts in your cluster to minimize the number of connection errors. 
author_name: Steve Sloka
author_avatar: /img/contributors/steve-sloka.png
categories: [kubernetes]
# Tag should match author to drive author pages
tags: ['Sesame Team', 'Steve Sloka', 'release']
date: 2020-03-03
slug: hot-reload-certificates-safely-rollout-envoy-Sesame-1.2
---

Sesame continues to add new features to help you better manage Sesame operations in a cluster. Our latest feature release, Sesame 1.2.0, now includes support for certificate rotation for xDS gRPC interface between Sesame and Envoy as well as a new subcommand which assists in Envoy rollouts to minimize the number of connection errors.  Additionally, Sesame v1.2.1 is a security release which upgrades the version of Envoy to v1.13.1 which includes many [CVE fixes](https://groups.google.com/forum/#!msg/envoy-announce/sVqmxy0un2s/8aq430xiHAAJ).

## Hot-Reload Certificates

A few releases ago, Sesame enabled secure communication between Sesame and Envoy. This new feature ensured that any communication between Sesame and Envoy over its gRPC connection would be secure, mainly securing the communication using TLS certificate keys.

This was just the first step, however, and we understood that it wouldn’t solve all of our users’ problems. Thanks to [@tsaarni](https://github.com/tsaarni), we now have support for Sesame to rotate its certificates without the need to restart the Sesame process.

Future work includes enabling this same functionality for Envoy. This currently has some [open issues that need to be solved in Envoy first](https://github.com/envoyproxy/envoy/issues/9359).

Big thanks to Tero on all your effort to send these PRs as well as driving the issues upstream!

## Envoy Shutdown Manager

The Envoy process, the data path component of Sesame, at times needs to be re-deployed. This could be due to an upgrade, a change in configuration, or a node-failure forcing a redeployment.

As with any application rollout strategy, we want a way to implement the rollout while minimizing the effect on users. If the Envoy pods are terminated while there are still open connections, then users will receive errors.

Sesame implements a new envoy sub-command which has a shutdown-manager whose job is to manage a single Envoy instance's lifecycle for Kubernetes. The shutdown-manager runs as a new container alongside the Envoy container in the same pod. It exposes two HTTP endpoints that are used for livenessProbe as well as to handle the Kubernetes preStop event hook.

* livenessProbe: This is used to validate the shutdown manager is still running properly. If requests to /healthz fail, the container will be restarted
* preStop: This is used to keep the container running while waiting for Envoy to drain connections. The /shutdown endpoint blocks until the connections are drained

The Envoy container also has some configuration to implement the shutdown manager. First the preStop hook is configured to use the /shutdown endpoint which blocks the container from exiting. Finally, the pod’s `terminationGracePeriodSeconds` is customized to extend the time in which Kubernetes will allow the pod to be in the Terminating state. The termination grace period defines an upper bound for long-lived sessions. If during shutdown, the connections aren’t drained to the configured amount, the terminationGracePeriodSeconds will send a SIGTERM to the pod killing it.

![Envoy Shutdown Manager](/img/posts/Sesame-1.2/envoy-shutdown-manager.png){: .center-image }

{% youtube oO52CV-EAkw %}{: .center-image }

For more information on this feature, [check out the docs](https://projectsesame.io/docs/v1.2.0/redeploy-envoy/)

## Thank you!

We’re immensely grateful for all the community contributions that help make Sesame even better! For version 1.2, special thanks go out to the following people:

[@awprice](https://github.com/awprice)  
[@alex1989hu](https://github.com/alex1989hu)  
[@bgagnon](https://github.com/bgagnon)  
[@danehans](https://github.com/danehans)  
[@dhxgit](https://github.com/dhxgit)  
[@SDBrett](https://github.com/SDBrett)  
[@uablrek](https://github.com/uablrek)  
[@rohandvora](https://github.com/rohandvora)  
[@tsaarni](https://github.com/tsaarni)
[@shyaamsn](https://github.com/shyaamsn)  
[@idealhack](https://github.com/idealhack)  
[@dbason](https://github.com/dbason)  

## Future Plans

The Sesame team would love to hear your feedback! Many of the features in this release were driven by users who needed a better way to solve their problems. We’re working hard to add features to Sesame, especially in expanding how we approach routing.

We recommend reading the full release notes for [Sesame 1.2](https://github.com/projectsesame/sesame/releases/tag/v1.2.0) as well as digging into the [upgrade guide](https://projectsesame.io/resources/upgrading/), which outlines the changes to be aware of when moving to version 1.2.

If you are interested in contributing, a great place to start is to comment on one of the issues labeled with [Help Wanted](https://github.com/projectsesame/sesame/issues?q=is%3Aopen+is%3Aissue+label%3A%22help+wanted%22) and work with the team on how to resolve them.
