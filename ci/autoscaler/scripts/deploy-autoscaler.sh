#!/bin/bash

set -euo pipefail

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
  # Determine if we need to upload a stemcell at this point.
  STEMCELL_OS=$(yq eval '.stemcells[] | select(.alias == "default").os' templates/app-autoscaler-deployment.yml)
  STEMCELL_VERSION=$(yq eval '.stemcells[] | select(.alias == "default").version' templates/app-autoscaler-deployment.yml)
  STEMCELL_NAME="bosh-google-kvm-${STEMCELL_OS}-go_agent"
  set +e
  STEMCELL_EXISTS=$(bosh stemcells | grep -c "${STEMCELL}")
  set -e

  if [[ "${STEMCELL_EXISTS}" == 0 ]]; then
    wget "https://bosh.io/d/stemcells/${STEMCELL_NAME}?v=${STEMCELL_VERSION}" -O stemcell.tgz
    bosh -n upload-stemcell stemcell.tgz
  fi

  CURRENT_COMMIT_HASH=$(git log -1 --pretty=format:"%H")
  set +e
  AUTOSCALER_EXISTS=$(bosh releases | grep -c "${CURRENT_COMMIT_HASH}")
  set -e
  if [[ "${AUTOSCALER_EXISTS}" == 1 ]]; then
    echo "the app-autoscaler release is already uploaded with the commit ${CURRENT_COMMIT_HASH}"
    echo "Attempting redeploy..." 

    bosh -n -d app-autoscaler \
      deploy templates/app-autoscaler-deployment.yml \
      -o example/operation/loggregator-certs-from-cf.yml \
      -v system_domain=${SYSTEM_DOMAIN} \
      -v cf_client_id=autoscaler_client_id \
      -v cf_client_secret=autoscaler_client_secret \
      -v skip_ssl_validation=true
    exit 0
  fi

  echo "Creating Release"
  bosh create-release --force --version=${CURRENT_COMMIT_HASH}

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
