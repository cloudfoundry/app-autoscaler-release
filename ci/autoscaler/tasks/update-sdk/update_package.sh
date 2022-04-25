#! /usr/bin/env bash

set -euo pipefail

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

${script_dir}/update_${type}_package.sh
${script_dir}/create_pr.sh