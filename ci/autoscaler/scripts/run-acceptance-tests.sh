#!/bin/bash

set -euo pipefail

VAR_DIR=bbl-state/bbl-state/vars
pushd bbl-state/bbl-state
  eval "$(bbl print-env)"
popd

export GOPATH=$PWD/app-autoscaler-release
pushd app-autoscaler-release/src/acceptance
cat > acceptance_config.json <<EOF
{
  "api": "api.${SYSTEM_DOMAIN}",
  "admin_user": "admin",
  "admin_password": "${CF_ADMIN_PASSWORD}",
  "apps_domain": "${SYSTEM_DOMAIN}",
  "skip_ssl_validation": ${SKIP_SSL_VALIDATION},
  "use_http": false,

  "service_name": "autoscaler",
  "service_plan": "autoscaler-free-plan",
  "aggregate_interval": 120,
  "name_prefix": "${NAME_PREFIX}",

  "autoscaler_api": "autoscaler.${SYSTEM_DOMAIN}",
  "service_offering_enabled": ${SERVICE_OFFERING_ENABLED}
}
EOF

CONFIG=$PWD/acceptance_config.json ./bin/test -race -nodes=3 -slowSpecThreshold=120 -trace ${SUITES}

popd
