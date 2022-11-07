#!/bin/bash
# shellcheck disable=SC2086
set -euo pipefail
script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${script_dir}/vars.source.sh"

deployment_manifest="${autoscaler_dir}/templates/app-autoscaler.yml"
bosh_deploy_opts="${BOSH_DEPLOY_OPTS:-""}"
bosh_upload_release_opts="${BOSH_UPLOAD_RELEASE_OPTS:-""}"
bosh_upload_stemcell_opts="${BOSH_UPLOAD_STEMCELL_OPTS:-""}"
serial="${SERIAL:-true}"
ops_files=${OPS_FILES:-"${autoscaler_dir}/operations/add-releases.yml\
 ${autoscaler_dir}/operations/instance-identity-cert-from-cf.yml\
 ${autoscaler_dir}/operations/add-postgres-variables.yml\
 ${autoscaler_dir}/operations/connect_to_postgres_with_certs.yml\
 ${autoscaler_dir}/operations/enable-nats-tls.yml\
 ${autoscaler_dir}/operations/loggregator-certs-from-cf.yml\
 ${autoscaler_dir}/operations/add-extra-plan.yml\
 ${autoscaler_dir}/operations/set-release-version.yml\
 ${autoscaler_dir}/operations/enable-log-cache.yml\
 ${autoscaler_dir}/operations/log-cache-syslog-server.yml"}

if [[ ! -d ${bbl_state_path} ]]; then
  echo "FAILED: Did not find bbl-state folder at ${bbl_state_path}"
  echo "Make sure you have checked out the app-autoscaler-env-bbl-state repository next to the app-autoscaler-release repository to run this target or indicate its location via BBL_STATE_PATH";
  exit 1;
  fi

if [[ ${buildin_mode} == "true" ]]; then ops_files+=" ${autoscaler_dir}/operations/use_buildin_mode.yml"; fi;

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

function deploy() {
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

  # Try to silence Prometheus but do not fail deployment if there's an error
  set +e
  ${script_dir}/silence_prometheus_alert.sh "BOSHJobEphemeralDiskPredictWillFill"
  ${script_dir}/silence_prometheus_alert.sh "BOSHJobProcessUnhealthy"
  ${script_dir}/silence_prometheus_alert.sh "BOSHJobUnhealthy"
  set -e

  echo " - Deploy options: '${bosh_deploy_opts}'"
  echo "# creating Bosh deployment '${deployment_name}' with version '${bosh_release_version}' in system domain '${system_domain}'   "
  bosh -n -d "${deployment_name}" \
    deploy "${deployment_manifest}" \
    ${OPS_FILES_TO_USE} \
    ${bosh_deploy_opts} \
    -v system_domain="${system_domain}" \
    -v serial="${serial}" \
    -v deployment_name="${deployment_name}" \
    -v app_autoscaler_version="${bosh_release_version}" \
    -v admin_password="${CF_ADMIN_PASSWORD}" \
    -v cf_client_id=autoscaler_client_id \
    -v cf_client_secret=autoscaler_client_secret \
    -v skip_ssl_validation=true


  echo "# Finish deploy: '${deployment_name}'"
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

function find_or_upload_stemcell() {
  # Determine if we need to upload a stemcell at this point.
  stemcell_os=$(yq eval '.stemcells[] | select(.alias == "default").os' $deployment_manifest)
  stemcell_version=$(yq eval '.stemcells[] | select(.alias == "default").version' $deployment_manifest)
  stemcell_name="bosh-google-kvm-${stemcell_os}-go_agent"

  if ! bosh stemcells | grep "${stemcell_name}" >/dev/null; then
    URL="https://bosh.io/d/stemcells/${stemcell_name}"
    if [ "${stemcell_version}" != "latest" ]; then
	    URL="${URL}?v=${stemcell_version}"
    fi
    wget "$URL" -O stemcell.tgz
    bosh -n upload-stemcell $bosh_upload_stemcell_opts stemcell.tgz
  fi
}

function find_or_upload_release() {
  if ! bosh releases | grep -E "${bosh_release_version}[*]*\s" > /dev/null; then

    local -r release_desc_file="dev_releases/app-autoscaler/app-autoscaler-${bosh_release_version}.yml"
    if [ ! -f "${release_desc_file}" ]
    then
      echo "Creating Release with bosh version ${bosh_release_version}"
      bosh create-release --force --version="${bosh_release_version}"
    else
      echo -e "Release with bosh-version ${bosh_release_version} already locally present. Reusing it."\
        "\n\tIf this does not work, please consider executing `bosh reset-release`."
    fi

    echo "Uploading Release"
    bosh upload-release $bosh_upload_release_opts "${release_desc_file}"
  else
    echo "the app-autoscaler release is already uploaded with the commit ${bosh_release_version}"
    echo "Attempting redeploy..."
  fi
}

pushd "${autoscaler_dir}" > /dev/null
  find_or_upload_stemcell
  find_or_upload_release
  deploy
popd > /dev/null