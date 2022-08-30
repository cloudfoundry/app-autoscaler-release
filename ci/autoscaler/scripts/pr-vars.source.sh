#!/bin/bash
# Source this file please.
# Moved to ci/  *DO NOT MODIFY MANUALLY*
# shellcheck disable=SC2155
export script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
echo ">> sourcing pr-vars.source.sh from ${script_dir}"

if [[ -z ${PR_NUMBER} ]]; then echo -e "?? WARN: no PR_NUMBER is set, will use the default\n"; fi

export PR_NUMBER=${PR_NUMBER:-44}
echo ">> PR_NUMBER: ${PR_NUMBER}"

export BUILDIN_MODE=${BUILDIN_MODE:-"false"}

export SERVICE_OFFERING_ENABLED=${SERVICE_OFFERING_ENABLED:-true}

export BBL_STATE_PATH=${BBL_STATE_PATH:-"${script_dir}/../../../../app-autoscaler-env-bbl-state/bbl-state"}
echo  ">> BBL_STATE_PATH: ${BBL_STATE_PATH}"

export SYSTEM_DOMAIN="autoscaler.app-runtime-interfaces.ci.cloudfoundry.org"
echo ">> SYSTEM_DOMAIN: ${SYSTEM_DOMAIN}"

export AUTOSCALER_DIR="${AUTOSCALER_DIR:-"./"}"
echo ">> AUTOSCALER_DIR: ${AUTOSCALER_DIR}"

export CI_DIR="${script_dir}/../ci"
echo ">> CI_DIR: ${CI_DIR}"

export SKIP_SSL_VALIDATION=true
echo  ">> SKIP_SSL_VALIDATION: ${SKIP_SSL_VALIDATION}"

export SUITES=${SUITES:-"api app broker"}
echo ">> SUITES: ${SUITES}"

export NODES=${NODES:-3}
echo ">> NODES: ${NODES}"

export DEPLOYMENT_NAME=${DEPLOYMENT_NAME:-"autoscaler-${PR_NUMBER}"}
if [[ "${BUILDIN_MODE}" == "true" ]]; then
    export DEPLOYMENT_NAME=${DEPLOYMENT_NAME}-buildin
    export SERVICE_OFFERING_ENABLED=false
    fi
echo ">> DEPLOYMENT_NAME: ${DEPLOYMENT_NAME}"
echo ">> BUILDIN_MODE: ${BUILDIN_MODE}"
echo ">> SERVICE_OFFERING_ENABLED: ${SERVICE_OFFERING_ENABLED}"

export SERVICE_NAME="${DEPLOYMENT_NAME}"
echo ">> SERVICE_NAME: ${SERVICE_NAME}"

export SERVICE_BROKER_NAME="${DEPLOYMENT_NAME}servicebroker"
echo ">> SERVICE_BROKER_NAME: ${SERVICE_BROKER_NAME}"

export NAME_PREFIX=${NAME_PREFIX:-"${DEPLOYMENT_NAME}-TESTS"}
echo ">> NAME_PREFIX: ${NAME_PREFIX}"
