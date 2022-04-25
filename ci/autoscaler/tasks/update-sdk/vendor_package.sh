#! /usr/bin/env bash

set -exuo pipefail

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

function vendor-package {
  local release=${1}
  local package=${2}
  local autoscaler_location=${3}

  local tmpdir_name=$(mktemp -d)
  trap "rm -rf ${tmpdir_name}" EXIT

  pushd "${tmpdir_name}" > /dev/null
    git clone --depth 1 "https://github.com/bosh-packages/${release}.git" .
    vendored_commit=$(git rev-parse HEAD)
  popd  > /dev/null

  pushd "${autoscaler_location}" > /dev/null
    bosh vendor-package "${package}" "${tmpdir_name}"
    echo "${vendored_commit}" >> "packages/${package}/vendored-commit"

    package_version_file="${tmpdir_name}/packages/${package}/version"
    if [[ -f "${package_version_file}" ]]; then
      cp "${package_version_file}" "packages/${package}"
    fi

    git --no-pager diff
  popd >/dev/null
}
