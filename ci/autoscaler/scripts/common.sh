
script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${script_dir}/vars.source.sh"

function bosh_login(){
  pushd "${bbl_state_path}" > /dev/null || exit
    eval "$(bbl print-env)"
  popd > /dev/null || exit
}

function cf_login(){
  cf api "https://api.${system_domain}" --skip-ssl-validation
  CF_ADMIN_PASSWORD=$(credhub get -n /bosh-autoscaler/cf/cf_admin_password -q)
  cf auth admin "$CF_ADMIN_PASSWORD"
}

function cleanup_organization(){
  echo "# Cleaning up from acceptance tests"
  pushd "${ci_dir}/../src/acceptance" > /dev/null || exit
    ./cleanup.sh
  popd > /dev/null || exit
}

function cleanup_service_broker(){
  echo "- Deleting service broker for deployment '${deployment_name}'"
  SERVICE_BROKER_EXISTS=$(cf service-brokers | grep -c "${service_broker_name}.${system_domain}" || true)
  if [[ $SERVICE_BROKER_EXISTS == 1 ]]; then
    echo "- Service Broker exists, deleting broker '${deployment_name}'"
    cf delete-service-broker "${deployment_name}" -f
  fi
}

function cleanup_bosh_deployment(){
  echo "- Deleting bosh deployment '${deployment_name}'"
  bosh delete-deployment -d "${deployment_name}" -n
}

function cleanup_bosh(){
  bosh clean-up --all -n
}


function cleanup_credhub(){
  echo "- Deleting credhub creds: '/bosh-autoscaler/${deployment_name}/*'"
  credhub delete -p "/bosh-autoscaler/${deployment_name}"
}

