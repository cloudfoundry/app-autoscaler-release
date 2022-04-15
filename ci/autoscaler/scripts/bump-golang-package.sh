#!/bin/bash

set -euo pipefail

function configure_git_credentials(){
  if [[ -z $(git config --global user.email) ]]; then
    git config --global user.email "$GIT_USER_EMAIL"
  fi

  if [[ -z $(git config --global user.name) ]]; then
      git config --global user.name "$GIT_USER_NAME"
  fi
}

# FIXME: Remove this inline function
function vendor-package {
  local release=${1}
  local package=${2}

  local tmpdir_name="$PWD/${release}-${RANDOM}"
  mkdir -p ${tmpdir_name}
  trap "rm -r -f ${tmpdir_name}" EXIT

  pushd ${tmpdir_name}
    git clone --depth 1 "https://github.com/bosh-packages/${release}.git" .
  popd
  bosh vendor-package ${package} ${tmpdir_name}
}

configure_git_credentials
git clone app-autoscaler-release updated-app-autoscaler-release

pushd app-autoscaler-release >/dev/null


  # FIXME final.yml is used for testing purpose.
  # add private.YAML to upload golang-release to GCS
  cat > config/final.yml <<EOF
---
blobstore:
  provider: local
  options:
    blobstore_path: ""
EOF
  echo "Generating private.yml..."
  yq eval -i '.blobstore.options.blobstore_path = ${PWD}' config/final.yml


  #FIXME use vendor-package from ./scripts/update_vendor_packages
  vendor-package golang-release golang-1.18-linux
  vendor-package java-release openjdk-11

  git status

popd


