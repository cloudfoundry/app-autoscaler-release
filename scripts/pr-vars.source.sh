#!/bin/bash

# Source this file please - used for manual debug. Adjust as needed.
# shellcheck disable=SC2155
export script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

export DEPLOYMENT_NAME="app-autoscaler-44"
export SERVICE_BROKER_NAME="app-autoscaler-44servicebroker"
export SERVICE_NAME="autoscaler-44"
export BBL_STATE_PATH="${script_dir}/../../app-autoscaler-env-bbl-state/bbl-state"
export SYSTEM_DOMAIN="autoscaler.ci.cloudfoundry.org"
export AUTOSCALER_DIR="${script_dir}/../"
export CI_DIR="${script_dir}/../../app-autoscaler-ci"

export OPS_FILES="${AUTOSCALER_DIR}/example/operation/loggregator-certs-from-cf.yml\
          ${AUTOSCALER_DIR}/example/operation/add-extra-plan.yml\
          ${AUTOSCALER_DIR}/example/operation/set-release-version.yml\
          ${AUTOSCALER_DIR}/example/operation/enable-name-based-deployments.yml"
export DEBUG=true


export SERVICE_OFFERING_ENABLED=true
export SKIP_SSL_VALIDATION=true
export NAME_PREFIX="${DEPLOYMENT_NAME}-TESTS"
export SUITES="api app broker"
export NODES=3
