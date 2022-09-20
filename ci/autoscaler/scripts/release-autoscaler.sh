#!/usr/bin/env bash
# NOTE: you can run this locally for testing !!!
# you need a github token (GITHUB_TOKEN) and beware that it adds a commit you need to drop each time also you need to remove dev_releases from root.
# GITHUB_TOKEN=ghp_[your token]   DEST=${PWD}/../../../build VERSION="8.0.0" BUILD_OPTS="--force" PREV_VERSION=6.0.0  ./release-autoscaler.sh

[ -n "${DEBUG}" ] && set -x
set -euo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
root_dir=$(realpath "${ROOT_DIR:-"${script_dir}/../../../"}" )

previous_version=${PREV_VERSION:-$(cat gh-release/tag)}
mkdir -p 'build'
build_path=$(realpath build)
build_opts=${BUILD_OPTS:-"--final"}
PERFORM_BOSH_RELEASE=${PERFORM_BOSH_RELEASE:-"true"}
REPO_OUT=${REPO_OUT:-}
export UPLOADER_KEY=${UPLOADER_KEY:-"NOT_SET"}
CI=${CI:-false}

RELEASE_TGZ="app-autoscaler-v${VERSION}.tgz"
ACCEPTANCE_TEST_TGZ="app-autoscaler-acceptance-tests-v${VERSION}.tgz"
SUM_FILE="${build_path}/artifacts/files.sum.sha256"
function create_release() {
   mkdir -p "${build_path}/artifacts"
   set -e
   local VERSION=$1
   local build_path=$2
   echo " - creating release '${VERSION}' in '${build_path}'"
   yq eval -i '.properties."autoscaler.apiserver.info.build".default = strenv(VERSION)' jobs/golangapiserver/spec
   # shellcheck disable=SC2086
   bosh create-release \
        ${build_opts} \
        --version "$VERSION" \
        --tarball="${build_path}/artifacts/${RELEASE_TGZ}"
}

function create_tests() {
  echo " - creating acceptance test artifact"
  pushd "${root_dir}" > /dev/null
    make acceptance-release VERSION="${VERSION}" DEST="${build_path}/artifacts/"
  popd > /dev/null
}

function commit_release(){
  # FIXME these should be configurable variables
  if [[ -z $(git config --global user.email) ]]; then
    git config --global user.email "ci@cloudfoundry.org"
  fi

  # FIXME these should be configurable variables
  if [[ -z $(git config --global user.name) ]]; then
    git config --global user.name "CI Bot"
  fi

  pushd "${root_dir}"
  git add -A
  git status
  git commit -m "created release v${VERSION}"
}

function create_bosh_config(){
   # generate the private.yml file with the credentials
   config_file="${root_dir}/config/private.yml"
    cat > "$config_file" <<EOF
---
blobstore:
  options:
    credentials_source: static
    json_key:
EOF
    echo 'Generating private.yml...'
    yq eval -i '.blobstore.options.json_key = strenv(UPLOADER_KEY)' "$config_file"
}

function generate_changelog(){
  [ -e "${build_path}/changelog.md" ] && return
  LAST_COMMIT_SHA="$(git rev-parse HEAD)"
  echo "Generating release notes including commits up to: ${LAST_COMMIT_SHA}"
  pushd src/changelog > /dev/null
    echo " - running changelog"
    go run main.go \
      --changelog-file "${build_path}/changelog.md" \
      --last-commit-sha-id "${LAST_COMMIT_SHA}"\
      --prev-rel-tag "${previous_version}"\
      --version-file "${build_path}/name"
  popd
}

pushd "${root_dir}" > /dev/null
  create_bosh_config
  generate_changelog

  echo "Displaying diff..."
  export GIT_PAGER=cat
  git diff

  VERSION=${VERSION:-$(cat "${build_path}/name")}
  echo "${VERSION}" > "${build_path}/tag"

  if [ "${PERFORM_BOSH_RELEASE}" == "true" ]; then
    create_release "${VERSION}" "${build_path}"
    create_tests "${VERSION}" "${build_path}"
    [ "${CI}" = "true" ] && commit_release

    sha256sum "${build_path}/artifacts/"* > "${build_path}/artifacts/files.sum.sha256"
    ACCEPTANCE_SHA256=$( grep "${ACCEPTANCE_TEST_TGZ}" "${SUM_FILE}" | awk '{print $1}' )
    RELEASE_SHA256=$( grep "${RELEASE_TGZ}" "${SUM_FILE}" | awk '{print $1}')
  else
    ACCEPTANCE_SHA256="dummy-sha"
    RELEASE_SHA256="dummy-sha"
  fi
  export ACCEPTANCE_SHA256
  export RELEASE_SHA256

  cat >> "${build_path}/changelog.md" <<EOF

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
  cat "${build_path}/changelog.md"
  echo "---------- end file ----------"
popd

[ -d "${REPO_OUT}" ] && cp -a app-autoscaler-release "${REPO_OUT}"
