#!/bin/bash

set -euo pipefail

VAR_DIR=bbl-state/bbl-state/vars

system_domain="${SYSTEM_DOMAIN:-autoscaler.ci.cloudfoundry.org}"
bbl_state_path="${BBL_STATE_PATH:-bbl-state/bbl-state}"
ops_files="${OPS_FILES:-()}"

VAR_DIR=bbl-state/bbl-state/vars
pushd ${bbl_state_path}
  eval "$(bbl print-env)"
popd

export UAA_CLIENT_SECRET=$(credhub get -n /bosh-autoscaler/cf/uaa_admin_client_secret --quiet)
CF_ADMIN_PASSWORD=$(credhub get -n /bosh-autoscaler/cf/cf_admin_password -q)

uaac target https://uaa.${system_domain} --skip-ssl-validation
uaac token client get admin -s $UAA_CLIENT_SECRET

set +e
exist=$(uaac client get autoscaler_client_id | grep -c NotFound)
set -e

function deploy () {
  bosh -n -d app-autoscaler \
    deploy templates/app-autoscaler-deployment.yml \
    ${OPS_FILES_TO_USE} \
    -v system_domain=${system_domain} \
    -v app_autoscaler_version=${CURRENT_COMMIT_HASH} \
    -v admin_password=${CF_ADMIN_PASSWORD} \
    -v cf_client_id=autoscaler_client_id \
    -v cf_client_secret=autoscaler_client_secret \
    -v skip_ssl_validation=true
}

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
  STEMCELL_EXISTS=$(bosh stemcells | grep -c "${STEMCELL_NAME}")
  set -e

  if [[ "${STEMCELL_EXISTS}" == 0 ]]; then
    URL="https://bosh.io/d/stemcells/${STEMCELL_NAME}"
    if [ "${STEMCELL_VERSION}" != "latest" ]; then
	    URL="${URL}?v=${STEMCELL_VERSION}"
    fi
    wget "$URL" -O stemcell.tgz
    bosh -n upload-stemcell stemcell.tgz
  fi

  OPS_FILES_TO_USE=""
  if [ -f REQUIRED_OPS_FILES ]; then
    for OPS_FILE in $(cat REQUIRED_OPS_FILES); do
      if [ -f "${OPS_FILE}" ]; then
        OPS_FILES_TO_USE="${OPS_FILES_TO_USE} -o ${OPS_FILE}"
      fi
    done
  fi
  for OPS_FILE in ${ops_files}; do
    if [ -f "${OPS_FILE}" ]; then
      OPS_FILES_TO_USE="${OPS_FILES_TO_USE} -o ${OPS_FILE}"
    fi
  done


  CURRENT_COMMIT_HASH=$(git log -1 --pretty=format:"%H")
      cat << EOF > release_version.yml
---
- type: replace
  path: /releases/name=app-autoscaler/version?
  value: ((app_autoscaler_version))
EOF

  OPS_FILES_TO_USE="${OPS_FILES_TO_USE} -o release_version.yml"
  echo " - Using Ops files: '${OPS_FILES_TO_USE}'"
  set +e
  AUTOSCALER_EXISTS=$(bosh releases | grep -c "${CURRENT_COMMIT_HASH}")
  set -e
  if [[ "${AUTOSCALER_EXISTS}" == 1 ]]; then
    echo "the app-autoscaler release is already uploaded with the commit ${CURRENT_COMMIT_HASH}"
    echo "Attempting redeploy..."
    deploy

    exit 0
  fi

  echo "Creating Release with bosh version ${CURRENT_COMMIT_HASH}"
  bosh create-release --force --version=${CURRENT_COMMIT_HASH}

  echo "Uploading Release"
  bosh upload-release

  echo "Deploying Release"
  deploy

popd
