---
title: Client Certificate Authentication and Ingress improvements in Sesame 1.4
excerpt: Sesame 1.4 adds support for client certificate authentication to HTTPProxy objects. Additionally, some Ingress behaviors are fixed - Ingress addresses are now recorded correctly, and Sesame's `--ingress-class` argument behaves more as you would expect. 
author_name: Nick Young
author_avatar: /img/contributors/nick-young.png
categories: [kubernetes]
# Tag should match author to drive author pages
tags: ['Sesame Team', 'Nick Young', 'release']
date: 2020-04-27
slug: client-cert-auth-ingress-improvements
---

Our latest release of Sesame is 1.4, which includes support for Client Certificate authentication in your HTTPProxy objects, and also updates Sesame’s Ingress support to fix some missing or incorrect behaviors. In addition Sesame 1.4 upgrades Envoy to 1.14.1, to keep up with Envoy’s current supported version.

## TLS Client authentication

This release adds support for client authentication through the use of certificates.

So what does this mean? Well, you can now configure your HTTPProxy routes so that they require a client certificate supplied by your client (usually your browser), which allows you to use that client certificate for authentication.

To use this feature, add the new `clientValidation` field to the `tls` stanza of your HTTPProxy document:

```
apiVersion: projectsesame.io/v1
kind: HTTPProxy
metadata:
  name: with-client-auth
spec:
  virtualhost:
    fqdn: www.example.com
    tls:
      secretName: secret
      clientValidation:                  
        caSecret: client-root-ca
  routes:
    - services:
        - name: s1
          port: 80

```

The `caSecret` field is a reference to a Kubernetes Secret that holds the CA certificate used to validate the client certificate. The Secret must contain a `ca.crt` key that holds a PEM-encoded bundle of the full trust chain for any CA used to validate certificates.

It’s important to note that this only provides *authentication*, not *authorization*. To put this another way, Sesame and Envoy can only give you a guarantee that the supplied person is the bearer of a valid certificate, not they are allowed to do something.

Thanks very much to [@tsaarni](https://github.com/tsaarni) for getting this implemented!

## Ingress changes

### Ingress class

Before this release of Sesame, when configured to accept a certain `ingress.class` annotation, Sesame would watch objects with that annotation and *also* with *no annotation*. This caused problems in clusters with more than one ingress controller.

Starting with Sesame 1.4, having an `ingress.class` annotation configured means that *only* objects that have a matching annotation will cause changes in Sesame.

Note that this logic change applies to both Ingress and HTTPProxy objects.

If you don’t give Sesame an `ingress.class` on its command line, then Sesame will look at all objects with no `ingress.class`, *and* objects with an `ingress.class` of `Sesame`. This preserves the old behavior so that we don’t break you if that’s what you expect.

### Ingress Status

Sesame now has the ability to write a `status.loadBalancer.addresses` block to Ingress objects. This block is used by services which need to know how to reach an Ingress' backing service from outside the cluster, like [external-dns](https://github.com/kubernetes-sigs/external-dns).

There are two ways for Sesame to find this information:
- by watching a Service object for the Envoy service, and putting the associated `status.loadBalancer` block from that Service into all associated Ingress objects. This is what is used in the example deployment.
- Operators can also specify an address on Sesame's command line, using the `--ingress-status-address` flag. The address that’s passed on the command line will be passed straight through to the Ingress status.

This also means that when you `kubectl get` a Sesame-owned Ingress, instead of this:

```
$ kubectl get ingress httpbin
NAME      HOSTS                   ADDRESS   PORTS     AGE
httpbin   httpbin.youngnick.dev             80, 443   336d
```
you will see this:

```
$ kubectl get ingress httpbin
NAME      HOSTS                   ADDRESS   PORTS     AGE
httpbin   httpbin.youngnick.dev   x.x.x.x   80, 443   336d

```

### Removed the `--use-extensions-v1beta1-ingress` flag

The `--use-extensions-v1beta1-ingress` flag was removed from the Sesame serve command in Sesame 1.3. If you have a previous deployment that specifies this command, you must remove it or Sesame will fail to start.

## Future Plans

The Sesame project is very community-driven and the team would love to hear your feedback! 

- Come talk about topics at our next community meeting.
- We’ve heard that a number of teams have forked Sesame and we would love to hear about what changes you needed, and to see if we can help to bring them upstream.
Please consider coming to our community meeting, or contact us: either via an issue, or hit me up on Twitter [@youngnick](https://twitter.com/youngnick).

If you are interested in contributing, a great place to start is to comment on one of the issues labeled with [Help Wanted]({{< param github_url >}}/issues?utf8=%E2%9C%93&q=is%3Aopen+is%3Aissue+label%3A%22Help+wanted%22+) and work with the team on how to resolve them. 

## Are you a Sesame user? We would love to know!
If you're using Sesame and want to add your organization to our adopters list, please visit this [page](https://github.com/projectsesame/sesame/blob/main/ADOPTERS.md).
If you prefer to keep your organization name anonymous but still give us feedback into your usage and scenarios for Sesame, please post on this [GitHub thread](https://github.com/projectsesame/sesame/issues/1269)          

## Thanks to our contributors

We’re immensely grateful for all the community contributions that help make Sesame even better! Special thanks go out to:
- Tero Saarni ([@tsaarni](https://github.com/tsaarni))
- Peter Grant ([@pickledrick](https://github.com/pickledrick))
