#! /usr/bin/env bash

set -euo pipefail
UPLOADER_KEY=${UPLOADER_KEY:-$( cat "${HOME}/.ssh/autoscaler_blobstore_uploader.key")}

function vendor-package {
  local release=${1}
  local package=${2}
  local version=${3}
  local package_location
  package_location=${root_dir}/${release}

  echo "# Building package for ${release} for version '${version}'"
  cat "${root_dir}/${release}/.git/ref" > "${root_dir}/vendored-commit"

  pushd "${autoscaler_dir}" > /dev/null
    # generate the private.yml file with the credentials
    echo "Generating private.yml..."
    cat > config/private.yml <<EOF
---
blobstore:
  options:
    credentials_source: static
    json_key:
EOF
    yq eval -i '.blobstore.options.json_key = strenv(UPLOADER_KEY)' config/private.yml

    bosh vendor-package "${package}" "${package_location}"
    cp "${root_dir}/vendored-commit" "packages/${package}/vendored-commit" && git add "packages/${package}/vendored-commit"
    echo -n "${version}" > "packages/${package}/version" && git add "packages/${package}/version"

    echo "# Git diff -----"
    git --no-pager diff
  popd >/dev/null
}
