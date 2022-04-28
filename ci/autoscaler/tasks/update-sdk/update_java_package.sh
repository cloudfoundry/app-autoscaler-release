#! /usr/bin/env bash

set -euo pipefail

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
source "${script_dir}/vendor_package.sh"

java_dir="${root_dir}/java-release"
java_version=$(cat "${java_dir}/packages/openjdk-11/spec" | grep -e "- jdk-" | sed -E 's/- jdk-(.*)\.tar\.gz/\1/g')

vendor-package java-release openjdk-11 "${java_version}"
