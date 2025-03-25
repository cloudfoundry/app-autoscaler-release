#!/bin/bash

set -euo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${script_dir}/vars.source.sh"

pushd "${bbl_state_path}"
  eval "$(bbl print-env)"
popd

cf api "https://api.${system_domain}" --skip-ssl-validation

cf_admin_password="$(credhub get -n /bosh-autoscaler/cf/cf_admin_password -q)"
cf auth admin "${cf_admin_password}"

set +e
existing_service_broker="$(cf curl v3/service_brokers | jq -r --arg service_broker_name "${deployment_name}" -r '.resources[] | select(.name == $service_broker_name) | .name')"
set -e

if [[ -n "$existing_service_broker" ]]; then
  echo "Service Broker ${existing_service_broker} already exists"
  echo " - cleaning up pr"
  pushd "${autoscaler_dir}/src/acceptance" > /dev/null && ./cleanup.sh && popd  > /dev/null
  echo " - deleting broker"
  cf delete-service-broker -f "${existing_service_broker}"
fi
set -x
echo "Creating service broker ${deployment_name} at 'https://${service_broker_name}.${system_domain}'"
autoscaler_service_broker_password=$(credhub get  -n "/bosh-autoscaler/${deployment_name}/service_broker_password" -q)
cf create-service-broker "${deployment_name}" autoscaler-broker-user "$autoscaler_service_broker_password" "https://${service_broker_name}.${system_domain}"

cf logout
