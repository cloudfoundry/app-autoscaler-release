#!/bin/bash

set -euo pipefail
script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${script_dir}/common.sh"

function get_autoscaler_deployments(){
  cf service-brokers | cut -d' ' -f1 |grep autoscaler
}

function main(){
  bosh_login
  cf_login

  deployments=($(get_autoscaler_deployments))
  for deployment in "${deployments[@]}" ; do :
    export deployment_name="${deployment}"
    export name_prefix="${deployment_name}-TESTS"

    cleanup_organization
    cleanup_service_broker
    cleanup_bosh_deployment
    cleanup_credhub
  done
  cleanup_bosh
}


main
