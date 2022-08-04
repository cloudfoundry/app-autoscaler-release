#!/bin/bash

# Source this file please - used for manual debug. Adjust as needed.
# shellcheck disable=SC2155
export script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
pr_number=${PR_NUMBER:-44}
export DEPLOYMENT_NAME=${DEPLOYMENT_NAME:-"app-autoscaler-${pr_number}"}
export SERVICE_BROKER_NAME="${DEPLOYMENT_NAME}servicebroker"
export BBL_STATE_PATH="${script_dir}/../../app-autoscaler-env-bbl-state/bbl-state"
export SYSTEM_DOMAIN="autoscaler.app-runtime-interfaces.ci.cloudfoundry.org"
export AUTOSCALER_DIR="${script_dir}/../"
export CI_DIR="${script_dir}/../ci"

export OPS_FILES="${AUTOSCALER_DIR}/example/operation/loggregator-certs-from-cf.yml\
            ${AUTOSCALER_DIR}/example/operation/add-extra-plan.yml\
            ${AUTOSCALER_DIR}/example/operation/set-release-version.yml\
            ${AUTOSCALER_DIR}/example/operation/enable-name-based-deployments.yml\
            ${AUTOSCALER_DIR}/example/operation/enable-log-cache.yml\
            ${AUTOSCALER_DIR}/example/operation/log-cache-syslog-server.yml\
            ${AUTOSCALER_DIR}/example/operation/use_buildin_mode.yml"

export SERVICE_OFFERING_ENABLED=false
export SKIP_SSL_VALIDATION=true
export NAME_PREFIX=${NAME_PREFIX:-"${DEPLOYMENT_NAME}-TESTS"}
export SUITES=${SUITES:-"api app broker"}
export NODES=3
