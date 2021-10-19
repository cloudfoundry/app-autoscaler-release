#!/bin/bash

set -euo pipefail

pushd bbl-state/bbl-state
  eval "$(bbl print-env)"
popd

set -x

# Could be removed?
bosh -d app-autoscaler delete-deployment -n
bosh delete-release app-autoscaler -n

RELEASE_URL=$(cat previous-stable-release/url)
RELEASE_SHA=$(cat previous-stable-release/sha1)

bosh upload-release --sha1 "$RELEASE_SHA" "$RELEASE_URL"
bosh releases

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
  OPS_FILES_TO_USE=""
  for OPS_FILE in $OPS_FILES; do
    if [ -f "${OPS_FILE}" ]; then
      OPS_FILES_TO_USE="${OPS_FILES_TO_USE} -o ${OPS_FILE}"
    fi
  done

  set +e
  bosh -n -d app-autoscaler \
    deploy templates/app-autoscaler-deployment.yml \
    ${OPS_FILES_TO_USE} \
    -v system_domain=${SYSTEM_DOMAIN} \
    -v cf_client_id=autoscaler_client_id \
    -v cf_client_secret=autoscaler_client_secret \
    -v skip_ssl_validation=true
  EXIT_CODE=$?
  set -e
  # FIXME this is a hack because of the database migrations
  if [ $EXIT_CODE != "0" ]; then
    bosh -n -d app-autoscaler \
      deploy templates/app-autoscaler-deployment.yml \
      ${OPS_FILES_TO_USE} \
      -v system_domain=${SYSTEM_DOMAIN} \
      -v cf_client_id=autoscaler_client_id \
      -v cf_client_secret=autoscaler_client_secret \
      -v skip_ssl_validation=true
  fi
popd
