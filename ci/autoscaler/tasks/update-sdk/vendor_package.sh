#! /usr/bin/env bash

set -exuo pipefail

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

function vendor-package {
  local release=${1}
  local package=${2}
  local autoscaler_location=${3}
  local package_location=${4}

  local tmpdir_name=$(mktemp -d)
  #trap "rm -rf ${tmpdir_name}" EXIT
  echo $tmpdir_name


  pushd "${package_location}" > /dev/null
    vendored_commit=$(cat ${package_location}.git/ref)
  popd

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
