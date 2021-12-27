---
title: Upgrading Sesame
layout: page
---

<!-- NOTE: this document should be formatted with one sentence per line to made reviewing easier. -->

This document describes the changes needed to upgrade your Sesame installation.

<div id="toc" class="navigation"></div>

## Upgrading Sesame 1.19.0 to 1.19.1

Sesame 1.19.1 is the current stable release.

### Required Envoy version

All users should ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.19.1`.

Please see the [Envoy Release Notes][34] for information about issues fixed in Envoy 1.19.1.

### The easy way to upgrade

If the following are true for you:

* Your installation is in the `projectsesame` namespace.
* You are using our [quickstart example][18] deployments.
* Your cluster can take a few minutes of downtime.

Then the simplest way to upgrade to 1.19.1 is to delete the `projectsesame` namespace and reapply one of the example configurations:

```bash
$ kubectl delete namespace projectsesame
$ kubectl apply -f {{< param base_url >}}/quickstart/v1.19.1/sesame.yaml
```

This will remove the Envoy and Sesame pods from your cluster and recreate them with the updated configuration.
If you're using a `LoadBalancer` Service, (which most of the examples do) deleting and recreating may change the public IP assigned by your cloud provider.
You'll need to re-check where your DNS names are pointing as well, using [Get your hostname or IP address][12].

### The less easy way

This section contains information for administrators who wish to apply the Sesame 1.19.0 to 1.19.1 changes manually.
The YAML files referenced in this section can be found by cloning the Sesame repository and checking out the `v1.19.1` tag.

If your version of Sesame is older than v1.19.0, please upgrade to v1.19.0 first, then upgrade to v1.19.1.

1. Update the Sesame CRDs:

    ```bash
    $ kubectl apply -f examples/sesame/01-crds.yaml
    ```

1. Users of the example deployment should reapply the certgen Job YAML which will re-generate the relevant Secrets in a format compatible with [cert-manager](https://cert-manager.io) TLS secrets.
   This will rotate the TLS certificates used for gRPC security.

    ```bash
    $ kubectl apply -f examples/sesame/02-job-certgen.yaml
    ```

1. Update the Sesame cluster role:

    ```bash
    $ kubectl apply -f examples/sesame/02-role-sesame.yaml
    ```

1. Upgrade the Sesame deployment:

    ```bash
    $ kubectl apply -f examples/sesame/03-sesame.yaml
    ```

1. Once the Sesame deployment has finished upgrading, update the Envoy DaemonSet:

    ```bash
    $ kubectl apply -f examples/sesame/03-envoy.yaml
    ```

## Upgrading Sesame 1.18.3 to 1.19.0

### Required Envoy version

All users should ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.19.1`.

Please see the [Envoy Release Notes][34] for information about issues fixed in Envoy 1.19.1.

### The easy way to upgrade

If the following are true for you:

* Your installation is in the `projectsesame` namespace.
* You are using our [quickstart example][18] deployments.
* Your cluster can take a few minutes of downtime.

Then the simplest way to upgrade to 1.19.0 is to delete the `projectsesame` namespace and reapply one of the example configurations:

```bash
$ kubectl delete namespace projectsesame
$ kubectl apply -f {{< param base_url >}}/quickstart/v1.19.0/sesame.yaml
```

This will remove the Envoy and Sesame pods from your cluster and recreate them with the updated configuration.
If you're using a `LoadBalancer` Service, (which most of the examples do) deleting and recreating may change the public IP assigned by your cloud provider.
You'll need to re-check where your DNS names are pointing as well, using [Get your hostname or IP address][12].

### The less easy way

This section contains information for administrators who wish to apply the Sesame 1.18.3 to 1.19.0 changes manually.
The YAML files referenced in this section can be found by cloning the Sesame repository and checking out the `v1.19.0` tag.

If your version of Sesame is older than v1.18.3, please upgrade to v1.18.3 first, then upgrade to v1.19.0.

1. Update the Sesame CRDs:

    ```bash
    $ kubectl apply -f examples/sesame/01-crds.yaml
    ```

