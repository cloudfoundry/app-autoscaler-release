#!/bin/bash
set -euo pipefail

# read content of config right at the beginning to avoid errors when switching directories
CONFIG_CONTENT="$(cat "${CONFIG:-}" 2> /dev/null || echo "")"
readonly CONFIG_CONTENT

function getConfItem(){
  local key="$1"
  local val

  val=$(jq -r ".${key}" <<< "${CONFIG_CONTENT}")

  if [ "${val}" = "null" ]; then
    return 1;
  fi

  echo "${val}"
}

function check_requirements(){
  if [ -z "${CONFIG_CONTENT}" ]; then
    echo "ERROR: Couldn't read content of config, please supply the path to config using CONFIG env variable."
    exit 1
  fi
}

function deploy(){
  local org space use_existing_organization
  org="test"
  space="test_$(whoami)"

  use_existing_organization="$(getConfItem use_existing_organization)"
  if ${use_existing_organization}; then
    org="$(getConfItem existing_organization)"
  fi

  cf create-org "${org}"
  cf target -o "${org}"
  cf create-space "${space}"
  cf target -s "${space}"

  local app_name app_domain service_name memory_mb service_broker service_plan service_offering_enabled
  app_name="test_app"
  app_domain="$(getConfItem apps_domain)"
  service_name="$(getConfItem service_name)"
  memory_mb="$(getConfItem node_memory_limit || echo 128)"
  service_broker="$(getConfItem service_broker)"
  service_plan="$(getConfItem service_plan)"
  service_offering_enabled="$(getConfItem service_offering_enabled || echo false)"

  # create app upfront to avoid restaging after binding to service happened
  cf create-app "${app_name}"

  if ${service_offering_enabled}; then
    cf enable-service-access "${service_name}" -b "${service_broker}" -p  "${service_plan}" -o "${org}"
    cf create-service "${service_name}" "${service_plan}" "${service_name}" -b "${service_broker}" --wait
    cf bind-service "${app_name}" "${service_name}"
  else
    echo "INFO: Skipping service instance setup because service_offering_enabled is '${service_offering_enabled}'"
  fi

  # make sure that the current directory is the one which contains the build artifacts like binary and manifest.yml
  local script_dir app_dir
  script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
  app_dir="$(realpath -e "${script_dir}/build")"
  pushd "${app_dir}" >/dev/null
  cf push \
    --var app_name="${app_name}" \
    --var app_domain="${app_domain}" \
    --var service_name="${service_name}" \
    --var instances=1 \
    --var memory_mb="${memory_mb}" \
    -b "binary_buildpack" \
    -f "manifest.yml" \
    -c "./app"
  popd > /dev/null
}

function main(){
  check_requirements
  deploy
}

main
