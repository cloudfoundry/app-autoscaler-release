#!/bin/bash

set -euo pipefail
script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${script_dir}/vars.source.sh"

cf_admin_password="${CF_ADMIN_PASSWORD:-}"
service_offering_enabled="${SERVICE_OFFERING_ENABLED:-true}"
skip_ssl_validation="${SKIP_SSL_VALIDATION:-true}"
skip_teardown="${SKIP_TEARDOWN:-false}"
use_existing_organization="${USE_EXISTING_ORGANIZATION:-false}"
existing_organization="${EXISTING_ORGANIZATION:-}"
use_existing_space="${USE_EXISTING_SPACE:-false}"
existing_space="${EXISTING_SPACE:-}"
suites=${SUITES:-"api app broker"}
ginkgo_opts="${GINKGO_OPTS:-}"
nodes="${NODES:-3}"
performance_app_count="${PERFORMANCE_APP_COUNT:-}"
performance_app_percentage_to_scale="${PERFORMANCE_APP_PERCENTAGE_TO_SCALE:-}"
performance_setup_workers="${PERFORMANCE_SETUP_WORKERS:-}"
performance_update_existing_org_quota=${PERFORMANCE_UPDATE_EXISTING_ORG_QUOTA:-true}
cpu_upper_threshold=${CPU_UPPER_THRESHOLD:-100}

if [[ ! -d ${bbl_state_path} ]]; then
  echo "FAILED: Did not find bbl-state folder at ${bbl_state_path}"
  echo "Make sure you have checked out the app-autoscaler-env-bbl-state repository next to the app-autoscaler-release repository to run this target or indicate its location via BBL_STATE_PATH";
  exit 1;
fi

if [[ -z ${cf_admin_password} ]]; then
  pushd "${bbl_state_path}"
    eval "$(bbl print-env)"
  popd

  cf_admin_password=$(credhub get -n /bosh-autoscaler/cf/cf_admin_password -q)
fi

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
	"use_existing_organization": ${use_existing_organization},
  "existing_organization": "${existing_organization}",
  "use_existing_space": ${use_existing_space},
  "existing_space": "${existing_space}",
  "service_plan": "autoscaler-free-plan",
  "aggregate_interval": 120,
	"default_timeout": 60,
	"cpu_upper_threshold": ${cpu_upper_threshold},
  "name_prefix": "${name_prefix}",

  "autoscaler_api": "${deployment_name}.${system_domain}",
  "service_offering_enabled": ${service_offering_enabled},

  "performance": {
    "app_count": ${performance_app_count},
    "app_percentage_to_scale": ${performance_app_percentage_to_scale},
    "setup_workers": ${performance_setup_workers},
    "update_existing_org_quota": ${performance_update_existing_org_quota}
  }
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
  SKIP_TEARDOWN=$skip_teardown CONFIG=$PWD/acceptance_config.json DEBUG=true ./bin/test -race -nodes="${nodes}" -trace $ginkgo_opts ${suites_to_run}
else
  log "Nothing to run!"
fi

popd >/dev/null
