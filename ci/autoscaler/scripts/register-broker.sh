#!/bin/bash

set -euo pipefail

system_domain="${SYSTEM_DOMAIN:-autoscaler.ci.cloudfoundry.org}"
service_broker_name="${SERVICE_BROKER_NAME:-autoscalerservicebroker}"
service_name="${SERVICE_NAME:-autoscaler}"
deployment_name="${DEPLOYMENT_NAME:-app-autoscaler}"
bbl_state_path="${BBL_STATE_PATH:-bbl-state/bbl-state}"
VAR_DIR=bbl-state/bbl-state/vars

pushd ${bbl_state_path}
  eval "$(bbl print-env)"
popd

cf api https://api.${system_domain} --skip-ssl-validation

CF_ADMIN_PASSWORD=$(credhub get -n /bosh-autoscaler/cf/cf_admin_password -q)
cf auth admin $CF_ADMIN_PASSWORD

set +e
SERVICE_BROKER_EXISTS=$(cf service-brokers | grep -c ${service_broker_name}.${system_domain})
set -e
if [[ $SERVICE_BROKER_EXISTS == 1 ]]; then
  echo "Service Broker already exists, assuming this is ok..."
  #cf delete-service-broker -f autoscaler
else
  echo "Creating service broker..."
  AUTOSCALER_SERVICE_BROKER_PASSWORD=$(credhub get  -n /bosh-autoscaler/${deployment_name}/autoscaler_service_broker_password -q)
  cf create-service-broker ${service_name} autoscaler_service_broker_user $AUTOSCALER_SERVICE_BROKER_PASSWORD https://${service_broker_name}.${system_domain}
fi

cf logout
