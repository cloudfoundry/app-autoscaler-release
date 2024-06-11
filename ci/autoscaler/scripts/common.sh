#!/bin/bash
script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${script_dir}/vars.source.sh"
set -euo pipefail

function retry(){
  max_retries=$1
  shift
  retries=0
  command="$*"
  until [ "${retries}" -eq "${max_retries}" ] || $command; do
    ((retries=retries+1))
    echo " - retrying command '${command}' attempt: ${retries}"
  done
  [ "${retries}" -lt "${max_retries}" ] || { echo "ERROR: Command '$*' failed after ${max_retries} attempts"; return 1; }
}

function bosh_login(){
  step "bosh login"
  if [[ ! -d ${bbl_state_path} ]]; then
    echo "FAILED: Did not find bbl-state folder at ${bbl_state_path}"
    echo "Make sure you have checked out the app-autoscaler-env-bbl-state repository next to the app-autoscaler-release repository to run this target or indicate its location via BBL_STATE_PATH";
    exit 1;
  fi

  pushd "${bbl_state_path}" > /dev/null
    eval "$(bbl print-env)"
  popd > /dev/null
}

function cf_login(){
  step "login to cf"
  cf api "https://api.${system_domain}" --skip-ssl-validation
  cf_admin_password="$(credhub get --quiet --name='/bosh-autoscaler/cf/cf_admin_password')"
  cf auth admin "$cf_admin_password"
}

function cleanup_acceptance_run(){
  step "cleaning up from acceptance tests"
  pushd "${ci_dir}/../src/acceptance" > /dev/null
    retry 5 ./cleanup.sh
  popd > /dev/null
}

function cleanup_service_broker(){
  step "deleting service broker for deployment '${deployment_name}'"
  SERVICE_BROKER_EXISTS=$(cf service-brokers | grep -c "${service_broker_name}.${system_domain}" || true)
  if [[ $SERVICE_BROKER_EXISTS == 1 ]]; then
    echo "- Service Broker exists, deleting broker '${deployment_name}'"
    retry 3 cf delete-service-broker "${deployment_name}" -f
  fi
}

function cleanup_bosh_deployment(){
  step "deleting bosh deployment '${deployment_name}'"
  retry 3 bosh delete-deployment -d "${deployment_name}" -n
}

function delete_releases(){
  step "deleting releases"
  if [ -n "${deployment_name}" ]
  then
    for release in $(bosh releases | grep -E "${deployment_name}\s+"  | awk '{print $2}')
    do
       echo "- Deleting bosh release '${release}'"
       bosh delete-release -n "app-autoscaler/${release}" &
    done
    wait
  fi
}

function cleanup_bosh(){
  step "cleaning up bosh"
  retry 3 bosh clean-up --all -n
}

function cleanup_credhub(){
  step "cleaning up credhub: '/bosh-autoscaler/${deployment_name}/*'"
  retry 3 credhub delete --path="/bosh-autoscaler/${deployment_name}"
}

function unset_vars() {
  unset PR_NUMBER
  unset DEPLOYMENT_NAME
  unset SYSTEM_DOMAIN
  unset BBL_STATE_PATH
  unset AUTOSCALER_DIR
  unset CI_DIR
  unset SERVICE_NAME
  unset SERVICE_BROKER_NAME
  unset NAME_PREFIX
  unset GINKGO_OPTS
}

function find_or_create_org(){
  local org_name="$1"
  if ! cf orgs | grep -q "${org_name}"; then
    cf create-org "${org_name}"
  fi
  cf target -o "${org_name}"
}

function find_or_create_space(){
  local space_name="$1"
  if ! cf spaces | grep -q "${space_name}"; then
    cf create-space "${space_name}"
  fi
  cf target -s "${space_name}"
}

function cf_target(){
  local org_name="$1"
  local space_name="$2"

  find_or_create_org "${org_name}"
  find_or_create_space "${space_name}"
}
