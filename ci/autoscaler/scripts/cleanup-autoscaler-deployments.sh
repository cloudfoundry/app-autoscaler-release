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
  set -x
  for deployment in "${deployments[@]}" ; do :
    export DEPLOYMENT_NAME="${deployment}"
    export NAME_PREFIX="${DEPLOYMENT_NAME}-TESTS"

    cleanup_organization
    cleanup_service_broker
    cleanup_bosh_deployment
    cleanup_credhub
  done
  set +x
  cleanup_bosh
}


main
