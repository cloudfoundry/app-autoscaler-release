#! /usr/bin/env bash
[ -n "${DEBUG}" ] && set -x
set -euo pipefail

script_dir="$(cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd)"
source "${script_dir}/vars.source.sh"
source "${script_dir}/vendor_package.sh"

golang_dir=${GOLANG_DIR:-"${autoscaler_dir}/../golang-release"}
golang_dir="$(realpath -e "${golang_dir}")"
export golang_dir

SED="sed"
which gsed >/dev/null && SED=gsed

golang_version=$(cat "${golang_dir}/packages/golang-1-linux/version")
golang_version="1.21.3"

step "updating go.mod files with golang version ${golang_version}"
find "${autoscaler_dir}" -name go.mod -type f -exec ${SED} -i "s/^[[:space:]]*go 1.*/go ${golang_version}/g" "{}" \;

step "updating go.work file with golang version ${golang_version}"
${SED} -i "s/^[[:space:]]*go 1.*/go ${golang_version}/g" "${autoscaler_dir}/go.work"

step "updating .tool-versions file with golang version ${golang_version}"
"${SED}" -i "s/golang 1.*/golang ${golang_version}/g" "${autoscaler_dir}/.tool-versions"

echo -n "${golang_version}" > "${autoscaler_dir}/version"
vendor-package "$golang_dir" golang-1-linux "${golang_version}"

