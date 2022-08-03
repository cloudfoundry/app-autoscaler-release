#!/bin/bash
set -euo pipefail
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
pushd "${script_dir}" > /dev/null
test_app_name=test_app
test_service_name=test_service
test_org="testing"
test_space="testing"
number_apps=${NUMBER_OF_APPS:-1}
service_offering="app-autoscaler-${PR_NUMBER}"
app_location="${script_dir}/../src/acceptance/assets/app/nodeApp"

function create_app {
   local app_name=$1
   cf push --var app_name="${app_name}"\
      --var app_domain=autoscaler.ci.cloudfoundry.org\
      --var service_name="${service_offering}"\
      --var instances=1\
      --var buildpack=nodejs_buildpack\
      --var node_tls_reject_unauthorized=0\
      -p "${app_location}"\
      -f "${app_location}/app_manifest.yml"\
      --no-start &
  cf app "${app_name}" --guid
  cf bind-service "${app_name}" "${test_service_name}"
  cf start "${app_name}"
}

# shellcheck disable=SC1091
source ./pr-vars.source.sh
cf create-org "${test_org}"
cf target -o "${test_org}"
cf create-space "${test_space}"
cf target -s "${test_space}"
cf enable-service-access "${service_offering}" -b "${service_offering}" -o "${test_org}"
cf create-service "${service_offering}" autoscaler-free-plan "${test_service_name}" -b "${service_offering}"
for app_number in $(seq 1 "${number_apps}") ; do
  app_name="${test_app_name}-${app_number}"
  echo " - creating app ${app_name}"
  create_app "${app_name}"
done
wait