#! /usr/bin/env bash

# make-release-tag.sh: This script assumes that you are on a branch and have
# otherwise prepared the release. It rewrites the Docker image version in
# the deployment YAML, then created a tag with a message containing the
# shortlog from the previous version.

readonly PROGNAME=$(basename "$0")
readonly OLDVERS="$1"
readonly NEWVERS="$2"

if [ -z "$OLDVERS" ] || [ -z "$NEWVERS" ]; then
    printf "Usage: %s OLDVERS NEWVERS\n" "$PROGNAME"
    exit 1
fi

set -o errexit
set -o nounset
set -o pipefail

readonly IMG="ghcr.io/projectsesame/sesame:$NEWVERS"

if [ -n "$(git tag --list "$NEWVERS")" ]; then
    printf "%s: tag '%s' already exists\n" "$PROGNAME" "$NEWVERS"
    exit 1
fi

# Wrap sed to deal with GNU and BSD sed flags.
run::sed() {
    local -r vers="$(sed --version < /dev/null 2>&1 | grep -q GNU && echo gnu || echo bsd)"
    case "$vers" in
        gnu) sed -i "$@" ;;
        *) sed -i '' "$@" ;;
    esac
}

# Update the image tags in the Sesame, Envoy and certgen manifests to the new version.
for example in examples/sesame/03-envoy.yaml examples/sesame/03-sesame.yaml examples/sesame/02-job-certgen.yaml ; do
    # The version might be main or OLDVERS depending on whether we are
    # tagging from the release branch or from main.
    run::sed \
        "-es|ghcr.io/projectsesame/sesame:main|$IMG|" \
        "-es|ghcr.io/projectsesame/sesame:$OLDVERS|$IMG|" \
        "$example"
done

# Update the certgen job name to the new version.
for example in examples/sesame/02-job-certgen.yaml ; do
    run::sed \
        "-es|sesame-certgen-main|sesame-certgen-$NEWVERS|" \
        "-es|sesame-certgen-$OLDVERS|sesame-certgen-$NEWVERS|" \
        "$example"
done

# Remove spec.ttlSecondsAfterFinished from the certgen job, as it is versioned
# for releases and doesn't need to be cleaned up.
for example in examples/sesame/02-job-certgen.yaml ; do
    run::sed \
        '-e/^[[:blank:]]*ttlSecondsAfterFinished:/d' \
        "$example"
done

make generate

# If pushing the tag failed, then we might have already committed the
# YAML updates. The "git commit" will fail if there are no changes, so
# make sure that there are changes to commit before we do it.
if git status -s examples/sesame 2>&1 | grep -E -q '^\s+[MADRCU]'; then
    git commit -s -m "Update Sesame Docker image to $NEWVERS." \
        examples/sesame/03-sesame.yaml \
        examples/sesame/03-envoy.yaml \
        examples/sesame/02-job-certgen.yaml \
        examples/render/sesame.yaml \
        examples/render/sesame-gateway.yaml
fi

git tag -F - "$NEWVERS" <<EOF
Tag $NEWVERS release.

$(git shortlog "$OLDVERS..HEAD")
EOF

printf "Created tag '%s'\n" "$NEWVERS"

# People set up their remotes in different ways, so we need to check
# which one to dry run against. Choose a remote name that pushes to the
# projectsesame org repository (i.e. not the user's Github fork).
readonly REMOTE=$(git remote -v | awk '$2~/projectsesame\/sesame/ && $3=="(push)" {print $1}' | head -n 1)
if [ -z "$REMOTE" ]; then
    printf "%s: unable to determine remote for %s\n" "$PROGNAME" "projectsesame/sesame"
    exit 1
fi

readonly BRANCH=$(git branch --show-current)

printf "Testing whether commit can be pushed\n"
git push --dry-run "$REMOTE" "$BRANCH"

printf "Testing whether tag '%s' can be pushed\n" "$NEWVERS"
git push --dry-run "$REMOTE" "$NEWVERS"

printf "Run 'git push %s %s' to push the commit and then 'git push %s %s' to push the tag if you are happy\n" "$REMOTE" "$BRANCH" "$REMOTE" "$NEWVERS"
