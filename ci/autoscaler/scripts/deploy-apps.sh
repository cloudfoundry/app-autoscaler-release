#! /usr/bin/env bash
# shellcheck disable=SC2086,SC2034,SC2155
set -euo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${script_dir}/common.sh"
source "${script_dir}/vars.source.sh"

pushd "${bbl_state_path}" > /dev/null
  eval "$(bbl print-env)"
popd > /dev/null

function deploy() {
  log "Deploying autoscaler apps for bosh deployment '${deployment_name}' "
  pushd "${autoscaler_dir}/src/autoscaler" > /dev/null
    make mta-deploy
  popd > /dev/null
}

bosh_login
cf_login
cf_target "${autoscaler_org}" "${autoscaler_space}"
deploy
