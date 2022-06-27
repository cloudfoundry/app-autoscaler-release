#!/bin/bash
set -euo pipefail
system_domain="${SYSTEM_DOMAIN:-autoscaler.ci.cloudfoundry.org}"
service_broker_name="${SERVICE_BROKER_NAME:-autoscalerservicebroker}"
service_name="${SERVICE_NAME:-autoscaler}"
deployment_name="${DEPLOYMENT_NAME:-app-autoscaler}"
autoscaler_root=${AUTOSCALER_DIR:-app-autoscaler-release}
bbl_state_path="${BBL_STATE_PATH:-bbl-state/bbl-state}"

cf api "https://api.${system_domain}" --skip-ssl-validation

pushd "${bbl_state_path}" > /dev/null
  eval "$(bbl print-env)"
popd > /dev/null

CF_ADMIN_PASSWORD=$(credhub get -n /bosh-autoscaler/cf/cf_admin_password -q)
cf auth admin "$CF_ADMIN_PASSWORD"

pushd "${autoscaler_root}/src/acceptance" > /dev/null
  ./cleanup.sh
popd  > /dev/null

SERVICE_BROKER_EXISTS=$(cf service-brokers | grep -c "${service_broker_name}.${system_domain}")
if [[ $SERVICE_BROKER_EXISTS == 1 ]]; then
  echo "Service Broker already exists, deleting..."
  cf delete-service-broker "${service_name}" -f
fi

bosh delete-deployment -d "${deployment_name}" -n

#set +e
#bosh delete-release "${deployment_name}" -n
#set -e

credhub delete -p "/bosh-autoscaler/${deployment_name}"