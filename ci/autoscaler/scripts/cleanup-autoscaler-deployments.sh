#!/bin/bash

set -euo pipefail
script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${script_dir}/common.sh"

function get_autoscaler_deployments(){
  bosh deployments --json | jq -r '.Tables[0].Rows[] | select(.release_s | contains("app-autoscaler/")) | .name'
}

function main(){
  bosh_login
  cf_login
  step "Deployments to cleanup: $(get_autoscaler_deployments)"
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

[ "${BASH_SOURCE[0]}" == "${0}" ] && main "$@"
