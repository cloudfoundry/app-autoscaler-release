#!/bin/bash
set -euo pipefail
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
pushd "${script_dir}" > /dev/null
test_app_name=test_app
test_service_name=test_service
test_org="testing"
test_space="testing"
service_offering="app-autoscaler-${PR_NUMBER}"
app_location="${script_dir}/../src/acceptance/assets/app/nodeApp"

source ./pr-vars.source.sh
cf target -o "${test_org}" -s "${test_space}"
cf enable-service-access "${service_offering}" -b "autoscaler-${PR_NUMBER}" -o "${test_org}"
cf create-service "${service_offering}" autoscaler-free-plan "${test_service_name}" -b "${SERVICE_NAME}"
cf push --var app_name="${test_app_name}"\
  --var app_domain=autoscaler.ci.cloudfoundry.org\
  --var service_name="${service_offering}"\
  --var instances=1\
  --var buildpack=nodejs_buildpack\
  --var node_tls_reject_unauthorized=0\
  -p "${app_location}"\
  -f "${app_location}/app_manifest.yml"\
  --no-start
cf app "${test_app_name}" --guid
cf bind-service "${test_app_name}" "${test_service_name}"
cf start "${test_app_name}"