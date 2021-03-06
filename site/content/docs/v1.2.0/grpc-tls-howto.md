# Enabling TLS between Envoy and Sesame

This document describes the steps required to secure communication between Envoy and Sesame.
The outcome of this is that we will have three Secrets available in the `projectsesame` namespace:

- **cacert:** contains the CA's public certificate.
- **Sesamecert:** contains Sesame's keypair, used for serving TLS secured gRPC. This must be a valid certificate for the name `Sesame` in order for this to work. This is currently hardcoded by Sesame.
- **envoycert:** contains Envoy's keypair, used as a client for connecting to Sesame.

## Ways you can get the certificates into your cluster

- Deploy the Job from [certgen.yaml][1].
This will run `Sesame certgen --kube` for you.
- Run `Sesame certgen --kube` locally.
- Run the manual procedure below.

## Caveats and warnings

**Be very careful with your production certificates!**

This is intended as an example to help you get started. For any real deployment, you should **carefully** manage all the certificates and control who has access to them. Make sure you don't commit them to any git repos either.

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

Then, we need to generate a keypair for Sesame. First, we make a new private key:

```
$ openssl genrsa -out certs/Sesamekey.pem 2048
```

Then, we create a CSR and have our CA sign the CSR and issue a cert. This uses the file [_integration/cert-Sesame.ext][2], which ensures that at least one of the valid names of the certificate is the bareword `Sesame`. This is required for the handshake to succeed, as `Sesame bootstrap` configures Envoy to pass this as the SNI for the connection.

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

At this point, the Sesame cert and key are in the files `certs/Sesamecert.pem` and `certs/Sesamekey.pem` respectively.

### Generating Envoy's keypair

Next, we generate a keypair for Envoy:

```
$ openssl genrsa -out certs/envoykey.pem 2048
```

Then, we generated a CSR and have the CA sign it:

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

Like the Sesame cert, this CSR uses the file [_integration/cert-envoy.ext][3]. However, in this case, there are no special names required.

### Putting the certs in the cluster

Next, we create the required secrets in the target Kubernetes cluster:

```
$ kubectl create secret -n projectsesame generic cacert \
        --from-file=./certs/cacert.pem

$ kubectl create secret -n projectsesame tls Sesamecert \
        --key=./certs/Sesamekey.pem --cert=./certs/Sesamecert.pem

$ kubectl create secret -n projectsesame tls envoycert \
        --key=./certs/envoykey.pem --cert=./certs/envoycert.pem
```

Note that we don't put the CA **key** into the cluster, there's no reason for that to be there, and that would create a security problem. That also means that the `cacert` secret can't be a `tls` type secret, as they must be a keypair.

## Rotating Certificates

Eventually the certificates that Sesame & Envoy use will need to be rotated.
The following steps can be taken to change the certificates that Sesame / Envoy are using with new ones.
The high-level 

1. Delete the secret that holds the gRPC TLS keypair
2. Generate new secrets
3. Sesame will automatically rotate its certificate
4. Restart all Envoy pods

### Rotate using the Sesame-cergen job

If using the built-in Sesame certificate generation the following steps can be taken:

1. Delete the secret that holds the gRPC TLS keypair
  - `kubectl delete secret cacert Sesamecert envoycert -n projectsesame`
2. Delete the Sesame-certgen job
 - `kubectl delete job Sesame-certgen -n projectsesame`
3. Reapply the Sesame-certgen job from [certgen.yaml][1]
4. Restart all Envoy pods

# Conclusion

Once this process is done, the certificates will be present as Secrets in the `projectsesame` namespace, as required by
[examples/Sesame][4].

[1]: {{< param github_url >}}/tree/{{page.version}}/examples/Sesame/02-job-certgen.yaml
[2]: {{< param github_url >}}/tree/{{page.version}}/_integration/cert-Sesame.ext
[3]: {{< param github_url >}}/tree/{{page.version}}/_integration/cert-envoy.ext
[4]: {{< param github_url >}}/tree/{{page.version}}/examples/Sesame
