#! /usr/bin/env bash
[ -n "${DEBUG}" ] && set -x
set -euo pipefail

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
source "${script_dir}/vars.source.sh"
create_pr=${CREATE_PR:-"false"}

# shellcheck disable=SC2154
"${script_dir}"/update_"${type}"_package.sh
[[ ${create_pr} == "true" ]] && "${script_dir}"/create_pr.sh