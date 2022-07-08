#! /usr/bin/env bash
[ -n "${DEBUG}" ] && set -x
set -euo pipefail

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
# shellcheck source=vendor_package.sh
source "${script_dir}/vendor_package.sh"

# shellcheck disable=SC2154
java_version=$(grep "${root_dir}/java-release/packages/openjdk-11/spec" -e "- jdk-" | sed -E 's/- jdk-(.*)\.tar\.gz/\1/g')
echo -n "${java_version}" > version

vendor-package java-release openjdk-11 "${java_version}"
