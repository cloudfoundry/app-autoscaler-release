#!/bin/bash

set -euo pipefail

cf api https://api.${SYSTEM_DOMAIN} --skip-ssl-validation

pushd bbl-state/bbl-state
  eval "$(bbl print-env)"
popd

CF_ADMIN_PASSWORD=$(credhub get -n /bosh-autoscaler/cf/cf_admin_password -q)
cf auth admin $CF_ADMIN_PASSWORD

set +e
pushd app-autoscaler-release/src/acceptance
  ./cleanup.sh
popd
set -e

set +e
SERVICE_BROKER_EXISTS=$(cf service-brokers | grep -c autoscalerservicebroker.${SYSTEM_DOMAIN}) 
set -e
if [[ $SERVICE_BROKER_EXISTS == 1 ]]; then
  echo "Service Broker already exists, deleting..."
  cf delete-service-broker autoscaler -f
fi

set +e
bosh delete-deployment -d app-autoscaler -n
set -e

set +e
bosh delete-release app-autoscaler -n
set -e

set +e
credhub delete -n /bosh-autoscaler/app-autoscaler/postgres_server
credhub delete -n /bosh-autoscaler/app-autoscaler/postgres_ca
credhub delete -n /bosh-autoscaler/app-autoscaler/metricsserver_client
credhub delete -n /bosh-autoscaler/app-autoscaler/metricsserver_server
credhub delete -n /bosh-autoscaler/app-autoscaler/metricsserver_ca
credhub delete -n /bosh-autoscaler/app-autoscaler/scheduler_client
credhub delete -n /bosh-autoscaler/app-autoscaler/scheduler_server
credhub delete -n /bosh-autoscaler/app-autoscaler/scheduler_ca
credhub delete -n /bosh-autoscaler/app-autoscaler/servicebroker_public_server
credhub delete -n /bosh-autoscaler/app-autoscaler/servicebroker_public_ca
credhub delete -n /bosh-autoscaler/app-autoscaler/servicebroker_client
credhub delete -n /bosh-autoscaler/app-autoscaler/servicebroker_server
credhub delete -n /bosh-autoscaler/app-autoscaler/servicebroker_ca
credhub delete -n /bosh-autoscaler/app-autoscaler/apiserver_client
credhub delete -n /bosh-autoscaler/app-autoscaler/apiserver_public_server
credhub delete -n /bosh-autoscaler/app-autoscaler/apiserver_public_ca
credhub delete -n /bosh-autoscaler/app-autoscaler/apiserver_server
credhub delete -n /bosh-autoscaler/app-autoscaler/apiserver_ca
credhub delete -n /bosh-autoscaler/app-autoscaler/eventgenerator_client
credhub delete -n /bosh-autoscaler/app-autoscaler/eventgenerator_server
credhub delete -n /bosh-autoscaler/app-autoscaler/eventgenerator_ca
credhub delete -n /bosh-autoscaler/app-autoscaler/scalingengine_client
credhub delete -n /bosh-autoscaler/app-autoscaler/scalingengine_server
credhub delete -n /bosh-autoscaler/app-autoscaler/scalingengine_ca
credhub delete -n /bosh-autoscaler/app-autoscaler/autoscaler_service_broker_password
credhub delete -n /bosh-autoscaler/app-autoscaler/database_password
set -e
