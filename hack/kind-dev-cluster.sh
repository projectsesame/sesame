#! /usr/bin/env bash

# kind-dev-cluster.sh: spin up a Sesame dev configuration in Kind
#
# This script starts a cluster in kind and deploys Sesame. We map
# the envoy listening ports to the host so that host traffic can
# easily be proxied. We deploy Sesame in insecure mode because we
# assume that the user will run a development Sesame locally on the
# host and set it as the Envoy xDS server.

readonly KIND=${KIND:-kind}
readonly KUBECTL=${KUBECTL:-kubectl}

readonly CLUSTER=${CLUSTER:-sesame}

readonly HERE=$(cd $(dirname $0) && pwd)
readonly REPO=$(cd ${HERE}/.. && pwd)

host::addresses() {
    case $(uname -s) in
    Darwin)
        networksetup -listallhardwareports | \
            awk '/Device/{print $2}' | \
            xargs -n1 ipconfig getifaddr
        ;;
    Linux)
        ip --json addr show up primary scope global primary permanent | \
            jq -r '.[].addr_info | .[] | select(.local) | .local'
        ;;
    *)
        echo 0.0.0.0
        ;;
    esac
}

kind::cluster::list() {
    ${KIND} get clusters
}

# Emit a Kind config that maps the envoy listener ports to the host.
kind::cluster::config() {
    cat <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
  extraPortMappings:
  - containerPort: 8080
    hostPort: 80
    listenAddress: "0.0.0.0"
  - containerPort: 8443
    hostPort: 443
    listenAddress: "0.0.0.0"
EOF
}

kind::cluster::create() {
    ${KIND} create cluster \
        --config <(kind::cluster::config) \
        --name ${CLUSTER}
}

kubectl::do() {
    ${KUBECTL} --context kind-${CLUSTER} "$@"
}

kubectl::apply() {
    kubectl::do apply -f "$@"
}

kind::cluster::create
kubectl::do get nodes

kubectl::apply ${REPO}/examples/sesame/00-common.yaml

kubectl::apply ${REPO}/examples/sesame/01-sesame-config.yaml
kubectl::apply ${REPO}/examples/sesame/01-crds.yaml
kubectl::apply ${REPO}/examples/sesame/02-rbac.yaml

# Skip 02-job-certgen.yaml, since we want to be running in
# insecure mode.

kubectl::apply ${REPO}/examples/sesame/02-service-sesame.yaml

# We don't need to create an envoy service, since kind has mapped
# the envoy ports to the host, so don't apply 02-service-envoy.yaml.

# We don't need to deploy sesame to the cluster because we expect
# the user to manually run a devel Sesame, so don't apply
# 03-sesame.yaml.

kubectl::apply ${REPO}/examples/sesame/03-envoy.yaml

# TODO(jpeach): figure out how to eliminate the manual CRD edits.
# Look into kustomize as an option.

cat <<EOF

Host IP address(es): $(host::addresses | tr '\n' ' ')
Next steps:

* Edit the envoy daemonset to remove the sesamecert and cacert secrets volume, and
  update the bootstrap container to point the xDS server to the host IP:

    ${KUBECTL} --context kind-${CLUSTER} --namespace projectsesame edit daemonset envoy

Run Sesame:

    Sesame serve --insecure --xds-address=0.0.0.0 --envoy-service-http-port=80 --envoy-service-https-port=443
EOF
