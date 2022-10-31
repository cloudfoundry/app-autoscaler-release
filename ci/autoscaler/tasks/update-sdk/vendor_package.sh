#! /usr/bin/env bash

set -euo pipefail


function vendor-package {
  local release=${1}
  local package=${2}
  local version=${3}
  local package_location
  package_location=${release}
  config_file=config/private.yml
  log "Building package for ${release} for version '${version}'"
  pushd "${root_dir}/${release}" > /dev/null && git rev-parse HEAD > "${root_dir}/vendored-commit" && popd > /dev/null

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

  bosh vendor-package "${package}" "${package_location}"
  cp "${root_dir}/vendored-commit" "packages/${package}/vendored-commit" && git add "packages/${package}/vendored-commit"
  echo -n "${version}" > "packages/${package}/version" && git add "packages/${package}/version"

  log "Git diff -----"
  git --no-pager diff
}
