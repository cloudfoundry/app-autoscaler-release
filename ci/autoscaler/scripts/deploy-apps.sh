#! /usr/bin/env bash
# shellcheck disable=SC2086,SC2034,SC2155
set -eu -o pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${script_dir}/common.sh"
source "${script_dir}/vars.source.sh"

bosh_login "${BBL_STATE_PATH}"

function deploy() {
	log "Deploying autoscaler apps for bosh deployment '${deployment_name}' "
	pushd "${autoscaler_dir}/src/autoscaler" > /dev/null
		VERSION="0.0.0-rc.${PR_NUMBER:-0}" make mta-deploy
	popd > /dev/null
}


cf_login
cf_target "${autoscaler_org}" "${autoscaler_space}"
deploy
