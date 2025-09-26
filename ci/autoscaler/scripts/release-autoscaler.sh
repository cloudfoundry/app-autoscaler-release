#! /usr/bin/env bash

# NOTE: you can run this locally for testing !!!
#
# ./script/local_release_autoscaler.sh


[ -n "${DEBUG}" ] && set -x

set -euo pipefail
script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${script_dir}/vars.source.sh"

previous_version=${PREV_VERSION:-$(cat gh-release/tag)}
mkdir -p "build"
build_path=$(realpath build)
build_opts=${BUILD_OPTS:-"--final"}
mkdir -p "keys"
keys_path="$(realpath keys)"
PERFORM_BOSH_RELEASE=${PERFORM_BOSH_RELEASE:-"true"}
REPO_OUT=${REPO_OUT:-}
export UPLOADER_KEY=${UPLOADER_KEY:-"NOT_SET"}
CI=${CI:-false}
SUM_FILE="${build_path}/artifacts/files.sum.sha256"

function create_release() {
   set -e
   mkdir -p "${build_path}/artifacts"
   local version=$1
   local build_path=$2
   local release_file=$3
   echo " - building new release from ${PWD} at revision $(git rev-parse HEAD)"
   echo " - creating release '${version}' in '${build_path}' as ${release_file}"

   yq eval -i ".properties.\"autoscaler.apiserver.info.build\".default = \"${version}\"" jobs/golangapiserver/spec
   yq eval -i ".build = \"${version}\"" src/autoscaler/api/default_info.json

   git add jobs/golangapiserver/spec
   git add src/autoscaler/api/default_info.json

   [ "${CI}" = "true" ] && git commit -S -m "Updated release version to ${version} in golangapiserver"

   # shellcheck disable=SC2086
   bosh create-release \
        ${build_opts} \
        --version "${version}" \
        --tarball="${build_path}/artifacts/${release_file}"
}

function create_mtar() {
  set -e
  mkdir -p "${build_path}/artifacts"
  local version=$1
  local build_path=$2
  echo " - creating autorscaler mtar artifact"
  pushd "${autoscaler_dir}" > /dev/null
    make mta-release VERSION="${version}" DEST="${build_path}/artifacts/"
  popd > /dev/null
}

function create_tests() {
  set -e
  mkdir -p "${build_path}/artifacts"
  local version=$1
  local build_path=$2
  echo " - creating acceptance test artifact"
  pushd "${autoscaler_dir}" > /dev/null
    make acceptance-release VERSION="${version}" DEST="${build_path}/artifacts/"
  popd > /dev/null
}

function commit_release(){
  pushd "${autoscaler_dir}"
  git add -A
  git status
  git commit -S -m "created release v${VERSION}"
}

function create_bosh_config(){
   # generate the private.yml file with the credentials
   config_file="${autoscaler_dir}/config/private.yml"
    cat > "$config_file" <<EOF
---
blobstore:
  options:
    credentials_source: static
    json_key:
EOF
    echo ' - Generating private.yml...'
    yq eval -i '.blobstore.options.json_key = strenv(UPLOADER_KEY)' "$config_file"
}

function setup_git(){
  if [[ -z $(git config --global user.email) ]]; then
    git config --global user.email "${AUTOSCALER_CI_BOT_EMAIL}"
  fi

  if [[ -z $(git config --global user.name) ]]; then
    git config --global user.name "${AUTOSCALER_CI_BOT_NAME}"
  fi

  public_key_path="${keys_path}/autoscaler-ci-bot-signing-key.pub"
  private_key_path="${keys_path}/autoscaler-ci-bot-signing-key"
  echo "$AUTOSCALER_CI_BOT_SIGNING_KEY_PUBLIC" > "${public_key_path}"
  echo "$AUTOSCALER_CI_BOT_SIGNING_KEY_PRIVATE" > "${private_key_path}"
  chmod 600 "${public_key_path}"
  chmod 600 "${private_key_path}"

  git config --global gpg.format ssh
  git config --global user.signingkey "${private_key_path}"
}


pushd "${autoscaler_dir}" > /dev/null
  setup_git
  create_bosh_config

  echo " - Displaying diff..."
  export GIT_PAGER=cat
  git diff

  VERSION=${VERSION:-$(cat "${build_path}/name")}
  echo "v${VERSION}" > "${build_path}/tag"
  if [ "${PERFORM_BOSH_RELEASE}" == "true" ]; then
    RELEASE_TGZ="app-autoscaler-v${VERSION}.tgz"
    ACCEPTANCE_TEST_TGZ="app-autoscaler-acceptance-tests-v${VERSION}.tgz"
    AUTOSCALER_MTAR="app-autoscaler-release-v${VERSION}.mtar"
    create_release "${VERSION}" "${build_path}" "${RELEASE_TGZ}"
    create_tests "${VERSION}" "${build_path}"
    create_mtar "${VERSION}" "${build_path}"
    [ "${CI}" = "true" ] && commit_release

    sha256sum "${build_path}/artifacts/"* > "${build_path}/artifacts/files.sum.sha256"
    ACCEPTANCE_SHA256=$( grep "${ACCEPTANCE_TEST_TGZ}$" "${SUM_FILE}" | awk '{print $1}' )
    RELEASE_SHA256=$( grep "${RELEASE_TGZ}$" "${SUM_FILE}" | awk '{print $1}')
    MTAR_SHA256=$( grep "${AUTOSCALER_MTAR}$" "${SUM_FILE}" | awk '{print $1}')
  else
    ACCEPTANCE_SHA256="dummy-sha"
    RELEASE_SHA256="dummy-sha"
    MTAR_SHA256="dummy-sha"
  fi
  export ACCEPTANCE_SHA256
  export RELEASE_SHA256
	export MTAR_SHA256

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
- name: app-autoscaler-mtar
  version: ${VERSION}
  url: https://storage.googleapis.com/app-autoscaler-releases/releases/app-autoscaler-release-v${VERSION}.mtar
  sha1: sha256:${MTAR_SHA256}
\`\`\`
EOF
  echo "---------- Changelog file ----------"
  cat "${build_path}/changelog.md"
  echo "---------- end file ----------"

popd > /dev/null
echo " - Completed"
