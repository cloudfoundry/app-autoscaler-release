#! /usr/bin/env bash
[ -n "${DEBUG}" ] && set -x
set -euo pipefail

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
workdir=${PWD}
${script_dir}/update_${type}_package.sh

cd ${workdir}
${script_dir}/create_pr.sh