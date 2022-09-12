#!/bin/bash

set -euo pipefail

system_domain="${SYSTEM_DOMAIN:-autoscaler.app-runtime-interfaces.ci.cloudfoundry.org}"
deployment_name="${DEPLOYMENT_NAME:-app-autoscaler}"
bbl_state_path="${BBL_STATE_PATH:-bbl-state/bbl-state}"
autoscaler_dir="${AUTOSCALER_DIR:-app-autoscaler-release}"
skip_teardown="${SKIP_TEARDOWN:-false}"
skip_ssl_validation="${SKIP_SSL_VALIDATION:-true}"
name_prefix="${NAME_PREFIX:-ASATS}"
# shellcheck disable=SC2034
buildin_mode="${BUILDIN_MODE:-false}"
service_offering_enabled="${SERVICE_OFFERING_ENABLED:-true}"
suites=${SUITES:-"api app broker"}
ginkgo_opts="${GINKGO_OPTS:-}"
nodes="${NODES:-3}"

if [[ ! -d ${bbl_state_path} ]]; then
  echo "FAILED: Did not find bbl-state folder at ${bbl_state_path}"
  echo "Make sure you have checked out the app-autoscaler-env-bbl-state repository next to the app-autoscaler-release repository to run this target or indicate its location via BBL_STATE_PATH";
  exit 1;
  fi


pushd "${bbl_state_path}"
  eval "$(bbl print-env)"
popd

cf_admin_password=$(credhub get -n /bosh-autoscaler/cf/cf_admin_password -q)

export GOPATH="$PWD/app-autoscaler-release"
pushd "${autoscaler_dir}/src/acceptance"
cat > acceptance_config.json <<EOF
{
  "api": "api.${system_domain}",
  "admin_user": "admin",
  "admin_password": "${cf_admin_password}",
  "apps_domain": "${system_domain}",
  "skip_ssl_validation": ${skip_ssl_validation},
  "use_http": false,
  "service_name": "${deployment_name}",
  "service_broker": "${deployment_name}",
  "service_plan": "autoscaler-free-plan",
  "aggregate_interval": 120,
  "name_prefix": "${name_prefix}",

  "autoscaler_api": "${deployment_name}.${system_domain}",
  "service_offering_enabled": ${service_offering_enabled}
}
EOF

suites_to_run=""
for suite in $suites; do
  echo "Checking suite $suite"
  if [[ -d "$suite" ]]; then
     echo "Adding suite $suite to list"
     suites_to_run="$suites_to_run $suite"
  fi
done

echo "Running $suites_to_run"

echo -e "\n>> ACCEPTANCE TEST ENV VARS:"
at_vars=(
system_domain
deployment_name
bbl_state_path
autoscaler_dir
skip_teardown
skip_ssl_validation
name_prefix
buildin_mode
service_offering_enabled
suites
ginkgo_opts
nodes
)

for var in "${at_vars[@]}"; do
  echo "$var: ${!var}"
  done
echo

#run suites
if [ "${suites_to_run}" != "" ]; then
  SKIP_TEARDOWN=$skip_teardown CONFIG=$PWD/acceptance_config.json ./bin/test -race -nodes="${nodes}" --slow-spec-threshold=120s -trace "$ginkgo_opts" "${suites_to_run}"
else
  echo "Nothing to run!"
fi

popd
