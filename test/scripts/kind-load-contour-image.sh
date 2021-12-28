#! /usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

readonly KIND=${KIND:-kind}

readonly LOAD_PREBUILT_IMAGE=${LOAD_PREBUILT_IMAGE:-false}
readonly CLUSTERNAME=${CLUSTERNAME:-sesame-e2e}

readonly HERE=$(cd $(dirname $0) && pwd)
readonly REPO=$(cd ${HERE}/../.. && pwd)

kind::cluster::exists() {
    ${KIND} get clusters | grep -q "$1"
}

kind::cluster::load::archive() {
    ${KIND} load image-archive \
        --name "${CLUSTERNAME}" \
        "$@"
}

kind::cluster::load::docker() {
    ${KIND} load docker-image \
        --name "${CLUSTERNAME}" \
        "$@"
}

if ! kind::cluster::exists "$CLUSTERNAME" ; then
    echo "cluster $CLUSTERNAME does not exist"
    exit 2
fi

if [ "${LOAD_PREBUILT_IMAGE}" = "true" ]; then
    kind::cluster::load::archive "$(ls ${REPO}/image/sesame-*.tar)"
else
    # Build the current version of Sesame.
    VERSION="v$$"
    make -C ${REPO} container IMAGE=ghcr.io/projectsesame/sesame VERSION=$VERSION

    # Also tag as main since test suites will use this tag unless overridden.
    docker tag ghcr.io/projectsesame/sesame:${VERSION} ghcr.io/projectsesame/sesame:main

    # Push the Sesame build image into the cluster.
    kind::cluster::load::docker ghcr.io/projectsesame/sesame:${VERSION}
    kind::cluster::load::docker ghcr.io/projectsesame/sesame:main
fi
