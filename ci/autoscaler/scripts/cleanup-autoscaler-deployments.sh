#!/bin/bash

set -euo pipefail
script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${script_dir}/common.sh"

function get_autoscaler_deployments(){
  bosh deployments --json --column=name | jq ".Tables[0].Rows[].name" -r  | grep -E "autoscaler|upgrade|performance"
}

function main(){
  bosh_login
  cf_login

  while IFS='' read -r deployment; do
    unset_vars
    export DEPLOYMENT_NAME="${deployment}"
    source "${script_dir}/vars.source.sh"

    cleanup_acceptance_run
    cleanup_service_broker
    cleanup_bosh_deployment
    cleanup_credhub
  done < <(get_autoscaler_deployments)

  cleanup_bosh
}


main
