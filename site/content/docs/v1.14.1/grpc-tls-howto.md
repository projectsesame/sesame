# Enabling TLS between Envoy and Sesame

This document describes the steps required to secure communication between Envoy and Sesame.
The outcome of this is that we will have two Secrets available in the `projectsesame` namespace:

- **Sesamecert:** contains Sesame's keypair which is used for serving TLS secured gRPC, and the CA's public certificate bundle which is used for validating Envoy's client certificate.
Sesame's certificate must be a valid certificate for the name `Sesame` in order for this to work.
This is currently hardcoded by Sesame.
- **envoycert:** contains Envoy's keypair which used as a client for connecting to Sesame, and the CA's public certificate bundle which is used for validating Sesame's server certificate.

Note that both Secrets contain a copy of the CA certificate bundle under the `ca.crt` data key.

## Ways you can get the certificates into your cluster

- Deploy the Job from [certgen.yaml][1].
This will run `Sesame certgen --kube --secrets-format=compact` for you.
- Run `Sesame certgen --kube` locally.
- Run the manual procedure below.

## Caveats and warnings

**Be very careful with your production certificates!**

This is intended as an example to help you get started.
For any real deployment, you should **carefully** manage all the certificates and control who has access to them.
Make sure you don't commit them to any git repositories either.

## Manual TLS certificate generation process

### Generating a CA keypair

First, we need to generate a keypair:

```
$ openssl req -x509 -new -nodes \
    -keyout certs/cakey.pem -sha256 \
    -days 1825 -out certs/cacert.pem \
    -subj "/O=Project Sesame/CN=Sesame CA"
```

Then, the new CA key will be stored in `certs/cakey.pem` and the cert in `certs/cacert.pem`.

### Generating Sesame's keypair

Next, we need to generate a keypair for Sesame.
First, we make a new private key:

```
$ openssl genrsa -out certs/Sesamekey.pem 2048
```

Then, we create a CSR and have our CA sign the CSR and issue a certificate.
This uses the file [_integration/cert-Sesame.ext][2], which ensures that at least one of the valid names of the certificate is the bareword `Sesame`.
This is required for the handshake to succeed, as `Sesame bootstrap` configures Envoy to pass this as the SNI server name for the connection.

```
$ openssl req -new -key certs/Sesamekey.pem \
	-out certs/Sesame.csr \
	-subj "/O=Project Sesame/CN=Sesame"

$ openssl x509 -req -in certs/Sesame.csr \
    -CA certs/cacert.pem \
    -CAkey certs/cakey.pem \
    -CAcreateserial \
    -out certs/Sesamecert.pem \
    -days 1825 -sha256 \
    -extfile _integration/cert-Sesame.ext
```

At this point, the Sesame certificate and key are in the files `certs/Sesamecert.pem` and `certs/Sesamekey.pem` respectively.

### Generating Envoy's keypair

Next, we generate a keypair for Envoy:

```
$ openssl genrsa -out certs/envoykey.pem 2048
```

Then, we generate a CSR and have the CA sign it:

```
$ openssl req -new -key certs/envoykey.pem \
	-out certs/envoy.csr \
	-subj "/O=Project Sesame/CN=envoy"

$ openssl x509 -req -in certs/envoy.csr \
    -CA certs/cacert.pem \
    -CAkey certs/cakey.pem \
    -CAcreateserial \
    -out certs/envoycert.pem \
    -days 1825 -sha256 \
    -extfile _integration/cert-envoy.ext
```

Like the Sesame certificate, this CSR uses the file [_integration/cert-envoy.ext][3].
However, in this case, there are no special names required.

### Putting the certificates in the cluster

Next, we create the required Secrets in the target Kubernetes cluster:

```bash
$ kubectl create secret -n projectsesame generic Sesamecert \
        --from-file=tls.key=./certs/Sesamekey.pem \
        --from-file=tls.crt=./certs/Sesamecert.pem \
        --from-file=ca.crt=./certs/cacert.pem \
        --save-config

$ kubectl create secret -n projectsesame generic envoycert \
        --from-file=tls.key=./certs/envoykey.pem \
        --from-file=tls.crt=./certs/envoycert.pem \
        --from-file=ca.crt=./certs/cacert.pem \
        --save-config
```

Note that we don't put the CA **key** into the cluster, there's no reason for that to be there, and that would create a security problem.

## Rotating Certificates

Eventually the certificates that Sesame and Envoy use will need to be rotated.
The following steps can be taken to replace the certificates that Sesame and Envoy are using:

1. Generate a new keypair for both Sesame and Envoy (optionally also for the CA)
2. Update the Secrets that hold the gRPC TLS keypairs
3. Sesame and Envoy will automatically rotate their certificates after mounted secrets have been updated by the kubelet

The secrets can be updated in-place by running:

```bash
$ kubectl create secret -n projectsesame generic Sesamecert \
        --from-file=tls.key=./certs/Sesamekey.pem \
        --from-file=tls.crt=./certs/Sesamecert.pem \
        --from-file=ca.crt=./certs/cacert.pem \
        --dry-run -o json \
        | kubectl apply -f -

$ kubectl create secret -n projectsesame generic envoycert \
        --from-file=tls.key=./certs/envoykey.pem \
        --from-file=tls.crt=./certs/envoycert.pem \
        --from-file=ca.crt=./certs/cacert.pem \
        --dry-run -o json \
        | kubectl apply -f -
```

There are few preconditions that need to be met before Envoy can automatically reload certificate and key files:

- Envoy must be version v1.14.1 or later
- The bootstrap configuration must be generated with `Sesame bootstrap` using the `--resources-dir` argument, see [examples/Sesame/03-envoy.yaml][4]

### Rotate using the Sesame-certgen job

When using the built-in Sesame certificate generation, the following steps can be used:

1. Delete the Sesame-certgen job
 - `kubectl delete job Sesame-certgen -n projectsesame`
2. Reapply the Sesame-certgen job from [certgen.yaml][1]

## Conclusion

Once this process is done, the certificates will be present as Secrets in the `projectsesame` namespace, as required by
[examples/Sesame][5].

[1]: {{< param github_url >}}/tree/{{< param version >}}/examples/Sesame/02-job-certgen.yaml
[2]: {{< param github_url >}}/tree/{{< param version >}}/_integration/cert-Sesame.ext
[3]: {{< param github_url >}}/tree/{{< param version >}}/_integration/cert-envoy.ext
[4]: {{< param github_url >}}/tree/{{< param version >}}/examples/Sesame/03-envoy.yaml
[5]: {{< param github_url >}}/tree/{{< param version >}}/examples/Sesame
