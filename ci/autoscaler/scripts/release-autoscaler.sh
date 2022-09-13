#!/usr/bin/env bash
[ -n "${DEBUG}" ] && set -x
set -euo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
root_dir=${ROOT_DIR:-"${script_dir}/../../../"}

mkdir -p 'generated-release'
previous_version=${PREV_VERSION:-$(cat gh-release/tag)}
generated=${DEST:-"$(realpath generated-release)"}
build_opts=${BUILD_OPTS:-"--final"}
VERSION=${VERSION:-$(cat "${generated}/name")}
PERFORM_BOSH_RELEASE=${PERFORM_BOSH_RELEASE:-"true"}
REPO_OUT=${REPO_OUT:-}

function create_release() {
   echo " - creating release"
   set -e
   local VERSION=$1
   local generated=$2
   bosh create-release \
        ${build_opts} \
        --version "$VERSION" \
        --tarball="app-autoscaler-v${VERSION}.tgz"

    RELEASE_TGZ="app-autoscaler-v${VERSION}.tgz"
    RELEASE_SHA256="$(sha256sum "${RELEASE_TGZ}" | head -n1 | awk '{print $1}')"
    mkdir -p "${generated}/artifacts"
    mv "app-autoscaler-v${VERSION}.tgz" "${generated}/artifacts/"
}

function create_tests() {
  echo " - creating acceptance test artifact"
  pushd "${root_dir}" > /dev/null
    make acceptance-release VERSION="${VERSION}" DEST="${generated}/artifacts/"
  popd > /dev/null
}

pushd "${root_dir}" > /dev/null
  # generate the private.yml file with the credentials
  cat > 'config/private.yml' <<EOF
---
blobstore:
  options:
    credentials_source: static
    json_key:
EOF
  echo 'Generating private.yml...'
  yq eval -i '.blobstore.options.json_key = strenv(UPLOADER_KEY)' config/private.yml


  LAST_COMMIT_SHA="$(git rev-parse HEAD)"
  echo "Generating release including commits up to: ${LAST_COMMIT_SHA}"
  pushd src/changelog > /dev/null
    echo " - running changelog"
    go run main.go \
      --changelog-file "${generated}/changelog.md" \
      --last-commit-sha-id "${LAST_COMMIT_SHA}"\
      --prev-rel-tag "${previous_version}"\
      --version-file "${generated}/name"
  popd

  export VERSION
  yq eval -i '.properties."autoscaler.apiserver.info.build".default = strenv(VERSION)' jobs/golangapiserver/spec

  echo "Displaying diff..."
  export GIT_PAGER=cat
  git diff

  if [ "${PERFORM_BOSH_RELEASE}" == "true" ]; then
    # FIXME these should be configurable variables
    if [[ -z $(git config --global user.email) ]]; then
      git config --global user.email "ci@cloudfoundry.org"
    fi

    # FIXME these should be configurable variables
    if [[ -z $(git config --global user.name) ]]; then
      git config --global user.name "CI Bot"
    fi

    git add jobs/golangapiserver/spec
    git commit -m "Updated release version to ${VERSION} in golangapiserver"

    create_release "${VERSION}" "${generated}"
    create_tests "${VERSION}" "${generated}"

    git add -A
    git status
    git commit -m "release v${VERSION}"
  else
    export RELEASE_SHA256="dummy-sha"
    export ACCEPTANCE_SHA256="dummy-sha"
  fi
  echo "${VERSION}" > "${generated}/tag"
  export ACCEPTANCE_SHA256=$(cat "${generated}/artifacts/"*.sha256)
  cat >> "${generated}/changelog.md" <<EOF

## Deployment

\`\`\`yaml
releases:
- name: app-autoscaler
  version: ${VERSION}
  url: https://storage.googleapis.com/app-autoscaler-releases/releases/app-autoscaler-v${VERSION}.tgz
  sha1: sha256:${RELEASE_SHA256}
- name: app-autoscaler-acceptance-tests
  version: ${VERSION}
  url: https://storage.googleapis.com/app-autoscaler-releases/releases/app-autoscaler-acceptance-tests-v${VERSION}.tgz
  sha1: sha256:${ACCEPTANCE_SHA256}
\`\`\`
EOF
  echo "---------- Changelog file ----------"
  cat "${generated}/changelog.md"
  echo "---------- end file ----------"
popd

[ -d "${REPO_OUT}" ] && cp -a app-autoscaler-release "${REPO_OUT}"
