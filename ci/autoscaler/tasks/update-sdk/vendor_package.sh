#! /usr/bin/env bash

set -exuo pipefail

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
export GOOGLE_APPLICATION_CREDENTIALS="${UPLOADER_KEY}"

function vendor-package {
  local release=${1}
  local package=${2}
  local autoscaler_location=${3}
  local package_location="$(pwd)/${release}"


  vendored_commit=$(cat "${release}/.git/ref")


  pushd "${autoscaler_location}" > /dev/null
    bosh vendor-package "${package}" "${package_location}"
    echo "${vendored_commit}" >> "packages/${package}/vendored-commit"

    package_version_file="${package_location}/packages/${package}/version"
    if [[ -f "${package_version_file}" ]]; then
      cp "${package_version_file}" "packages/${package}"
    fi

    git --no-pager diff
  popd >/dev/null
}
