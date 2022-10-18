#!/bin/bash

set -euo pipefail
script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${script_dir}/vars.source.sh"

ginkgo_opts="${GINKGO_OPTS:-}"
nodes="${NODES:-3}"
service_offering_enabled="${SERVICE_OFFERING_ENABLED:-true}"
skip_ssl_validation=${SKIP_SSL_VALIDATION:-'true'}
skip_teardown="${SKIP_TEARDOWN:-false}"
suites=${SUITES:-"api app broker"}

pushd "${bbl_state_path}" >/dev/null
  eval "$(bbl print-env)"
popd >/dev/null

cf_admin_password=$(credhub get -n /bosh-autoscaler/cf/cf_admin_password -q)

pushd "${autoscaler_dir}/src/acceptance" >/dev/null
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
  log "checking suite $suite"
  if [[ -d "$suite" ]]; then
     log "Adding suite '$suite' to list"
     suites_to_run="$suites_to_run $suite"
  fi
done

step "running $suites_to_run"

#run suites
if [ "${suites_to_run}" != "" ]; then
  # shellcheck disable=SC2086
  SKIP_TEARDOWN=$skip_teardown CONFIG=$PWD/acceptance_config.json ./bin/test -race -nodes="${nodes}" --slow-spec-threshold=120s -trace $ginkgo_opts ${suites_to_run}
else
  log "Nothing to run!"
fi

popd >/dev/null
