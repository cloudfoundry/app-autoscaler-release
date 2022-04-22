#! /usr/bin/env bash

set -euo pipefail

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

source script_dir/vendor_package.sh

vendor-package golang-release golang-1-linux app-autoscaler-release

golang_version=$( cat "${script_dir}/../packages/golang-1-linux/version")
stripped_go_version=$(echo "${golang_version}" | cut -d . -f -2)

sed -i '' "s/^[[:space:]]*go 1.*/go $stripped_go_version/g" "${script_dir}/../src/acceptance/go.mod"
sed -i '' "s/^[[:space:]]*go 1.*/go $stripped_go_version/g" "${script_dir}/../src/autoscaler/go.mod"
sed -i '' "s/^[[:space:]]*go 1.*/go $stripped_go_version/g" "${script_dir}/../src/changelog/go.mod"
sed -i '' "s/^[[:space:]]*go 1.*/go $stripped_go_version/g" "${script_dir}/../src/changeloglockcleaner/go.mod"

sed -i '' "s/golang 1.*/golang $golang_version/g" "${script_dir}/../.tool-versions"
