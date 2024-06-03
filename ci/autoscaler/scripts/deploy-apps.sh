#! /usr/bin/env bash
# shellcheck disable=SC2086,SC2034,SC2155
set -euo pipefail
script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${script_dir}/vars.source.sh"


pushd "${bbl_state_path}" > /dev/null
  eval "$(bbl print-env)"
popd > /dev/null

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

function cf_login(){
  cf_admin_password="$(credhub get -n /bosh-autoscaler/cf/cf_admin_password -q)"
  cf auth admin "${cf_admin_password}"
}


function setup_autoscaler_uaac(){
  local uaac_authorities="cloud_controller.read,cloud_controller.admin,uaa.resource,routing.routes.write,routing.routes.read,routing.router_groups.read"
  local autoscaler_secret="autoscaler_client_secret"
  local uaa_client_secret=$(credhub get -n /bosh-autoscaler/cf/uaa_admin_client_secret --quiet)
  uaac target "https://uaa.${system_domain}" --skip-ssl-validation > /dev/null
  uaac token client get admin -s "${uaa_client_secret}" > /dev/null

  if uaac client get autoscaler_client_id >/dev/null; then
    step "updating autoscaler uaac client"
    uaac client update "autoscaler_client_id" \
      --authorities "$uaac_authorities" > /dev/null
  else
    step "creating autoscaler uaac client"
    uaac client add "autoscaler_client_id" \
      --authorized_grant_types "client_credentials" \
      --authorities "$uaac_authorities" \
      --secret "$autoscaler_secret" > /dev/null
  fi
}

function cf_target(){
  local org_name="$1"
  local space_name="$2"
  find_or_create_org "${org_name}"
  find_or_create_space "${space_name}"
}

function deploy() {
  pushd "${autoscaler_dir}/src/autoscaler/metricsforwarder" > /dev/null
    log "Deploying metricsforwarder"
    make cf-push
    make stop-metricsforwarder-vm
  popd > /dev/null
}


log "Deploying autoscaler apps for bosh deployment '${deployment_name}' "

pushd "${autoscaler_dir}" > /dev/null
  cf_login
  cf_target "${autoscaler_org}" "${autoscaler_space}"
  deploy
popd > /dev/null
