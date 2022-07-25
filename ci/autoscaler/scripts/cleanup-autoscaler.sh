#!/bin/bash
set -euo pipefail
system_domain="${SYSTEM_DOMAIN:-autoscaler.ci.cloudfoundry.org}"
service_name="${SERVICE_NAME:-autoscaler}"
deployment_name="${DEPLOYMENT_NAME:-app-autoscaler}"
service_broker_name="${deployment_name}servicebroker"
autoscaler_root=${AUTOSCALER_DIR:-app-autoscaler-release}
bbl_state_path="${BBL_STATE_PATH:-bbl-state/bbl-state}"
RELEASE_SHA=${RELEASE_SHA:-""}
CURRENT_COMMIT_HASH=${CURRENT_COMMIT_HASH:-${RELEASE_SHA}}

cf api "https://api.${system_domain}" --skip-ssl-validation

pushd "${bbl_state_path}" > /dev/null
  eval "$(bbl print-env)"
popd > /dev/null

CF_ADMIN_PASSWORD=$(credhub get -n /bosh-autoscaler/cf/cf_admin_password -q)
cf auth admin "$CF_ADMIN_PASSWORD"

echo "# Cleaning up from acceptance tests"
pushd "${autoscaler_root}/src/acceptance" > /dev/null
  ./cleanup.sh
popd  > /dev/null

echo "# Cleaning up from Bosh deployments"
SERVICE_BROKER_EXISTS=$(cf service-brokers | grep -c "${service_broker_name}.${system_domain}" || true)
if [[ $SERVICE_BROKER_EXISTS == 1 ]]; then
  echo "- Service Broker exists, deleting broker '${service_name}'"
  cf delete-service-broker "${service_name}" -f
fi

echo "- Deleting bosh deployment '${deployment_name}'"
bosh delete-deployment -d "${deployment_name}" -n

if [ -n "${deployment_name}" ]
then
  for release in $(bosh releases | grep "${deployment_name} " | awk '{print $2}')
  do
     echo "- Deleting bosh release '${release}'"
     bosh delete-release -n "app-autoscaler/${release}"
  done
fi

echo "- Deleting credhub creds: '/bosh-autoscaler/${deployment_name}/*'"
credhub delete -p "/bosh-autoscaler/${deployment_name}"
