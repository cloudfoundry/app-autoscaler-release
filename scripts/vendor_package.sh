#! /usr/bin/env bash

set -euo pipefail

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

function vendor-package {
  local release=${1}
  local package=${2}

  local tmpdir_name="/tmp/${release}-${RANDOM}"
  mkdir -p "${tmpdir_name}"
  trap "rm -r -f ${tmpdir_name}" EXIT

  pushd "${tmpdir_name}"
    git clone --depth 1 "https://github.com/bosh-packages/${release}.git" .
    vendored_commit=$(git rev-parse HEAD)
  popd

  bosh vendor-package "${package}" "${tmpdir_name}"
  echo "${vendored_commit}" >> "${script_dir}/../packages/${package}/vendored-commit"

  package_version_file="${tmpdir_name}/packages/${package}/version"
  if [[ -f "${package_version_file}" ]]; then
    cp "${package_version_file}" "${script_dir}/../packages/${package}"
  fi
}
