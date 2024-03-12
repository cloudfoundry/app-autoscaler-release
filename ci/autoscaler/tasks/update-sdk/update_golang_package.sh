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
golang_version="1.21.3"

step "updating go.work file with golang version ${golang_version}"
go work edit -go "${golang_version}" "${autoscaler_dir}/go.work"

step "updating go.mod files with golang version ${golang_version}"
find "${autoscaler_dir}" -name go.mod -type f -exec go mod edit -go "${golang_version}" "{}" \;

step "updating .tool-versions file with golang version ${golang_version}"
sed -i "s/golang 1.*/golang ${golang_version}/g" "${autoscaler_dir}/.tool-versions"

echo -n "${golang_version}" > "${autoscaler_dir}/version"
vendor-package "${golang_dir}" golang-1-linux "${golang_version}"

