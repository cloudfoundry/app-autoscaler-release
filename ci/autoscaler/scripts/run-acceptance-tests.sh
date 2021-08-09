#!/bin/bash

set -euo pipefail

VAR_DIR=autoscaler-env-bbl-state/bbl-state/vars
pushd autoscaler-env-bbl-state/bbl-state
  eval "$(bbl print-env)"
popd

cf api https://api.${SYSTEM_DOMAIN} --skip-ssl-validation

CF_ADMIN_PASSWORD=$(credhub get -n /bosh-autoscaler/cf/cf_admin_password -q)
cf auth admin $CF_ADMIN_PASSWORD

# cf login -a https://api.bosh-lite.com -u admin -p admin --skip-ssl-validation

# set +e
# cf delete-service-broker -f autoscaler
# set -e

set +e
SERVICE_BROKER_EXISTS=$(cf service-brokers | grep -c autoscalerservicebroker.${SYSTEM_DOMAIN}) 
set -e
if [[ $SERVICE_BROKER_EXISTS == 1 ]]; then
  echo "Service Broker already exists"
else
  AUTOSCALER_SERVICE_BROKER_PASSWORD=$(credhub get  -n /bosh-autoscaler/app-autoscaler/autoscaler_service_broker_password -q)
  cf create-service-broker autoscaler autoscaler_service_broker_user $AUTOSCALER_SERVICE_BROKER_PASSWORD https://autoscalerservicebroker.${SYSTEM_DOMAIN}
  #cf enable-service-access autoscaler
fi


export GOPATH=$PWD/app-autoscaler-release
pushd app-autoscaler-release/src/acceptance
cat > acceptance_config.json <<EOF
{
  "api": "api.${SYSTEM_DOMAIN}",
  "admin_user": "admin",
  "admin_password": "${CF_ADMIN_PASSWORD}",
  "apps_domain": "${SYSTEM_DOMAIN}",
  "skip_ssl_validation": ${SKIP_SSL_VALIDATION},
  "use_http": false,

  "service_name": "autoscaler",
  "service_plan": "autoscaler-free-plan",
  "aggregate_interval": 120,

  "autoscaler_api": "autoscaler.${SYSTEM_DOMAIN}",
  "service_offering_enabled": ${SERVICE_OFFERING_ENABLED}
}
EOF

CONFIG=$PWD/acceptance_config.json ./bin/test -nodes=3 -slowSpecThreshold=120 -trace api app

popd























VAR_DIR=autoscaler-env-bbl-state/bbl-state/vars
pushd autoscaler-env-bbl-state/bbl-state
  eval "$(bbl print-env)"
popd

export UAA_CLIENT_SECRET=$(credhub get -n /bosh-autoscaler/cf/uaa_admin_client_secret --quiet)

uaac target https://uaa.$SYSTEM_DOMAIN --skip-ssl-validation
uaac token client get admin -s $UAA_CLIENT_SECRET

set +e
exist=$(uaac client get autoscaler_client_id | grep -c NotFound)
set -e

if [[ $exist == 0 ]]; then
  echo "Updating client token"
  uaac client update "autoscaler_client_id" \
	    --authorities "cloud_controller.read,cloud_controller.admin,uaa.resource,routing.routes.write,routing.routes.read,routing.router_groups.read"
else
  echo "Creating client token"
  uaac client add "autoscaler_client_id" \
	--authorized_grant_types "client_credentials" \
	--authorities "cloud_controller.read,cloud_controller.admin,uaa.resource,routing.routes.write,routing.routes.read,routing.router_groups.read" \
	--secret "autoscaler_client_secret"
fi

pushd app-autoscaler-release

  set +e
  autoscalerExists=$(bosh releases | grep -c app-autoscaler)
  set -e
  if [[ $autoscalerExists == 1 ]];then
    deployedCommitHash=$(bosh releases | grep app-autoscaler | awk -F ' ' '{print $3}' | sed 's/\+//g')
    currentCommitHash=$(git log -1 --pretty=format:"%H")
    set +e
    theSame=$(echo ${currentCommitHash} | grep -c ${deployedCommitHash})
    set -e
    if [[ $theSame == 1 ]];then
      echo "the app-autoscaler deployed ${deployedCommitHash} and the current ${currentCommitHash} are the same"
      echo "Deploying Release"
      bosh -n -d app-autoscaler \
        deploy templates/app-autoscaler-deployment.yml \
        -o example/operation/loggregator-certs-from-cf.yml \
        -v system_domain=${SYSTEM_DOMAIN} \
        -v cf_client_id=autoscaler_client_id \
        -v cf_client_secret=autoscaler_client_secret \
        -v skip_ssl_validation=true
      exit 0
    else
      echo "the app-autoscaler deployed ${deployedCommitHash} and the current ${currentCommitHash} are NOT the same"
    fi
  fi

  release_version=$(git log --pretty=format:"%H" -1)
  echo "Creating Release"
  bosh create-release --force --version=${release_version}

  echo "Uploading Release"
  bosh upload-release

  echo "Deploying Release"
  bosh -n -d app-autoscaler \
    deploy templates/app-autoscaler-deployment.yml \
    -o example/operation/loggregator-certs-from-cf.yml \
    -v system_domain=${SYSTEM_DOMAIN} \
    -v cf_client_id=autoscaler_client_id \
    -v cf_client_secret=autoscaler_client_secret \
    -v skip_ssl_validation=true

popd
