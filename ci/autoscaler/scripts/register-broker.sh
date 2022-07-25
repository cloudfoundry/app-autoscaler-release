#!/bin/bash

set -euo pipefail

system_domain="${SYSTEM_DOMAIN:-autoscaler.ci.cloudfoundry.org}"
deployment_name="${DEPLOYMENT_NAME:-app-autoscaler}"
service_broker_name="${deployment_name}servicebroker"
bbl_state_path="${BBL_STATE_PATH:-bbl-state/bbl-state}"

pushd "${bbl_state_path}"
  eval "$(bbl print-env)"
popd

cf api "https://api.${system_domain}" --skip-ssl-validation

cf_admin_password=$(credhub get -n /bosh-autoscaler/cf/cf_admin_password -q)
cf auth admin "${cf_admin_password}"

set +e
existing_service_broker=$(cf service-brokers | grep "${service_broker_name}.${system_domain}" |  cut -d' ' -f1)
set -e

if [[ ! -z "$existing_service_broker" ]]; then
  echo "Service Broker ${existing_service_broker} already exists, assuming this is ok..."
  cf delete-service-broker ${existing_service_broker} -f
fi

echo "Creating service broker ${deployment_name} at 'https://${service_broker_name}.${system_domain}'"
autoscaler_service_broker_password=$(credhub get  -n "/bosh-autoscaler/${deployment_name}/autoscaler_service_broker_password" -q)
cf create-service-broker "${deployment_name}" autoscaler_service_broker_user "$autoscaler_service_broker_password" "https://${service_broker_name}.${system_domain}"

cf logout
