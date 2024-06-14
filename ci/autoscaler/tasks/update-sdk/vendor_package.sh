#! /usr/bin/env bash

set -euo pipefail
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
source "${script_dir}/vars.source.sh"

function write_vendor_commit(){
  pushd "${1}" > /dev/null
    git rev-parse HEAD > "${autoscaler_dir}/vendored-commit"
  popd > /dev/null
}

function vendor-package {
  local release=${1}
  local package=${2}
  local version=${3}
  local package_location
  package_location=${release}
  config_file="${autoscaler_dir}/config/private.yml"
  log "Building package for ${release} for version '${version}'"
  write_vendor_commit "${release}"

    # generate the private.yml file with the credentials
  step "Generating private.yml..."
  cat > "${config_file}" <<EOF
---
blobstore:
  options:
    credentials_source: static
    json_key:
EOF
  export UPLOADER_KEY=${UPLOADER_KEY:-$(cat "${HOME}/.ssh/autoscaler_blobstore_uploader.key")}
  yq eval -i '.blobstore.options.json_key = strenv(UPLOADER_KEY)' "${config_file}"

  pushd "${autoscaler_dir}" > /dev/null
    step "vendoring package ${package}"
    bosh vendor-package "${package}" "${package_location}"

    vendor_commit_file="${autoscaler_dir}/packages/${package}/vendored-commit"
    version_commit_file="${autoscaler_dir}/packages/${package}/version"
    cp "${autoscaler_dir}/vendored-commit" "${vendor_commit_file}" && git add "${vendor_commit_file}"
    cp "${autoscaler_dir}/version" "${version_commit_file}" && git add "${version_commit_file}"

    log "Git diff -----"
    git --no-pager diff
  popd > /dev/null
}
