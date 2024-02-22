#!/bin/bash
set -euo pipefail

config="$(cat "${CONFIG:-}")"

function getConfItem(){
  val=$(jq -r ".$1" <<< "${config}")
  if [ "$val" = "null" ]; then return 1; fi
  echo "$val"
}

if [ -z "${config}" ]; then
  echo "ERROR: Please supply the config using CONFIG env variable"
  exit 1
fi

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"

app_dir="$(realpath -e "${script_dir}/build")"
app_domain="$(getConfItem apps_domain)"
service_name="$(getConfItem service_name)"
memory_mb="$(getConfItem node_memory_limit||echo 128)"
service_broker="$(getConfItem service_broker)"
service_plan="$(getConfItem service_plan)"

cf create-org "test"
cf target -o "test"
cf create-space "test_app"
cf target -s "test_app"

cf enable-service-access "${service_name}" -b "${service_broker}" -p  "${service_plan}" -o test
cf create-service "${service_name}" "${service_plan}" "${service_name}" -b "${service_broker}" --wait

pushd "${app_dir}" >/dev/null
cf push \
  --var app_name="test_app" \
  --var app_domain="${app_domain}" \
  --var service_name="${service_name}" \
  --var instances=1 \
  --var memory_mb="${memory_mb}" \
  --var buildpacks="binary_buildpack" \
  -f "manifest.yml" \
  -c ./app
popd > /dev/null
