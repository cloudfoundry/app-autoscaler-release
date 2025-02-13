#! /usr/bin/env bash
[ -n "${DEBUG}" ] && set -x
set -euo pipefail

script_dir="$(cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd)"
source "${script_dir}/vars.source.sh"
source "${script_dir}/vendor_package.sh"

golang_dir=${GOLANG_DIR:-"${autoscaler_dir}/../golang-release"}
golang_dir="$(realpath -e "${golang_dir}")"
export golang_dir

golang_version=$(cat "${golang_dir}/packages/golang-1-linux/version")

step "updating go.mod files with golang version ${golang_version}"
find "${autoscaler_dir}" -name go.mod -type f -exec go mod edit -go "${golang_version}" "{}" \;

step "updating devbox with golang version ${golang_version}"
devbox add --config "${autoscaler_dir}" "go@${golang_version}"

echo -n "${golang_version}" > "${autoscaler_dir}/version"
vendor-package "${golang_dir}" golang-1-linux "${golang_version}"

