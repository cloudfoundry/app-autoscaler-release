#!/bin/bash
# shellcheck disable=SC2086
set -euo pipefail
script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${script_dir}/common.sh"

bosh_login
cf_login

cf autoscaling-api "https://autoscaler-${PR_NUMBER}.${SYSTEM_DOMAIN}"
