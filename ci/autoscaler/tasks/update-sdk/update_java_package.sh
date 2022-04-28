#! /usr/bin/env bash

set -euo pipefail

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
source "${script_dir}/vendor_package.sh"

java_version=$(cat "${root_dir}/java-release/packages/openjdk-11/spec" | grep -e "- jdk-" | sed -E 's/- jdk-(.*)\.tar\.gz/\1/g')
cat ${java_version} > version

vendor-package java-release openjdk-11 "${java_version}"
