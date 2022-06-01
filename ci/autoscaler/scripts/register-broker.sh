#!/bin/bash

set -euo pipefail

bbl_state_path="${BBL_STATE_PATH:-bbl-state/bbl-state}"
VAR_DIR=bbl-state/bbl-state/vars

pushd ${bbl_state_path}
  eval "$(bbl print-env)"
popd

cf api https://api.${SYSTEM_DOMAIN} --skip-ssl-validation

CF_ADMIN_PASSWORD=$(credhub get -n /bosh-autoscaler/cf/cf_admin_password -q)
cf auth admin $CF_ADMIN_PASSWORD

set +e
SERVICE_BROKER_EXISTS=$(cf service-brokers | grep -c autoscalerservicebroker.${SYSTEM_DOMAIN})
set -e
if [[ $SERVICE_BROKER_EXISTS == 1 ]]; then
  echo "Service Broker already exists, assuming this is ok..."
  #cf delete-service-broker -f autoscaler
else
  echo "Creating service broker..."
  AUTOSCALER_SERVICE_BROKER_PASSWORD=$(credhub get  -n /bosh-autoscaler/app-autoscaler/autoscaler_service_broker_password -q)
  cf create-service-broker autoscaler autoscaler_service_broker_user $AUTOSCALER_SERVICE_BROKER_PASSWORD https://autoscalerservicebroker.${SYSTEM_DOMAIN}
fi

cf logout
