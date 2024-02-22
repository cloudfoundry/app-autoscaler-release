#!/bin/bash
set -euo pipefail

# read content of config right at the beginning to avoid errors after switching directories
readonly CONFIG_CONTENT
CONFIG_CONTENT="$(cat "${CONFIG:-}" 2> /dev/null || echo "")"

function getConfItem(){
  local key="$1"
  local val

  val=$(jq -r ".${key}" <<< "${CONFIG_CONTENT}")

  if [ "${val}" = "null" ]; then
    return 1;
  fi

  echo "${val}"
}

function main(){
  if [ -z "${CONFIG_CONTENT}" ]; then
    echo "ERROR: Please supply the config using CONFIG env variable"
    exit 1
  fi

  local org, space
  org="test"
  space="test_$(whoami)"

  cf create-org "${org}"
  cf target -o "${org}"
  cf create-space "${space}"
  cf target -s "${space}"

  local app_name, app_domain, service_name, memory_mb, service_broker, service_plan
  app_name="test_app"
  app_domain="$(getConfItem apps_domain)"
  service_name="$(getConfItem service_name)"
  memory_mb="$(getConfItem node_memory_limit||echo 128)"
  service_broker="$(getConfItem service_broker)"
  service_plan="$(getConfItem service_plan)"

  cf enable-service-access "${service_name}" -b "${service_broker}" -p  "${service_plan}" -o test
  cf create-service "${service_name}" "${service_plan}" "${service_name}" -b "${service_broker}" -t "app-autoscaler" --wait

  # make sure that the current directory is the one which contains the build artifacts like binary and manifest.yml
  local script_dir, app_dir
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

  cf bind-service "${app_name}" "${service_name}"

  # restaging app so that it is able to access the VCAP environment
  cf restage "${app_name}"
}

main
