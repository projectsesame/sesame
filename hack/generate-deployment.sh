#! /usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

readonly HERE=$(cd "$(dirname "$0")" && pwd)
readonly REPO=$(cd "${HERE}"/.. && pwd)
readonly PROGNAME=$(basename "$0")


readonly TARGET="${REPO}/examples/render/sesame.yaml"

exec > >(git stripspace >"$TARGET")

cat <<EOF
# This file is generated from the individual YAML files by $PROGNAME. Do not
# edit this file directly but instead edit the source files and re-render.
#
# Generated from:
EOF

(cd "${REPO}" && ls examples/sesame/*.yaml) | \
  awk '{printf "#       %s\n", $0}'

echo "#"
echo

# certgen uses the ':latest' image tag, so it always needs to be pulled. Everything
# else correctly uses versioned image tags so we should use IfNotPresent.
for y in "${REPO}/examples/sesame/"*.yaml ; do
    echo # Ensure we have at least one newline between joined fragments.
    case $y in
    */02-job-certgen.yaml)
        cat "$y"
        ;;
    *)
        sed 's/imagePullPolicy: Always/imagePullPolicy: IfNotPresent/g' < "$y"
        ;;
    esac
done

