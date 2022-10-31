#! /usr/bin/env bash
[ -n "${DEBUG}" ] && set -x
set -euo pipefail

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
# shellcheck source=vendor_package.sh
source "${script_dir}/vars.source.sh"
source "${script_dir}/vendor_package.sh"
export java_dir=${JAVA_DIR:-$( realpath -e "${autoscaler_dir}/../java-release")}

# shellcheck disable=SC2154
java_version=$(grep "${java_dir}/packages/openjdk-17/spec" -e "- jdk-" | sed -E 's/- jdk-(.*)\.tar\.gz/\1/g')
echo -n "${java_version}" > version

vendor-package java-release openjdk-17 "${java_version}"
