#!/bin/bash

set -euo pipefail
export DEPLOYMENT_NAME="app-autoscaler-44"
export SERVICE_BROKER_NAME="app-autoscaler-44servicebroker"
export SERVICE_NAME="autoscaler-44"
export BBL_STATE_PATH="../app-autoscaler-env-bbl-state/bbl-state"
export SYSTEM_DOMAIN="autoscaler.ci.cloudfoundry.org"
export AUTOSCALER_DIR="${PWD}"
export CI_DIR="../app-autoscaler-ci/"
export SERVICE_OFFERING_ENABLED=true
export SKIP_SSL_VALIDATION=true
export NAME_PREFIX="TESTS-${DEPLOYMENT_NAME}"
export SUITES="api app broker"
export NODES=8
export SKIP_TEARDOWN=true
"${CI_DIR}/autoscaler/scripts/run-acceptance-tests.sh"