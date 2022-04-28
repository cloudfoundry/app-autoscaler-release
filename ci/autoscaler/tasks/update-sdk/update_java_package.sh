#! /usr/bin/env bash

set -euo pipefail

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

source "${script_dir}/vendor_package.sh"
vendor-package java-release openjdk-11 "$(java_version)"
