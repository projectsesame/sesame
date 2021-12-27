# Sesame Configuration

A configuration file can be passed to the `Sesame serve` command which specified additional properties that Sesame should use when starting up.
This file is passed to Sesame via a ConfigMap which is mounted as a volume to the Sesame pod.

Following is an example ConfigMap with configuration file included: 

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: sesame
  namespace: projectsesame
data:
  Sesame.yaml: |
    # should Sesame expect to be running inside a k8s cluster
    # incluster: true
    #
    # path to kubeconfig (if not running inside a k8s cluster)
    # kubeconfig: /path/to/.kube/config
    #
    # disable ingressroute permitInsecure field
    # disablePermitInsecure: false
    tls:
      # minimum TLS version that Sesame will negotiate
      # minimumProtocolVersion: "1.1"
    # The following config shows the defaults for the leader election.
    # leaderelection:
      # configmap-name: leader-elect
      # configmap-namespace: projectsesame
```

_Note:_ The default example `Sesame` includes this [file][1] for easy deployment of Sesame.

[1]: {{< param github_url >}}/tree/{{page.version}}/examples/Sesame/01-Sesame-config.yaml
