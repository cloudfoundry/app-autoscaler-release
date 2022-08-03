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

if [[ -n "$existing_service_broker" ]]; then
  echo "Service Broker ${existing_service_broker} already exists"
  cf delete-service-broker -f ${existing_service_broker}
else
  echo "Creating service broker ${deployment_name} at 'https://${service_broker_name}.${system_domain}'"
  autoscaler_service_broker_password=$(credhub get  -n "/bosh-autoscaler/${deployment_name}/service_broker_password" -q)
  cf create-service-broker "${depoyment_name}" autoscaler-broker-user "$autoscaler_service_broker_password" "https://${service_broker_name}.${system_domain}"
fi

cf logout