1. Users of the example deployment should reapply the certgen Job YAML which will re-generate the relevant Secrets in a format compatible with [cert-manager](https://cert-manager.io) TLS secrets.
   This will rotate the TLS certificates used for gRPC security.

    ```bash
    $ kubectl apply -f examples/sesame/02-job-certgen.yaml
    ```

1. Update the Sesame cluster role:

    ```bash
    $ kubectl apply -f examples/sesame/02-role-sesame.yaml
    ```

1. Upgrade the Sesame deployment:

    ```bash
    $ kubectl apply -f examples/sesame/03-sesame.yaml
    ```

1. Once the Sesame deployment has finished upgrading, update the Envoy DaemonSet:

    ```bash
    $ kubectl apply -f examples/sesame/03-envoy.yaml
    ```

## Upgrading Sesame 1.18.2 to 1.18.3

### Required Envoy version

All users should ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.19.1`.

Please see the [Envoy Release Notes][34] for information about issues fixed in Envoy 1.19.1.

### The easy way to upgrade

If the following are true for you:

* Your installation is in the `projectsesame` namespace.
* You are using our [quickstart example][18] deployments.
* Your cluster can take a few minutes of downtime.

Then the simplest way to upgrade to 1.18.3 is to delete the `projectsesame` namespace and reapply one of the example configurations:

```bash
$ kubectl delete namespace projectsesame
$ kubectl apply -f {{< param base_url >}}/quickstart/v1.18.3/sesame.yaml
```

This will remove the Envoy and Sesame pods from your cluster and recreate them with the updated configuration.
If you're using a `LoadBalancer` Service, (which most of the examples do) deleting and recreating may change the public IP assigned by your cloud provider.
You'll need to re-check where your DNS names are pointing as well, using [Get your hostname or IP address][12].

### The less easy way

This section contains information for administrators who wish to apply the Sesame 1.18.2 to 1.18.3 changes manually.
The YAML files referenced in this section can be found by cloning the Sesame repository and checking out the `v1.18.3` tag.

If your version of Sesame is older than v1.18.2, please upgrade to v1.18.2 first, then upgrade to v1.18.3.

1. Users of the example deployment should reapply the certgen Job YAML which will re-generate the relevant Secrets in a format compatible with [cert-manager](https://cert-manager.io) TLS secrets.
   This will rotate the TLS certificates used for gRPC security.

    ```bash
    $ kubectl apply -f examples/sesame/02-job-certgen.yaml
    ```

1. Update your RBAC definitions:

    ```bash
    $ kubectl apply -f examples/sesame/02-rbac.yaml
    ```

1. Upgrade your Sesame deployment:

    ```bash
    $ kubectl apply -f examples/sesame/03-sesame.yaml
    ```

1. Once the Sesame deployment has finished upgrading, update the Envoy DaemonSet:

    ```bash
    $ kubectl apply -f examples/sesame/03-envoy.yaml
    ```


## Upgrading Sesame 1.18.1 to 1.18.2

### Required Envoy version

All users should ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.19.1`.

Please see the [Envoy Release Notes][34] for information about issues fixed in Envoy 1.19.1.

### The easy way to upgrade

If the following are true for you:

* Your installation is in the `projectsesame` namespace.
* You are using our [quickstart example][18] deployments.
* Your cluster can take a few minutes of downtime.

Then the simplest way to upgrade to 1.18.2 is to delete the `projectsesame` namespace and reapply one of the example configurations:

```bash
$ kubectl delete namespace projectsesame
$ kubectl apply -f {{< param base_url >}}/quickstart/v1.18.2/sesame.yaml
```

This will remove the Envoy and Sesame pods from your cluster and recreate them with the updated configuration.
If you're using a `LoadBalancer` Service, (which most of the examples do) deleting and recreating may change the public IP assigned by your cloud provider.
You'll need to re-check where your DNS names are pointing as well, using [Get your hostname or IP address][12].

### The less easy way

This section contains information for administrators who wish to apply the Sesame 1.18.1 to 1.18.2 changes manually.
The YAML files referenced in this section can be found by cloning the Sesame repository and checking out the `v1.18.2` tag.

If your version of Sesame is older than v1.18.1, please upgrade to v1.18.1 first, then upgrade to v1.18.2.

1. Users of the example deployment should reapply the certgen Job YAML which will re-generate the relevant Secrets in a format compatible with [cert-manager](https://cert-manager.io) TLS secrets.
   This will rotate the TLS certificates used for gRPC security.

    ```bash
    $ kubectl apply -f examples/sesame/02-job-certgen.yaml
    ```

1. Update your RBAC definitions:

    ```bash
    $ kubectl apply -f examples/sesame/02-rbac.yaml
    ```

1. Upgrade your Sesame deployment:

    ```bash
    $ kubectl apply -f examples/sesame/03-sesame.yaml
    ```

1. Once the Sesame deployment has finished upgrading, update the Envoy DaemonSet:

    ```bash
    $ kubectl apply -f examples/sesame/03-envoy.yaml
    ```

## Upgrading Sesame 1.18.0 to 1.18.1

### Required Envoy version

All users should ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.19.1`.

Please see the [Envoy Release Notes][34] for information about issues fixed in Envoy 1.19.1.

### The easy way to upgrade

If the following are true for you:

* Your installation is in the `projectsesame` namespace.
* You are using our [quickstart example][18] deployments.
* Your cluster can take a few minutes of downtime.

Then the simplest way to upgrade to 1.18.1 is to delete the `projectsesame` namespace and reapply one of the example configurations:

```bash
$ kubectl delete namespace projectsesame
$ kubectl apply -f {{< param base_url >}}/quickstart/v1.18.1/sesame.yaml
```

This will remove the Envoy and Sesame pods from your cluster and recreate them with the updated configuration.
If you're using a `LoadBalancer` Service, (which most of the examples do) deleting and recreating may change the public IP assigned by your cloud provider.
You'll need to re-check where your DNS names are pointing as well, using [Get your hostname or IP address][12].

### The less easy way

This section contains information for administrators who wish to apply the Sesame 1.18.0 to 1.18.1 changes manually.
The YAML files referenced in this section can be found by cloning the Sesame repository and checking out the `v1.18.1` tag.

If your version of Sesame is older than v1.18.0, please upgrade to v1.18.0 first, then upgrade to v1.18.1.

1. Users of the example deployment should reapply the certgen Job YAML which will re-generate the relevant Secrets in a format compatible with [cert-manager](https://cert-manager.io) TLS secrets.
   This will rotate the TLS certificates used for gRPC security.

    ```bash
    $ kubectl apply -f examples/sesame/02-job-certgen.yaml
    ```

1. Update your RBAC definitions:

    ```bash
    $ kubectl apply -f examples/sesame/02-rbac.yaml
    ```

1. Upgrade your Sesame deployment:

    ```bash
    $ kubectl apply -f examples/sesame/03-sesame.yaml
    ```

1. Once the Sesame deployment has finished upgrading, update the Envoy DaemonSet:

    ```bash
    $ kubectl apply -f examples/sesame/03-envoy.yaml
    ```


## Upgrading Sesame 1.17.1 to 1.18.0

**If you utilize ExternalName services in your cluster, please note that this release disables Sesame processing such services by default.**
**Please see [this CVE](https://github.com/projectsesame/sesame/security/advisories/GHSA-5ph6-qq5x-7jwc) for context and the [1.18.0 release notes](https://github.com/projectsesame/sesame/releases/tag/v1.18.0).**

### Required Envoy version

All users should ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.19.0`.

Please see the [Envoy Release Notes][33] for information about issues fixed in Envoy 1.19.0.

### The easy way to upgrade

If the following are true for you:

* Your installation is in the `projectsesame` namespace.
* You are using our [quickstart example][18] deployments.
* Your cluster can take a few minutes of downtime.

Then the simplest way to upgrade to 1.18.0 is to delete the `projectsesame` namespace and reapply one of the example configurations:

```bash
$ kubectl delete namespace projectsesame
$ kubectl apply -f {{< param base_url >}}/quickstart/v1.18.0/sesame.yaml
```

This will remove the Envoy and Sesame pods from your cluster and recreate them with the updated configuration.
If you're using a `LoadBalancer` Service, (which most of the examples do) deleting and recreating may change the public IP assigned by your cloud provider.
You'll need to re-check where your DNS names are pointing as well, using [Get your hostname or IP address][12].

### The less easy way

This section contains information for administrators who wish to apply the Sesame 1.17.1 to 1.18.0 changes manually.
The YAML files referenced in this section can be found by cloning the Sesame repository and checking out the `v1.18.0` tag.

If your version of Sesame is older than v1.17.1, please upgrade to v1.17.1 first, then upgrade to v1.18.0.

1. The Sesame CRD definitions must be re-applied to the cluster, since a number of compatible changes and additions have been made to the Sesame API:

    ```bash
    $ kubectl apply -f examples/sesame/01-crds.yaml
    ```

1. Users of the example deployment should reapply the certgen Job YAML which will re-generate the relevant Secrets in a format compatible with [cert-manager](https://cert-manager.io) TLS secrets.
   This will rotate the TLS certificates used for gRPC security.

    ```bash
    $ kubectl apply -f examples/sesame/02-job-certgen.yaml
    ```

1. Update your RBAC definitions:

    ```bash
    $ kubectl apply -f examples/sesame/02-rbac.yaml
    ```

1. Upgrade your Sesame deployment:

    ```bash
    $ kubectl apply -f examples/sesame/03-sesame.yaml
    ```

1. Once the Sesame deployment has finished upgrading, update the Envoy DaemonSet:

    ```bash
    $ kubectl apply -f examples/sesame/03-envoy.yaml
    ```

## Upgrading Sesame 1.17.0 to 1.17.1

### Required Envoy version

All users should ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.18.3`.

Please see the [Envoy Release Notes][32] for information about issues fixed in Envoy 1.18.3.

### The easy way to upgrade

If the following are true for you:

* Your installation is in the `projectsesame` namespace.
* You are using our [quickstart example][18] deployments.
* Your cluster can take a few minutes of downtime.

Then the simplest way to upgrade to 1.17.1 is to delete the `projectsesame` namespace and reapply one of the example configurations:

```bash
$ kubectl delete namespace projectsesame
$ kubectl apply -f {{< param base_url >}}/quickstart/v1.17.1/sesame.yaml
```

This will remove the Envoy and Sesame pods from your cluster and recreate them with the updated configuration.
If you're using a `LoadBalancer` Service, (which most of the examples do) deleting and recreating may change the public IP assigned by your cloud provider.
You'll need to re-check where your DNS names are pointing as well, using [Get your hostname or IP address][12].

### The less easy way

This section contains information for administrators who wish to apply the Sesame 1.17.0 to 1.17.1 changes manually.
The YAML files referenced in this section can be found by cloning the Sesame repository and checking out the `v1.17.1` tag.

If your version of Sesame is older than v1.17.0, please upgrade to v1.17.0 first, then upgrade to v1.17.1.

1. Users of the example deployment should reapply the certgen Job YAML which will re-generate the relevant Secrets in a format compatible with [cert-manager](https://cert-manager.io) TLS secrets.
   This will rotate the TLS certificates used for gRPC security.

    ```bash
    $ kubectl apply -f examples/sesame/02-job-certgen.yaml
    ```

1. Update your RBAC definitions:

    ```bash
    $ kubectl apply -f examples/sesame/02-rbac.yaml
    ```

1. Upgrade your Sesame deployment:

    ```bash
    $ kubectl apply -f examples/sesame/03-sesame.yaml
    ```

1. Once the Sesame deployment has finished upgrading, update the Envoy DaemonSet:

    ```bash
    $ kubectl apply -f examples/sesame/03-envoy.yaml
    ```

## Upgrading Sesame 1.16.0 to 1.17.0

### Required Envoy version

All users should ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.18.3`.

Please see the [Envoy Release Notes][32] for information about issues fixed in Envoy 1.18.3.

### The easiest way to upgrade (alpha)
For existing Sesame Operator users, complete the following steps to upgrade Sesame:

- Verify the operator is running v1.16.0, and it's deployment status is "Available=True".
- Verify the status of all Sesame custom resources are "Available=True".
- Update the operator's image to v1.17.0:
   ```bash
   $ kubectl patch deploy/sesame-operator -n sesame-operator -p '{"spec":{"template":{"spec":{"containers":[{"name":"sesame-operator","image":"docker.io/projectsesame/sesame-operator:v1.17.0"}]}}}}'
   ```
- The above command will upgrade the operator. After the operator runs the new version, it will upgrade Sesame.
- Verify the operator and Sesame are running the new version.
- Verify all Sesame custom resources are "Available=True".

### The easy way to upgrade

If the following are true for you:

* Your installation is in the `projectsesame` namespace.
* You are using our [quickstart example][18] deployments.
* Your cluster can take a few minutes of downtime.

Then the simplest way to upgrade to 1.17.0 is to delete the `projectsesame` namespace and reapply one of the example configurations:

```bash
$ kubectl delete namespace projectsesame
$ kubectl apply -f {{< param base_url >}}/quickstart/v1.17.0/sesame.yaml
```

This will remove the Envoy and Sesame pods from your cluster and recreate them with the updated configuration.
If you're using a `LoadBalancer` Service, (which most of the examples do) deleting and recreating may change the public IP assigned by your cloud provider.
You'll need to re-check where your DNS names are pointing as well, using [Get your hostname or IP address][12].

### The less easy way

This section contains information for administrators who wish to apply the Sesame 1.16.0 to 1.17.0 changes manually.
The YAML files referenced in this section can be found by cloning the Sesame repository and checking out the `v1.17.0` tag.

If your version of Sesame is older than v1.16.0, please upgrade to v1.16.0 first, then upgrade to v1.17.0.

1. Users of the example deployment should reapply the certgen Job YAML which will re-generate the relevant Secrets in a format compatible with [cert-manager](https://cert-manager.io) TLS secrets.
   This will rotate the TLS certificates used for gRPC security.

    ```bash
    $ kubectl apply -f examples/sesame/02-job-certgen.yaml
    ```

1. Update your RBAC definitions:

    ```bash
    $ kubectl apply -f examples/sesame/02-rbac.yaml
    ```

1. Upgrade your Sesame deployment:

    ```bash
    $ kubectl apply -f examples/sesame/03-sesame.yaml
    ```

1. Once the Sesame deployment has finished upgrading, update the Envoy DaemonSet:

    ```bash
    $ kubectl apply -f examples/sesame/03-envoy.yaml
    ```


## Upgrading Sesame 1.15.1 to 1.16.0

### Required Envoy version

All users should ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.18.3`.

Please see the [Envoy Release Notes][32] for information about issues fixed in Envoy 1.18.3.

### The easiest way to upgrade (alpha)
For existing Sesame Operator users, complete the following steps to upgrade Sesame:

- Verify the operator is running v1.15.1, and it's deployment status is "Available=True".
- Verify the status of all Sesame custom resources are "Available=True".
- Update the operator's image to v1.16.0:
   ```bash
   $ kubectl patch deploy/sesame-operator -n sesame-operator -p '{"spec":{"template":{"spec":{"containers":[{"name":"sesame-operator","image":"docker.io/projectsesame/sesame-operator:v1.16.0"}]}}}}'
   ```
- The above command will upgrade the operator. After the operator runs the new version, it will upgrade Sesame.
- Verify the operator and Sesame are running the new version.
- Verify all Sesame custom resources are "Available=True".

### The easy way to upgrade

If the following are true for you:

* Your installation is in the `projectsesame` namespace.
* You are using our [quickstart example][18] deployments.
* Your cluster can take a few minutes of downtime.

Then the simplest way to upgrade to 1.16.0 is to delete the `projectsesame` namespace and reapply one of the example configurations:

```bash
$ kubectl delete namespace projectsesame
$ kubectl apply -f {{< param base_url >}}/quickstart/v1.16.0/sesame.yaml
```

This will remove the Envoy and Sesame pods from your cluster and recreate them with the updated configuration.
If you're using a `LoadBalancer` Service, (which most of the examples do) deleting and recreating may change the public IP assigned by your cloud provider.
You'll need to re-check where your DNS names are pointing as well, using [Get your hostname or IP address][12].

### The less easy way

This section contains information for administrators who wish to apply the Sesame 1.15.1 to 1.16.0 changes manually.
The YAML files referenced in this section can be found by cloning the Sesame repository and checking out the `v1.16.0` tag.

If your version of Sesame is older than v1.15.1, please upgrade to v1.15.1 first, then upgrade to v1.16.0.

1. Users of the example deployment should reapply the certgen Job YAML which will re-generate the relevant Secrets in the new format, which is compatible with [cert-manager](https://cert-manager.io) TLS secrets.
   This will rotate the TLS certificates used for gRPC security.

    ```bash
    $ kubectl apply -f examples/sesame/02-job-certgen.yaml
    ```

1. Upgrade your Sesame deployment:

    ```bash
    $ kubectl apply -f examples/sesame/03-sesame.yaml
    ```

1. Once the Sesame deployment has finished upgrading, update the Envoy DaemonSet:

    ```bash
    $ kubectl apply -f examples/sesame/03-envoy.yaml
    ```

## Upgrading Sesame 1.15.0 to 1.15.1

### Required Envoy version

All users should ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.18.3`.

Please see the [Envoy Release Notes][32] for information about issues fixed in Envoy 1.18.3.

### The easy way to upgrade

If the following are true for you:

* Your installation is in the `projectsesame` namespace.
* You are using our [quickstart example][18] deployments.
* Your cluster can take a few minutes of downtime.

Then the simplest way to upgrade to 1.15.1 is to delete the `projectsesame` namespace and reapply one of the example configurations:

```bash
$ kubectl delete namespace projectsesame
$ kubectl apply -f {{< param base_url >}}/quickstart/v1.15.1/sesame.yaml
```

This will remove the Envoy and Sesame pods from your cluster and recreate them with the updated configuration.
If you're using a `LoadBalancer` Service, (which most of the examples do) deleting and recreating may change the public IP assigned by your cloud provider.
You'll need to re-check where your DNS names are pointing as well, using [Get your hostname or IP address][12].

### The less easy way

This section contains information for administrators who wish to apply the Sesame 1.15.0 to 1.15.01 changes manually.
The YAML files referenced in this section can be found by cloning the Sesame repository and checking out the `v1.15.1` tag.

If your version of Sesame is older than v1.15.0, please upgrade to v1.15.0 first, then upgrade to v1.15.1.

1. Users of the example deployment should reapply the certgen Job YAML which will re-generate the relevant Secrets in the new format, which is compatible with [cert-manager](https://cert-manager.io) TLS secrets.
This will rotate the TLS certificates used for gRPC security.

    ```bash
    $ kubectl apply -f examples/sesame/02-job-certgen.yaml
    ```

1. Upgrade your Sesame deployment:

    ```bash
    $ kubectl apply -f examples/sesame/03-sesame.yaml
    ```

1. Once the Sesame deployment has finished upgrading, update the Envoy DaemonSet:

    ```bash
    $ kubectl apply -f examples/sesame/03-envoy.yaml
    ```

## Upgrading Sesame 1.14.1 to 1.15.0

### Required Envoy version

All users should ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.18.2`.

Please see the [Envoy Release Notes][31] for information about issues fixed in Envoy 1.18.2.

### The easy way to upgrade

If the following are true for you:

* Your installation is in the `projectsesame` namespace.
* You are using our [quickstart example][18] deployments.
* Your cluster can take a few minutes of downtime.

Then the simplest way to upgrade to 1.15.0 is to delete the `projectsesame` namespace and reapply one of the example configurations:

```bash
$ kubectl delete namespace projectsesame
$ kubectl apply -f {{< param base_url >}}/quickstart/v1.15.0/sesame.yaml
```

This will remove the Envoy and Sesame pods from your cluster and recreate them with the updated configuration.
If you're using a `LoadBalancer` Service, (which most of the examples do) deleting and recreating may change the public IP assigned by your cloud provider.
You'll need to re-check where your DNS names are pointing as well, using [Get your hostname or IP address][12].

### The less easy way

This section contains information for administrators who wish to apply the Sesame 1.14.1 to 1.15.0 changes manually.
The YAML files referenced in this section can be found by cloning the Sesame repository and checking out the `v1.15.0` tag.

If your version of Sesame is older than v1.14, please upgrade to v1.14 first, then upgrade to v1.15.0.

1. The Sesame CRD definitions must be re-applied to the cluster, since a number of compatible changes and additions have been made to the Sesame API:

    ```bash
    $ kubectl apply -f examples/sesame/01-crds.yaml
    ```

1. Users of the example deployment should reapply the certgen Job YAML which will re-generate the relevant Secrets in the new format, which is compatible with [cert-manager](https://cert-manager.io) TLS secrets.
This will rotate the TLS certificates used for gRPC security.

    ```bash
    $ kubectl apply -f examples/sesame/02-job-certgen.yaml
    ```

1. This release includes an update to RBAC rules. Update the Sesame ClusterRole with the following:

    ```bash
    $ kubectl apply -f examples/sesame/02-role-sesame.yaml
    ```

1. This release includes changes to support Ingress wildcard hosts that require Envoy to be upgraded *before* Sesame. Update the Envoy DaemonSet:

    ```bash
    $ kubectl apply -f examples/sesame/03-envoy.yaml
    ```

1. Once the Envoy DaemonSet has finished updating, upgrade your Sesame deployment:

    ```bash
    $ kubectl apply -f examples/sesame/03-sesame.yaml
    ```

## Upgrading Sesame 1.14.0 to 1.14.1

### Required Envoy version

All users should ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.17.2`.

Please see the [Envoy Release Notes][30] for information about issues fixed in Envoy 1.17.2.

### The easy way to upgrade

If the following are true for you:

* Your installation is in the `projectsesame` namespace.
* You are using our [quickstart example][18] deployments.
* Your cluster can take a few minutes of downtime.

Then the simplest way to upgrade to 1.14.1 is to delete the `projectsesame` namespace and reapply one of the example configurations:

```bash
$ kubectl delete namespace projectsesame
$ kubectl apply -f {{< param base_url >}}/quickstart/v1.14.1/sesame.yaml
```

This will remove the Envoy and Sesame pods from your cluster and recreate them with the updated configuration.
If you're using a `LoadBalancer` Service, (which most of the examples do) deleting and recreating may change the public IP assigned by your cloud provider.
You'll need to re-check where your DNS names are pointing as well, using [Get your hostname or IP address][12].

### The less easy way

This section contains information for administrators who wish to apply the Sesame 1.14.0 to 1.14.1 changes manually.
The YAML files referenced in this section can be found by cloning the Sesame repository and checking out the `v1.14.1` tag.

Users of the example deployment should first reapply the certgen Job YAML which will re-generate the relevant Secrets in the new format, which is compatible with [cert-manager](https://cert-manager.io) TLS secrets.
This will rotate the TLS certificates used for gRPC security.

If your version of Sesame is older than v1.14, please upgrade to v1.14 first, then upgrade to v1.14.1.

```bash
$ kubectl apply -f examples/sesame/02-job-certgen.yaml
```

Upgrade your Sesame deployment:

```bash
$ kubectl apply -f examples/sesame/03-sesame.yaml
```

Once the Sesame deployment has finished upgrading, update the Envoy DaemonSet:

```bash
$ kubectl apply -f examples/sesame/03-envoy.yaml
```

## Upgrading Sesame 1.13.1 to 1.14.0

### Required Envoy version

All users should ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.17.1`.

Please see the [Envoy Release Notes][29] for information about issues fixed in Envoy 1.17.1.

### The easy way to upgrade

If the following are true for you:

* Your installation is in the `projectsesame` namespace.
* You are using our [quickstart example][18] deployments.
* Your cluster can take a few minutes of downtime.

Then the simplest way to upgrade to 1.14.0 is to delete the `projectsesame` namespace and reapply one of the example configurations:

```bash
$ kubectl delete namespace projectsesame
$ kubectl apply -f {{< param base_url >}}/quickstart/v1.14.0/sesame.yaml
```

This will remove the Envoy and Sesame pods from your cluster and recreate them with the updated configuration.
If you're using a `LoadBalancer` Service, (which most of the examples do) deleting and recreating may change the public IP assigned by your cloud provider.
You'll need to re-check where your DNS names are pointing as well, using [Get your hostname or IP address][12].

### The less easy way

This section contains information for administrators who wish to apply the Sesame 1.13.1 to 1.14.0 changes manually.
The YAML files referenced in this section can be found by cloning the Sesame repository and checking out the `v1.14.0` tag.

The Sesame CRD definition must be re-applied to the cluster, since a number of compatible changes and additions have been made to the Sesame API:

```bash
$ kubectl apply -f examples/sesame/01-crds.yaml
```

Users of the example deployment should first reapply the certgen Job YAML which will re-generate the relevant Secrets in the new format, which is compatible with [cert-manager](https://cert-manager.io) TLS secrets.
This will rotate the TLS certificates used for gRPC security.

```bash
$ kubectl apply -f examples/sesame/02-job-certgen.yaml
```

If your version of Sesame is older than v1.13, please upgrade to v1.13 first, then upgrade to v1.14.0.

This release includes an update to the Envoy service ports. Upgrade your Envoy service with the following:

```bash
$ kubectl apply -f examples/sesame/02-service-envoy.yaml
```

Upgrade your Sesame deployment:

```bash
$ kubectl apply -f examples/sesame/03-sesame.yaml
```

Once the Sesame deployment has finished upgrading, update the Envoy DaemonSet:

```bash
$ kubectl apply -f examples/sesame/03-envoy.yaml
```

## Upgrading Sesame 1.12.0 to 1.13.1

### Required Envoy version

All users should ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.17.1`.

Please see the [Envoy Release Notes][29] for information about issues fixed in Envoy 1.17.1.

### The easy way to upgrade

If the following are true for you:

* Your installation is in the `projectsesame` namespace.
* You are using our [quickstart example][18] deployments.
* Your cluster can take a few minutes of downtime.

Then the simplest way to upgrade to 1.13.1 is to delete the `projectsesame` namespace and reapply one of the example configurations:

```bash
$ kubectl delete namespace projectsesame
$ kubectl apply -f {{< param base_url >}}/quickstart/v1.13.1/sesame.yaml
```

This will remove the Envoy and Sesame pods from your cluster and recreate them with the updated configuration.
If you're using a `LoadBalancer` Service, (which most of the examples do) deleting and recreating may change the public IP assigned by your cloud provider.
You'll need to re-check where your DNS names are pointing as well, using [Get your hostname or IP address][12].

### The less easy way

This section contains information for administrators who wish to apply the Sesame 1.12.0 to 1.13.1 changes manually.
The YAML files referenced in this section can be found by cloning the Sesame repository and checking out the `v1.13.1` tag.

The Sesame CRD definition must be re-applied to the cluster, since a number of compatible changes and additions have been made to the Sesame API:

```bash
$ kubectl apply -f examples/sesame/01-crds.yaml
```

Users of the example deployment should first reapply the certgen Job YAML which will re-generate the relevant Secrets in the new format, which is compatible with [cert-manager](https://cert-manager.io) TLS secrets.
This will rotate the TLS certificates used for gRPC security.

```bash
$ kubectl apply -f examples/sesame/02-job-certgen.yaml
```

If your version of Sesame is older than v1.12, please upgrade to v1.12 first, then upgrade to v1.13.1.

Upgrade your Sesame deployment:

```bash
$ kubectl apply -f examples/sesame/03-sesame.yaml
```

Once the Sesame deployment has finished upgrading, update the Envoy DaemonSet:

```bash
$ kubectl apply -f examples/sesame/03-envoy.yaml
```

## Upgrading Sesame 1.11.0 to 1.12.0

### Required Envoy version

All users should ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.17.0`.

Please see the [Envoy Release Notes][28] for information about issues fixed in Envoy 1.17.0.

### The easy way to upgrade

If the following are true for you:

* Your installation is in the `projectsesame` namespace.
* You are using our [quickstart example][18] deployments.
* Your cluster can take a few minutes of downtime.

Then the simplest way to upgrade to 1.12.0 is to delete the `projectsesame` namespace and reapply one of the example configurations:

```bash
$ kubectl delete namespace projectsesame
$ kubectl apply -f {{< param base_url >}}/quickstart/v1.12.0/sesame.yaml
```

This will remove the Envoy and Sesame pods from your cluster and recreate them with the updated configuration.
If you're using a `LoadBalancer` Service, (which most of the examples do) deleting and recreating may change the public IP assigned by your cloud provider.
You'll need to re-check where your DNS names are pointing as well, using [Get your hostname or IP address][12].

### The less easy way

This section contains information for administrators who wish to apply the Sesame 1.11.0 to 1.12.0 changes manually.
The YAML files referenced in this section can be found by cloning the Sesame repository and checking out the `v1.12.0` tag.

The Sesame CRD definition must be re-applied to the cluster, since a number of compatible changes and additions have been made to the Sesame API:

```bash
$ kubectl apply -f examples/sesame/01-crds.yaml
```

Users of the example deployment should first reapply the certgen Job YAML which will re-generate the relevant Secrets in the new format, which is compatible with [cert-manager](https://cert-manager.io) TLS secrets.
This will rotate the TLS certificates used for gRPC security.

```bash
$ kubectl apply -f examples/sesame/02-job-certgen.yaml
```

If your version of Sesame is older than v1.11, please upgrade to v1.11 first, then upgrade to v1.12.

Upgrade your Sesame deployment:

```bash
$ kubectl apply -f examples/sesame/03-sesame.yaml
```

Note that the Sesame deployment needs to be updated before the Envoy daemon set since it contains backwards-compatible changes that are required in order to work with Envoy 1.17.0.
Once the Sesame deployment has finished upgrading, update the Envoy daemon set:

```bash
$ kubectl apply -f examples/sesame/03-envoy.yaml
```

## Upgrading Sesame 1.10.0 to 1.11.0

### Required Envoy version

All users should ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.16.2`.

Please see the [Envoy Release Notes][27] for information about issues fixed in Envoy 1.16.2.

### The easy way to upgrade

If the following are true for you:

* Your installation is in the `projectsesame` namespace.
* You are using our [quickstart example][18] deployments.
* Your cluster can take a few minutes of downtime.

Then the simplest way to upgrade to 1.11.0 is to delete the `projectsesame` namespace and reapply one of the example configurations:

```bash
$ kubectl delete namespace projectsesame
$ kubectl apply -f {{< param base_url >}}/quickstart/v1.11.0/sesame.yaml
```

This will remove the Envoy and Sesame pods from your cluster and recreate them with the updated configuration.
If you're using a `LoadBalancer` Service, (which most of the examples do) deleting and recreating may change the public IP assigned by your cloud provider.
You'll need to re-check where your DNS names are pointing as well, using [Get your hostname or IP address][12].

### The less easy way

This section contains information for administrators who wish to apply the Sesame 1.10.0 to 1.11.0 changes manually.
The YAML files referenced in this section can be found by cloning the Sesame repository and checking out the `v1.11.0` tag.

The Sesame CRD definition must be re-applied to the cluster, since a number of compatible changes and additions have been made to the Sesame API:

```bash
$ kubectl apply -f examples/sesame/01-crds.yaml
```

Users of the example deployment should first reapply the certgen Job YAML which will re-generate the relevant Secrets in the new format, which is compatible with [cert-manager](https://cert-manager.io) TLS secrets.
This will rotate the TLS certificates used for gRPC security.

```bash
$ kubectl apply -f examples/sesame/02-job-certgen.yaml
```

If your version of Sesame is older than v1.10, please upgrade to v1.10 first, then upgrade to v1.11.
For more information, see the [xDS Migration Guide][26].

```bash
$ kubectl apply -f examples/sesame/03-sesame.yaml
$ kubectl apply -f examples/sesame/03-envoy.yaml
```

## Upgrading Sesame 1.9.0 to 1.10.0

### Required Envoy version

All users should ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.16.0`.

Please see the [Envoy Release Notes][25] for information about issues fixed in Envoy 1.16.0.

### The easy way to upgrade

If the following are true for you:

 * Your installation is in the `projectsesame` namespace.
 * You are using our [quickstart example][18] deployments.
 * Your cluster can take a few minutes of downtime.

Then the simplest way to upgrade to 1.10.0 is to delete the `projectsesame` namespace and reapply one of the example configurations:

```bash
$ kubectl delete namespace projectsesame
$ kubectl apply -f {{< param base_url >}}/quickstart/v1.10.0/sesame.yaml
```

This will remove the Envoy and Sesame pods from your cluster and recreate them with the updated configuration.
If you're using a `LoadBalancer` Service, (which most of the examples do) deleting and recreating may change the public IP assigned by your cloud provider.
You'll need to re-check where your DNS names are pointing as well, using [Get your hostname or IP address][12].

### The less easy way

This section contains information for administrators who wish to apply the Sesame 1.9.0 to 1.10.0 changes manually.
The YAML files referenced in this section can be found by cloning the Sesame repository and checking out the `v1.10.0` tag.

The Sesame CRD definition must be re-applied to the cluster, since a number of compatible changes and additions have been made to the Sesame API:

```bash
$ kubectl apply -f examples/sesame/01-crds.yaml
```

Users of the example deployment should first reapply the certgen Job YAML which will re-generate the relevant Secrets in the new format, which is compatible with [cert-manager](https://cert-manager.io) TLS secrets.
This will rotate the TLS certificates used for gRPC security.

```bash
$ kubectl apply -f examples/sesame/02-job-certgen.yaml
```

If your cluster cannot take downtime, it's important to first upgrade Sesame to v1.10.0 then upgrade your Envoy pods.
This is due to an Envoy xDS Resource API upgrade to `v3`.
See the [xDS Migration Guide][26] for more information.

```bash
$ kubectl apply -f examples/sesame/03-sesame.yaml
$ kubectl apply -f examples/sesame/03-envoy.yaml
```

## Upgrading Sesame 1.8.2 to 1.9.0

### Required Envoy version

All users should ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.15.1`.

Please see the [Envoy Release Notes][23] for information about issues fixed in Envoy 1.15.1.

### The easy way to upgrade

If the following are true for you:

 * Your installation is in the `projectsesame` namespace.
 * You are using our [quickstart example][18] deployments.
 * Your cluster can take a few minutes of downtime.

Then the simplest way to upgrade to 1.9.0 is to delete the `projectsesame` namespace and reapply one of the example configurations:

```bash
$ kubectl delete namespace projectsesame
$ kubectl apply -f {{< param base_url >}}/quickstart/v1.9.0/sesame.yaml
```

This will remove the Envoy and Sesame pods from your cluster and recreate them with the updated configuration.
If you're using a `LoadBalancer` Service, (which most of the examples do) deleting and recreating may change the public IP assigned by your cloud provider.
You'll need to re-check where your DNS names are pointing as well, using [Get your hostname or IP address][12].

### The less easy way

This section contains information for administrators who wish to apply the Sesame 1.8.2 to 1.9.0 changes manually.
The YAML files referenced in this section can be found by cloning the Sesame repository and checking out the `v1.9.0` tag.

The Sesame CRD definition must be re-applied to the cluster, since a number of compatible changes and additions have been made to the Sesame API:

```bash
$ kubectl apply -f examples/sesame/01-crds.yaml
```

Users of the example deployment should first reapply the certgen Job YAML which will re-generate the relevant Secrets in the new format, which is compatible with [cert-manager](https://cert-manager.io) TLS secrets.
This will rotate the TLS certificates used for gRPC security.

```bash
$ kubectl apply -f examples/sesame/02-job-certgen.yaml
```

### Removing the IngressRoute CRDs

As a reminder, support for `IngressRoute` was officially dropped in v1.6.
If you haven't already migrated to `HTTPProxy`, see [the IngressRoute to HTTPProxy migration guide][24] for instructions on how to do so.
Once you have migrated, delete the `IngressRoute` and related CRDs:

```bash
$ kubectl delete crd ingressroutes.sesame.heptio.com
$ kubectl delete crd tlscertificatedelegations.sesame.heptio.com
```

## Upgrading Sesame 1.7.0 to 1.8.0

### Required Envoy version

All users should ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.15.0`.

Please see the [Envoy Release Notes][23] for information about issues fixed in Envoy 1.15.0.

### The easy way to upgrade

If the following are true for you:

 * Your installation is in the `projectsesame` namespace.
 * You are using our [quickstart example][18] deployments.
 * Your cluster can take a few minutes of downtime.

Then the simplest way to upgrade to 1.8.0 is to delete the `projectsesame` namespace and reapply one of the example configurations:

```bash
$ kubectl delete namespace projectsesame
$ kubectl apply -f {{< param base_url >}}/quickstart/v1.8.0/sesame.yaml
```

This will remove the Envoy and Sesame pods from your cluster and recreate them with the updated configuration.
If you're using a `LoadBalancer` Service, (which most of the examples do) deleting and recreating may change the public IP assigned by your cloud provider.
You'll need to re-check where your DNS names are pointing as well, using [Get your hostname or IP address][12].

### The less easy way

This section contains information for administrators who wish to apply the Sesame 1.7.0 to 1.8.0 changes manually.
The YAML files referenced in this section can be found by cloning the Sesame repository and checking out the `v1.8.0` tag.

The Sesame CRD definition must be re-applied to the cluster, since a number of compatible changes and additions have been made to the Sesame API:

```bash
$ kubectl apply -f examples/sesame/01-crds.yaml
```

Users of the example deployment should first reapply the certgen Job YAML which will re-generate the relevant Secrets in the new format, which is compatible with [cert-manager](https://cert-manager.io) TLS secrets. This will rotate the TLS certificates used for gRPC security.


```bash
$ kubectl apply -f examples/sesame/02-job-certgen.yaml
```

### Removing the IngressRoute CRDs

As a reminder, support for `IngressRoute` was officially dropped in v1.6.
If you haven't already migrated to `HTTPProxy`, see [the IngressRoute to HTTPProxy migration guide][24] for instructions on how to do so.
Once you have migrated, delete the `IngressRoute` and related CRDs:

```bash
$ kubectl delete crd ingressroutes.sesame.heptio.com
$ kubectl delete crd tlscertificatedelegations.sesame.heptio.com
```

## Upgrading Sesame 1.6.1 to 1.7.0

### Required Envoy version

All users should ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.15.0`.

Please see the [Envoy Release Notes][23] for information about issues fixed in Envoy 1.15.0.

### The easy way to upgrade

If the following are true for you:

 * Your installation is in the `projectsesame` namespace.
 * You are using our [quickstart example][18] deployments.
 * Your cluster can take few minutes of downtime.

Then the simplest way to upgrade to 1.7.0 is to delete the `projectsesame` namespace and reapply one of the example configurations:

```bash
$ kubectl delete namespace projectsesame
$ kubectl apply -f {{< param base_url >}}/quickstart/v1.7.0/sesame.yaml
```

This will remove the Envoy and Sesame pods from your cluster and recreate them with the updated configuration.
If you're using a `LoadBalancer` Service, (which most of the examples do) deleting and recreating may change the public IP assigned by your cloud provider.
You'll need to re-check where your DNS names are pointing as well, using [Get your hostname or IP address][12].

### The less easy way

This section contains information for administrators who wish to apply the Sesame 1.6.1 to 1.7.0 changes manually.
The YAML files referenced in this section can be found by cloning the Sesame repository and checking out the `v1.7.0` tag.

The Sesame CRD definition must be re-applied to the cluster, since a number of compatible changes and additions have been made to the Sesame API:

```bash
$ kubectl apply -f examples/sesame/01-crds.yaml
```

Users of the example deployment should first reapply the certgen Job YAML which will re-generate the relevant Secrets in the new format, which is compatible with [cert-manager](https://cert-manager.io) TLS secrets. This will rotate the TLS certs used for gRPC security.


```bash
$ kubectl apply -f examples/sesame/02-job-certgen.yaml
```

To consume the new Secrets, reapply the Envoy Daemonset and the Sesame Deployment YAML.
All the Pods will gracefully restart and reconnect using the new TLS Secrets.
After this, the gRPC session between Sesame and Envoy can be re-keyed by regenerating the Secrets.

```bash
$ kubectl apply -f examples/sesame/03-sesame.yaml
$ kubectl apply -f examples/sesame/03-envoy.yaml
```

### Removing the IngressRoute CRDs

As a reminder, support for `IngressRoute` was officially dropped in v1.6.
If you haven't already migrated to `HTTPProxy`, see [the IngressRoute to HTTPProxy migration guide][24] for instructions on how to do so.
Once you have migrated, delete the `IngressRoute` and related CRDs:

```bash
$ kubectl delete crd ingressroutes.sesame.heptio.com
$ kubectl delete crd tlscertificatedelegations.sesame.heptio.com
```

## Upgrading Sesame 1.5.1 to 1.6.1

### Required Envoy version

All users should ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.14.3`.

Please see the [Envoy Release Notes][22] for information about issues fixed in Envoy 1.14.3.

### The easy way to upgrade

If the following are true for you:

 * Your installation is in the `projectsesame` namespace.
 * You are using our [quickstart example][18] deployments.
 * Your cluster can take few minutes of downtime.

Then the simplest way to upgrade to 1.6.1 is to delete the `projectsesame` namespace and reapply one of the example configurations:

```bash
$ kubectl delete crd ingressroutes.sesame.heptio.com
$ kubectl delete crd tlscertificatedelegations.sesame.heptio.com
$ kubectl delete namespace projectsesame
$ kubectl apply -f {{< param base_url >}}/quickstart/v1.6.1/sesame.yaml
```

This will remove the IngressRoute CRD, and both the Envoy and Sesame pods from your cluster and recreate them with the updated configuration.
If you're using a `LoadBalancer` Service, (which most of the examples do) deleting and recreating may change the public IP assigned by your cloud provider.
You'll need to re-check where your DNS names are pointing as well, using [Get your hostname or IP address][12].

### The less easy way

This section contains information for administrators who wish to apply the Sesame 1.5.1 to 1.6.1 changes manually.
The YAML files referenced in this section can be found by cloning the Sesame repository and checking out the `v1.6.1` tag.

The Sesame CRD definition must be re-applied to the cluster, since a number of compatible changes and additions have been made to the Sesame API:

```bash
$ kubectl apply -f examples/sesame/01-crds.yaml
```

Administrators should also remove the IngressRoute CRDs:
```bash
$ kubectl delete crd ingressroutes.sesame.heptio.com
$ kubectl delete crd tlscertificatedelegations.sesame.heptio.com
```

Users of the example deployment should first reapply the certgen Job YAML which will re-generate the relevant Secrets in the new format, which is compatible with [cert-manager](https://cert-manager.io) TLS secrets. This will rotate the TLS certs used for gRPC security.


```bash
$ kubectl apply -f examples/sesame/02-job-certgen.yaml
```

To consume the new Secrets, reapply the Envoy Daemonset and the Sesame Deployment YAML.
All the Pods will gracefully restart and reconnect using the new TLS Secrets.
After this, the gRPC session between Sesame and Envoy can be re-keyed by regenerating the Secrets.

```bash
$ kubectl apply -f examples/sesame/03-sesame.yaml
$ kubectl apply -f examples/sesame/03-envoy.yaml
```

If you are upgrading from Sesame 1.6.0, the only required change is to upgrade the version of the Envoy image version from `v1.14.2` to `v1.14.3`.
The Sesame image can optionally be upgraded to `v1.6.1`.


## Upgrading Sesame 1.4.0 to 1.5.1

### Required Envoy version

All users should ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.14.2`.

Please see the [Envoy Release Notes][21] for information about issues fixed in Envoy 1.14.2.

### The easy way to upgrade

If the following are true for you:

 * Your installation is in the `projectsesame` namespace.
 * You are using our [quickstart example][18] deployments.
 * Your cluster can take few minutes of downtime.

Then the simplest way to upgrade to 1.5.1 is to delete the `projectsesame` namespace and reapply one of the example configurations:

```bash
$ kubectl delete namespace projectsesame
$ kubectl apply -f {{< param base_url >}}/quickstart/v1.5.1/sesame.yaml
```

This will remove both the Envoy and Sesame pods from your cluster and recreate them with the updated configuration.
If you're using a `LoadBalancer` Service, (which most of the examples do) deleting and recreating may change the public IP assigned by your cloud provider.
You'll need to re-check where your DNS names are pointing as well, using [Get your hostname or IP address][12].

### The less easy way

This section contains information for administrators who wish to apply the Sesame 1.4.0 to 1.5.1 changes manually.
The YAML files referenced in this section can be found by cloning the Sesame repository and checking out the `v1.5.1` tag.

The Sesame CRD definition must be re-applied to the cluster, since a number of compatible changes and additions have been made to the Sesame API:

```bash
$ kubectl apply -f examples/sesame/01-crds.yaml
```

In this release, the format of the TLS Secrets that are used to secure the gRPC session between Envoy and Sesame has changed.
This means that the Envoy Daemonset and the Sesame Deployment have been changed to mount the the TLS secrets volume differently.
Users of the example deployment should first reapply the certgen Job YAML which will re-generate the relevant Secrets in the new format, which is compatible with [cert-manager](https://cert-manager.io) TLS secrets.


```bash
$ kubectl apply -f examples/sesame/02-job-certgen.yaml
```

To consume the new Secrets, reapply the Envoy Daemonset and the Sesame Deployment YAML.
All the Pods will gracefully restart and reconnect using the new TLS Secrets.
After this, the gRPC session between Sesame and Envoy can be re-keyed by regenerating the Secrets.

```bash
$ kubectl apply -f examples/sesame/03-sesame.yaml
$ kubectl apply -f examples/sesame/03-envoy.yaml
```

Users who secure the gRPC session with their own certificate may need to modify the Envoy Daemonset and the Sesame Deployment to ensure that their Secrets are correctly mounted within the corresponding Pod containers.
When making these changes, be sure to retain the `--resources-dir` flag to the `Sesame bootstrap` command so that Envoy will be configured with reloadable TLS certificate support.

## Upgrading Sesame 1.3.0 to 1.4.0

### Required Envoy version

All users should ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.14.1`.

Please see the [Envoy Release Notes][20] for information about issues fixed in Envoy 1.14.1.

### The easy way to upgrade

If the following are true for you:

 * Your installation is in the `projectsesame` namespace.
 * You are using our [quickstart example][18] deployments.
 * Your cluster can take few minutes of downtime.

Then the simplest way to upgrade to 1.4.0 is to delete the `projectsesame` namespace and reapply one of the example configurations:

```bash
$ kubectl delete namespace projectsesame
$ kubectl apply -f {{< param base_url >}}/quickstart/v1.4.0/sesame.yaml
```

This will remove both the Envoy and Sesame pods from your cluster and recreate them with the updated configuration.
If you're using a `LoadBalancer` Service, (which most of the examples do) deleting and recreating may change the public IP assigned by your cloud provider.
You'll need to re-check where your DNS names are pointing as well, using [Get your hostname or IP address][12].

**Note:** If you deployed Sesame into a different namespace than `projectsesame` with a standard example, please delete that namespace.
Then in your editor of choice do a search and replace for `projectsesame` and replace it with your preferred name space and apply the updated manifest.

### The less easy way

This section contains information for administrators who wish to apply the Sesame 1.3.0 to 1.4.0 changes manually.

#### Upgrade to Sesame 1.4.0

Change the Sesame image version to `docker.io/projectsesame/Sesame:v1.4.0`

Because there has been a change to Envoy to add a serviceaccount, you need to reapply the Sesame CRDs and RBAC.

From within a clone of the repo, checkout `release-1.4`, then you can:

```bash
kubectl apply -f examples/sesame/00-common.yaml
kubectl apply -f examples/sesame/01-crds.yaml
kubectl apply -f examples/sesame/02-rbac.yaml
```

If you are using our Envoy daemonset:

```bash
kubectl apply -f examples/sesame/03-envoy.yaml
```

Otherwise, you should add the new `envoy` `serviceAccount` to your Envoy deployment.
This will be used in the future to add further container-level security via PodSecurityPolicies.

## Upgrading Sesame 1.2.1 to 1.3.0

### Required Envoy version

All users should ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.13.1`.

Please see the [Envoy Release Notes][17] for information about issues fixed in Envoy 1.13.1.

### The easy way to upgrade

If the following are true for you:

 * Your installation is in the `projectsesame` namespace.
 * You are using our [quickstart example][18] deployments.
 * Your cluster can take few minutes of downtime.

Then the simplest way to upgrade to 1.3.0 is to delete the `projectsesame` namespace and reapply one of the example configurations:

```bash
$ kubectl delete namespace projectsesame
$ kubectl apply -f {{< param base_url >}}/quickstart/v1.3.0/sesame.yaml
```

This will remove both the Envoy and Sesame pods from your cluster and recreate them with the updated configuration.
If you're using a `LoadBalancer` Service, (which most of the examples do) deleting and recreating may change the public IP assigned by your cloud provider.
You'll need to re-check where your DNS names are pointing as well, using [Get your hostname or IP address][12].

**Note:** If you deployed Sesame into a different namespace than `projectsesame` with a standard example, please delete that namespace.
Then in your editor of choice do a search and replace for `projectsesame` and replace it with your preferred name space and apply the updated manifest.

### The less easy way

This section contains information for administrators who wish to apply the Sesame 1.2.1 to 1.3.0 changes manually.

#### Upgrade to Sesame 1.3.0

Change the Sesame image version to `docker.io/projectsesame/Sesame:v1.3.0`

## Upgrading Sesame 1.2.0 to 1.2.1

### Required Envoy version

All users should ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.13.1`.

Please see the [Envoy Release Notes][17] for information about issues fixed in Envoy 1.13.1.

### The easy way to upgrade

If the following are true for you:

 * Your installation is in the `projectsesame` namespace.
 * You are using our [quickstart example][18] deployments.
 * Your cluster can take few minutes of downtime.

Then the simplest way to upgrade to 1.2.1 is to delete the `projectsesame` namespace and reapply one of the example configurations.
From the root directory of the repository:

```bash
$ kubectl delete namespace projectsesame
$ kubectl apply -f {{< param base_url >}}/quickstart/v1.2.1/sesame.yaml
```

This will remove both the Envoy and Sesame pods from your cluster and recreate them with the updated configuration.
If you're using a `LoadBalancer` Service, (which most of the examples do) deleting and recreating may change the public IP assigned by your cloud provider.
You'll need to re-check where your DNS names are pointing as well, using [Get your hostname or IP address][12].

**Note:** If you deployed Sesame into a different namespace than `projectsesame` with a standard example, please delete that namespace.
Then in your editor of choice do a search and replace for `projectsesame` and replace it with your preferred name space and apply the updated manifest.

### The less easy way

This section contains information for administrators who wish to apply the Sesame 1.2.0 to 1.2.1 changes manually.

#### Upgrade to Sesame 1.2.1

Change the Sesame image version to `docker.io/projectsesame/Sesame:v1.2.1`.

#### Upgrade to Envoy 1.13.1

Sesame 1.2.1 requires Envoy 1.13.1.
Change the Envoy image version to `docker.io/envoyproxy/envoy:v1.13.1`.

_Note: Envoy 1.13.1 includes fixes to a number of [CVEs][19]_

## Upgrading Sesame 1.1.0 to 1.2.1

### Required Envoy version

All users should ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.13.1`.

Please see the [Envoy Release Notes][17] for information about issues fixed in Envoy 1.13.1.

### The easy way to upgrade

If the following are true for you:

 * Your installation is in the `projectsesame` namespace.
 * You are using our [quickstart example][18] deployments.
 * Your cluster can take few minutes of downtime.

Then the simplest way to upgrade to 1.2.1 is to delete the `projectsesame` namespace and reapply one of the example configurations.
From the root directory of the repository:

```bash
$ kubectl delete namespace projectsesame
$ kubectl apply -f {{< param base_url >}}/quickstart/v1.2.1/sesame.yaml
```

This will remove both the Envoy and Sesame pods from your cluster and recreate them with the updated configuration.
If you're using a `LoadBalancer` Service, (which most of the examples do) deleting and recreating may change the public IP assigned by your cloud provider.
You'll need to re-check where your DNS names are pointing as well, using [Get your hostname or IP address][12].

**Note:** If you deployed Sesame into a different namespace than `projectsesame` with a standard example, please delete that namespace.
Then in your editor of choice do a search and replace for `projectsesame` and replace it with your preferred name space and apply the updated manifest.

### The less easy way

This section contains information for administrators who wish to apply the Sesame 1.1.0 to 1.2.1 changes manually.

#### Upgrade to Sesame 1.2.1

Change the Sesame image version to `docker.io/projectsesame/Sesame:v1.2.1`.

#### Upgrade to Envoy 1.13.1

Sesame 1.2.1 requires Envoy 1.13.1.
Change the Envoy image version to `docker.io/envoyproxy/envoy:v1.13.0`.

#### Envoy shutdown manager

Sesame 1.2.1 introduces a new sidecar to aid graceful shutdown of the Envoy pod.
Consult [shutdown manager]({% link docs/v1.2.1/redeploy-envoy.md %}) documentation for installation instructions.

## Upgrading Sesame 1.0.1 to 1.1.0

### Required Envoy version

All users should ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.12.2`.

Please see the [Envoy Release Notes][15] for information about issues fixed in Envoy 1.12.2.

### The easy way to upgrade

If the following are true for you:

 * Your installation is in the `projectsesame` namespace.
 * You are using one of the [example][1] deployments.
 * Your cluster can take few minutes of downtime.

Then the simplest way to upgrade to 1.1.0 is to delete the `projectsesame` namespace and reapply one of the example configurations.
From the root directory of the repository:

```bash
$ kubectl delete namespace projectsesame
$ kubectl apply -f examples/<your-desired-deployment>
```

This will remove both the Envoy and Sesame pods from your cluster and recreate them with the updated configuration.
If you're using a `LoadBalancer` Service, (which most of the examples do) deleting and recreating may change the public IP assigned by your cloud provider.
You'll need to re-check where your DNS names are pointing as well, using [Get your hostname or IP address][12].

**Note:** If you deployed Sesame into a different namespace than `projectsesame` with a standard example, please delete that namespace.
Then in your editor of choice do a search and replace for `projectsesame` and replace it with your preferred name space and apply the updated manifest.

**Note:** If you are deploying to a cluster where you have previously installed alpha versions of the Sesame API, applying the Sesame CRDs in `examples/Sesame` may fail with a message similar to `Invalid value: "v1alpha1": must appear in spec.versions`. In this case, you need to delete the old CRDs and apply the new ones.

```bash
$ kubectl delete namespace projectsesame
$ kubectl get crd  | awk '/projectsesame.io|sesame.heptio.com/{print $1}' | xargs kubectl delete crd
$ kubectl apply -f examples/<your-desired-deployment>
```

### The less easy way

This section contains information for administrators who wish to apply the Sesame 1.0.1 to 1.1.0 changes manually.

#### Upgrade to Sesame 1.1.0

Change the Sesame image version to `docker.io/projectsesame/Sesame:v1.1.0`.

#### Upgrade to Envoy 1.12.2

Sesame 1.1.0 requires Envoy 1.12.2. Change the Envoy image version to `docker.io/envoyproxy/envoy:v1.12.2`.

## Upgrading Sesame 1.0.0 to 1.0.1

### The easy way to upgrade

If you are running Sesame 1.0.0, the easy way to upgrade to Sesame 1.0.1 is to reapply the [quickstart yaml][16].

```bash
$ kubectl apply -f {{< param base_url >}}/quickstart/v1.0.1/sesame.yaml
```

### The less easy way

This section contains information for administrators who wish to manually upgrade from Sesame 1.0.0 to Sesame 1.0.1.

#### Sesame version

Ensure the Sesame image version is `docker.io/projectsesame/Sesame:v1.0.1`.

#### Envoy version

Ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.12.2`.

Please see the [Envoy Release Notes][15] for information about issues fixed in Envoy 1.12.2.

## Upgrading Sesame 0.15.3 to 1.0.0

### Required Envoy version

The required version of Envoy remains unchanged.
Ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.11.2`.

### The easy way to upgrade

If the following are true for you:

 * Your previous installation is in the `projectsesame` namespace.
 * You are using one of the [example][2] deployments.
 * Your cluster can take few minutes of downtime.

Then the simplest way to upgrade is to delete the `projectsesame` namespace and reapply the `examples/Sesame` sample manifest.
From the root directory of the repository:

```bash
$ kubectl delete namespace projectsesame
$ kubectl apply -f examples/sesame
```

This will remove both the Envoy and Sesame pods from your cluster and recreate them with the updated configuration.
If you're using a `LoadBalancer` Service, deleting and recreating may change the public IP assigned by your cloud provider.
You'll need to re-check where your DNS names are pointing as well, using [Get your hostname or IP address][12].

### The less easy way

This section contains information for administrators who wish to manually upgrade from Sesame 0.15.3 to Sesame 1.0.0.

#### Upgrade to Sesame 1.0.0

Change the Sesame image version to `docker.io/projectsesame/Sesame:v1.0.0`.

Note that as part of sunsetting the Heptio brand, Sesame Docker images have moved from `gcr.io/heptio-images` to `docker.io/projectsesame`.

#### Reapply HTTPProxy and IngressRoute CRD definitions

Sesame 1.0.0 ships with updated OpenAPIv3 validation schemas.

Sesame 1.0.0 promotes the HTTPProxy CRD to v1.
HTTPProxy is now considered stable, and there will only be additive, compatible changes in the future.
See the [HTTPProxy documentation][3] for more information.

```bash
$ kubectl apply -f examples/sesame/01-crds.yaml
```

#### Update deprecated `Sesame.heptio.com` annotations

All the annotations with the prefix `Sesame.heptio.com` have been migrated to their respective `projectsesame.io` counterparts.
The deprecated `Sesame.heptio.com` annotations will be recognized through the Sesame 1.0 release, but are scheduled to be removed after Sesame 1.0.

See the [annotation documentation][4] for more information.

#### Update old `projectsesame.io/v1alpha1` group versions

If you are upgrading a cluster that you previously installed a 1.0.0 release candidate, note that Sesame 1.0.0 moves the HTTPProxy CRD from `projectsesame.io/v1alpha1` to `projectsesame.io/v1` and will no longer recognize the former group version.

Please edit your HTTPProxy documents to update their group version to `projectsesame.io/v1`.

#### Check for HTTPProxy v1 schema changes

As part of finalizing the HTTPProxy v1 schema, three breaking changes have been introduced.
If you are upgrading a cluster that you previously installed a Sesame 1.0.0 release candidate, you may need to edit HTTPProxy object to conform to the upgraded schema.

* The per-route prefix rewrite key, `prefixRewrite` has been removed.
  See [#899][5] for the status of its replacement.

* The per-service health check key, `healthcheck` has moved to per-route and has been renamed `healthCheckPolicy`.

<table class="table table-borderless" style="border: none;">
<tr><th>Before:</th><th>After:</th></tr>

<tr>
<td><pre><code class="language-yaml" data-lang="yaml">
spec:
  routes:
  - conditions:
    - prefix: /
    services:
    - name: www
      port: 80
      healthcheck:
      - path: /healthy
        intervalSeconds: 5
        timeoutSeconds: 2
        unhealthyThresholdCount: 3
        healthyThresholdCount: 5
</code></pre></td>

<td>
<pre><code class="language-yaml" data-lang="yaml">
spec:
  routes:
  - conditions:
    - prefix: /
    healthCheckPolicy:
    - path: /healthy
      intervalSeconds: 5
      timeoutSeconds: 2
      unhealthyThresholdCount: 3
      healthyThresholdCount: 5
    services:
    - name: www
      port: 80
</code></pre></td>

</tr>
</table>

* The per-service load balancer strategy key, `strategy` has moved to per-route and has been renamed `loadBalancerPolicy`.

<table class="table table-borderless" style="border: none;">
<tr><th>Before:</th><th>After:</th></tr>

<tr>
<td><pre><code class="language-yaml" data-lang="yaml">
spec:
  routes:
  - conditions:
    - prefix: /
    services:
    - name: www
      port: 80
      strategy: WeightedLeastRequest
</code></pre></td>

<td><pre><code class="language-yaml" data-lang="yaml">
spec:
  routes:
  - conditions:
    - prefix: /
    loadBalancerPolicy:
      strategy: WeightedLeastRequest
    services:
    - name: www
      port: 80
</code></pre></td>

</tr>
</table>

##### Check for Sesame namespace change

As part of sunsetting the Heptio brand the `heptio-Sesame` namespace has been renamed to `projectsesame`.
Sesame assumes it will be deployed into the `projectsesame` namespace.

If you deploy Sesame into a different namespace you will need to pass `Sesame bootstrap --namespace=<namespace>` and update the leader election parameters in the [`Sesame.yaml` configuration][6]
as appropriate.

#### Split deployment/daemonset now the default

We have changed the example installation to use a separate pod installation, where Sesame is in a Deployment and Envoy is in a Daemonset.
Separated pod installations separate the lifecycle of Sesame and Envoy, increasing operability.
Because of this, we are marking the single pod install type as officially deprecated.
If you are still running a single pod install type, please review the [`Sesame` example][7] and either adapt it or use it directly.

#### Verify leader election

Sesame 1.0.0 enables leader election by default.
No specific configuration is required if you are using the [example deployment][7].

Leader election requires that Sesame have write access to a ConfigMap
called `leader-elect` in the project-Sesame namespace.
This is done with the [Sesame-leaderelection Role][8] in the [example RBAC][9].
The namespace and name of the configmap are configurable via the configuration file.

The leader election mechanism no longer blocks serving of gRPC until an instance becomes the leader.
Leader election controls writing status back to Sesame CRDs (like HTTPProxy and IngressRoute) so that only one Sesame pod writes status at a time.

Should you wish to disable leader election, pass `Sesame serve --disable-leader-election`.

#### Envoy pod readiness checks

Update the readiness checks on your Envoy pod's spec to reflect Envoy 1.11.1's `/ready` endpoint
```yaml
readinessProbe:
  httpGet:
    path: /ready
    port: 8002
```

#### Root namespace restriction

The `Sesame serve --ingressroute-root-namespaces` flag has been renamed to `--root-namespaces`.
If you use this feature please update your deployments.

## Upgrading Sesame 0.14.x to 0.15.3

Sesame 0.15.3 requires changes to your deployment manifests to explicitly opt in, or opt out of, secure communication between Sesame and Envoy.

Sesame 0.15.3 also adds experimental support for leader election which may be useful for installations which have split their Sesame and Envoy containers into separate pods.
A configuration we call _split deployment_.

### Breaking change

Sesame's `Sesame serve` now requires that either TLS certificates be available, or you supply the `--insecure` parameter.

**If you do not supply TLS details or `--insecure`, `Sesame serve` will not start.**

### Recommended Envoy version

All users should ensure the Envoy image version is `docker.io/envoyproxy/envoy:v1.11.2`.

Please see the [Envoy Release Notes][10] for information about issues fixed in Envoy 1.11.2.

### The easy way to upgrade

If the following are true for you:

 * Your installation is in the `heptio-Sesame` namespace.
 * You are using one of the [example][11] deployments.
 * Your cluster can take few minutes of downtime.

Then the simplest way to upgrade to 0.15.3 is to delete the `heptio-Sesame` namespace and reapply one of the example configurations.
From the root directory of the repository:

```bash
$ kubectl delete namespace heptio-sesame
$ kubectl apply -f examples/<your-desired-deployment>
```

If you're using a `LoadBalancer` Service, (which most of the examples do) deleting and recreating may change the public IP assigned by your cloud provider.
You'll need to re-check where your DNS names are pointing as well, using [Get your hostname or IP address][12].

**Note:** If you deployed Sesame into a different namespace than heptio-Sesame with a standard example, please delete that namespace.
Then in your editor of choice do a search and replace for `heptio-Sesame` and replace it with your preferred name space and apply the updated manifest.

### The less easy way

This section contains information for administrators who wish to apply the Sesame 0.14.x to 0.15.3 changes manually.

#### Upgrade to Sesame 0.15.3

Due to the sun setting on the Heptio brand, from v0.15.0 onwards our images are now served from the docker hub repository [`docker.io/projectsesame/Sesame`][13]

Change the Sesame image version to `docker.io/projectsesame/Sesame:v0.15.3`.

#### Enabling TLS for gRPC

You *must* either enable TLS for gRPC serving, or put `--insecure` into your `Sesame serve` startup line.
If you are running with both Sesame and Envoy in a single pod, the existing deployment examples have already been updated with this change.

If you are running using the `ds-hostnet-split` example or a derivative, we strongly recommend that you generate new certificates for securing your gRPC communication between Sesame and Envoy.

There is a Job in the `ds-hostnet-split` directory that will use the new `Sesame certgen` command to generate a CA and then sign Sesame and Envoy keypairs, which can also then be saved directly to Kubernetes as Secrets, ready to be mounted into your Sesame and Envoy Deployments and Daemonsets.

If you would like more detail, see [grpc-tls-howto.md][14], which explains your options.

#### Upgrade to Envoy 1.11.2

Sesame 0.15.3 requires Envoy 1.11.2. Change the Envoy image version to `docker.io/envoyproxy/envoy:v1.11.2`.

#### Enabling Leader Election

Sesame 0.15.3 adds experimental support for leader election.
Enabling leader election will mean that only one of the Sesame pods will actually serve gRPC traffic.
This will ensure that all Envoy's take their configuration from the same Sesame.
You can enable leader election with the `--enable-leader-election` flag to `Sesame serve`.

If you have deployed Sesame and Envoy in their own pods--we call this split deployment--you should enable leader election so all envoy pods take their configuration from the lead Sesame.

To enable leader election, the following must be true for you:

- You are running in a split Sesame and Envoy setup.
  That is, there are separate Sesame and Envoy pod(s).

In order for leader election to work, you must make the following changes to your setup:

- The Sesame Deployment must have its readiness probe changed too TCP readiness probe configured to check port 8001 (the gRPC port), as non-leaders will not serve gRPC, and Envoys may not be properly configured if they attempt to connect to a non-leader Sesame.
  That is, you will need to change:

```yaml
        readinessProbe:
          httpGet:
            path: /healthz
            port: 8000
```
to

```yaml
        readinessProbe:
          tcpSocket:
            port: 8001
          initialDelaySeconds: 15
          periodSeconds: 10
```
inside the Pod spec.
- The update strategy for the Sesame deployment must be changed to `Recreate` instead of `RollingUpdate`, as pods will never become Ready (since they won't pass the readiness probe).
  Add

```yaml
  strategy:
    type: Recreate
```
to the top level of the Pod spec.
- Leader election is currently hard-coded to use a ConfigMap named `Sesame` in this namespace for the leader election lock.
If you are using a newer installation of Sesame, this may be present already, if not, the leader election library will create an empty ConfigMap for you.

Once these changes are made, add `--enable-leader-election` to your `Sesame serve` command.
The leader will perform and log its operations as normal, and the non-leaders will block waiting to become leader.
You can inspect the state of the leadership using

```bash
$ kubectl describe configmap -n heptio-sesame sesame
```

and checking the annotations that store exact details using

```bash
$ kubectl get configmap -n heptio-sesame -o yaml sesame
```

[1]: https://github.com/projectsesame/sesame/tree/main/examples/Sesame
[2]: https://github.com/projectsesame/sesame/blob/v1.0.0/examples
[3]: /docs/main/config/fundamentals
[4]: /docs/main/config/annotations
[5]: https://github.com/projectsesame/sesame/issues/899
[6]: /docs/main/configuration
[7]: https://github.com/projectsesame/sesame/blob/main/examples/Sesame/README.md
[8]: https://github.com/projectsesame/sesame/blob/v1.0.0/examples/Sesame/02-rbac.yaml#L71
[9]: https://github.com/projectsesame/sesame/blob/main/examples/Sesame/02-rbac.yaml
[10]: https://www.envoyproxy.io/docs/envoy/v1.11.2/intro/version_history
[11]: https://github.com/projectsesame/sesame/blob/v0.15.3/examples/
[12]: /docs/main/deploy-options
[13]: https://hub.docker.com/r/projectsesame/Sesame
[14]: /docs/main/grpc-tls-howto
[15]: https://www.envoyproxy.io/docs/envoy/v1.12.2/intro/version_history
[16]: /getting-started
[17]: https://www.envoyproxy.io/docs/envoy/v1.13.1/intro/version_history
[18]: https://projectsesame.io/quickstart/main/Sesame.yaml
[19]: https://groups.google.com/forum/?utm_medium=email&utm_source=footer#!msg/envoy-announce/sVqmxy0un2s/8aq430xiHAAJ
[20]: https://www.envoyproxy.io/docs/envoy/v1.14.1/intro/version_history
[21]: https://www.envoyproxy.io/docs/envoy/v1.14.2/intro/version_history
[22]: https://www.envoyproxy.io/docs/envoy/v1.14.3/intro/version_history
[23]: https://www.envoyproxy.io/docs/envoy/v1.15.0/version_history/current
[24]: /guides/ingressroute-to-httpproxy/
[25]: https://www.envoyproxy.io/docs/envoy/v1.16.0/version_history/current
[26]: /guides/xds-migration
[27]: https://www.envoyproxy.io/docs/envoy/v1.16.2/version_history/current
[28]: https://www.envoyproxy.io/docs/envoy/v1.17.0/version_history/current
[29]: https://www.envoyproxy.io/docs/envoy/v1.17.1/version_history/current
[30]: https://www.envoyproxy.io/docs/envoy/v1.17.2/version_history/current
[31]: https://www.envoyproxy.io/docs/envoy/v1.18.2/version_history/current
[32]: https://www.envoyproxy.io/docs/envoy/v1.18.3/version_history/current
[33]: https://www.envoyproxy.io/docs/envoy/v1.19.0/version_history/current
[34]: https://www.envoyproxy.io/docs/envoy/v1.19.1/version_history/current
