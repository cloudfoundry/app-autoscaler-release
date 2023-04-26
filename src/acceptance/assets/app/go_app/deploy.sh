#!/bin/bash

set -euo pipefail
config=${CONFIG:-}
function getConfItem(){
  val=$(jq -r ".$1" "${config}")
  if [ "$val" = "null" ]; then return 1; fi
  echo "$val"
}
if [ -z "${config}" ]; then
  echo "ERROR: Please supply the config using CONFIG env variable"
  exit 1
fi

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
app_dir="$( realpath -e "${script_dir}/build")"
service_name=$(getConfItem service_name)

cf create-org test
cf target -o test
cf create-space test_app
cf target -s test_app

cf enable-service-access "${service_name}" -b "$(getConfItem service_broker)" -p  "$(getConfItem service_plan)" -o test
cf create-service "${service_name}" "$(getConfItem service_plan)" "${service_name}" -b "$(getConfItem service_broker)" --wait

pushd "${app_dir}" >/dev/null
cf push\
  --var app_name=test_app\
  --var app_domain="$(getConfItem apps_domain)"\
  --var service_name="${service_name}"\
  --var instances=1\
  --var memory_mb="$(getConfItem node_memory_limit||echo 128)"\
  --var buildpack="binary_buildpack"\
  -f "manifest.yml"\
  -c ./app
popd > /dev/null