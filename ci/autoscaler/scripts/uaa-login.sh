#!/bin/bash
# shellcheck disable=SC2086
set -euo pipefail
script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${script_dir}/common.sh"
source "${script_dir}/vars.source.sh"

bosh_login "${BBL_STATE_PATH}"
uaa_login
