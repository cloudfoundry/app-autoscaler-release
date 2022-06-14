#!/bin/bash

set -euo pipefail

system_domain="${SYSTEM_DOMAIN:-autoscaler.ci.cloudfoundry.org}"
service_name="${SERVICE_NAME:-autoscaler}"
deployment_name="${DEPLOYMENT_NAME:-autoscaler}"
bbl_state_path="${BBL_STATE_PATH:-bbl-state/bbl-state}"
autoscaler_dir="${AUTOSCALER_DIR:-app-autoscaler-release}"

pushd ${bbl_state_path}
  eval "$(bbl print-env)"
popd

CF_ADMIN_PASSWORD=$(credhub get -n /bosh-autoscaler/cf/cf_admin_password -q)

export GOPATH="$PWD/app-autoscaler-release"
pushd "${autoscaler_dir}/src/acceptance"
cat > acceptance_config.json <<EOF
{
  "api": "api.${system_domain}",
  "admin_user": "admin",
  "admin_password": "${CF_ADMIN_PASSWORD}",
  "apps_domain": "${system_domain}",
  "skip_ssl_validation": ${SKIP_SSL_VALIDATION},
  "use_http": false,
  "service_name": "${deployment_name}",
  "service_broker": "${service_name}",
  "service_plan": "autoscaler-free-plan",
  "aggregate_interval": 120,
  "name_prefix": "${NAME_PREFIX}",

  "autoscaler_api": "${deployment_name}.${system_domain}",
  "service_offering_enabled": ${SERVICE_OFFERING_ENABLED}
}
EOF

SUITES_TO_RUN=""
for SUITE in $SUITES; do
  echo "Checking suite $SUITE"
  if [[ -d "$SUITE" ]]; then
     echo "Adding suite $SUITE to list"
     SUITES_TO_RUN="$SUITES_TO_RUN $SUITE"
  fi
done

echo "Running $SUITES_TO_RUN"

if [ "${SUITES_TO_RUN}" != "" ]; then
  CONFIG=$PWD/acceptance_config.json ./bin/test -skip "mtls" -race -nodes=${NODES} --slow-spec-threshold=120s -trace ${SUITES_TO_RUN}
else
  echo "Nothing to run!"
fi

popd
