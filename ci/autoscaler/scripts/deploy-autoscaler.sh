#!/bin/bash
# shellcheck disable=SC2086
set -euo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

system_domain="${SYSTEM_DOMAIN:-autoscaler.app-runtime-interfaces.ci.cloudfoundry.org}"
bbl_state_path="${BBL_STATE_PATH:-bbl-state/bbl-state}"
deployment_name="${DEPLOYMENT_NAME:-app-autoscaler}"
autoscaler_dir="${AUTOSCALER_DIR:-app-autoscaler-release}"
deployment_manifest="${script_dir}/../../../templates/app-autoscaler.yml"
bosh_fix_releases="${BOSH_FIX_RELEASES:-false}"
ops_files=${OPS_FILES:-"${autoscaler_dir}/operations/add-releases.yml\
 ${autoscaler_dir}/operations/instance-identity-cert-from-cf.yml\
 ${autoscaler_dir}/operations/add-postgres-variables.yml\
 ${autoscaler_dir}/operations/enable-nats-tls.yml\
 ${autoscaler_dir}/operations/loggregator-certs-from-cf.yml\
 ${autoscaler_dir}/operations/add-extra-plan.yml\
 ${autoscaler_dir}/operations/set-release-version.yml\
 ${autoscaler_dir}/operations/enable-log-cache.yml\
 ${autoscaler_dir}/operations/log-cache-syslog-server.yml"}
CURRENT_COMMIT_HASH=$(cd "${autoscaler_dir}"; git log -1 --pretty=format:"%H")
bosh_release_version=${RELEASE_VERSION:-${CURRENT_COMMIT_HASH}-${deployment_name}}

pushd "${bbl_state_path}" > /dev/null
  eval "$(bbl print-env)"
popd > /dev/null

echo "# Deploying autoscaler '${bosh_release_version}' with name '${deployment_name}' "

UAA_CLIENT_SECRET=$(credhub get -n /bosh-autoscaler/cf/uaa_admin_client_secret --quiet)
export UAA_CLIENT_SECRET
CF_ADMIN_PASSWORD=$(credhub get -n /bosh-autoscaler/cf/cf_admin_password -q)

uaac target "https://uaa.${system_domain}" --skip-ssl-validation
uaac token client get admin -s "$UAA_CLIENT_SECRET"

set +e
exist=$(uaac client get autoscaler_client_id | grep -c NotFound)
set -e

function deploy () {
  bosh_deploy_args=""

  if [[ $bosh_fix_releases == "true" ]]; then
    bosh_fix_releases="${BOSH_FIX_RELEASES:-true}"
    bosh_deploy_args="$bosh_deploy_args --fix-releases"
  fi

  echo " - Deploy args: '${bosh_deploy_args}'"

  echo "# creating Bosh deployment '${deployment_name}' with version '${bosh_release_version}' in system domain '${system_domain}'   "
  bosh -n -d "${deployment_name}" \
    deploy "$deployment_manifest" \
    ${OPS_FILES_TO_USE} \
    ${bosh_deploy_args} \
    -v system_domain="${system_domain}" \
    -v deployment_name="${deployment_name}" \
    -v app_autoscaler_version="${bosh_release_version}" \
    -v admin_password="${CF_ADMIN_PASSWORD}" \
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

pushd "$autoscaler_dir" > /dev/null
  # Determine if we need to upload a stemcell at this point.
  #TODO refactor out function for stemcell check and update.
  STEMCELL_OS=$(yq eval '.stemcells[] | select(.alias == "default").os' $deployment_manifest)
  STEMCELL_VERSION=$(yq eval '.stemcells[] | select(.alias == "default").version' $deployment_manifest)
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
  for OPS_FILE in ${ops_files}; do
    if [ -f "${OPS_FILE}" ]; then
      OPS_FILES_TO_USE="${OPS_FILES_TO_USE} -o ${OPS_FILE}"
    else
      echo "ERROR: could not find ops file ${OPS_FILE} in ${PWD}"
      exit 1
    fi
  done

  echo " - Using Ops files: '${OPS_FILES_TO_USE}'"
  AUTOSCALER_RELEASE_EXISTS=$(bosh releases | grep -c "${bosh_release_version}" || true)
  echo "Checking if release:'${bosh_release_version}' exists: ${AUTOSCALER_RELEASE_EXISTS}"
  if [[ "${AUTOSCALER_RELEASE_EXISTS}" == 0 ]]; then
    echo "Creating Release with bosh version ${bosh_release_version}"
    bosh create-release --force --version="${bosh_release_version}"

    echo "Uploading Release"
    bosh upload-release "dev_releases/app-autoscaler/app-autoscaler-${bosh_release_version}.yml"
  else
    echo "the app-autoscaler release is already uploaded with the commit ${bosh_release_version}"
    echo "Attempting redeploy..."
  fi

  deploy
popd > /dev/null
