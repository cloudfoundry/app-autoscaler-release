#!/bin/bash

set -euo pipefail
script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${script_dir}/common.sh"

function main(){
  bosh_login
  cf_login
  cleanup_acceptance_run
  cleanup_service_broker
  cleanup_bosh_deployment
  cleanup_credhub
}

[ "${BASH_SOURCE[0]}" == "${0}" ] && main "$@"
