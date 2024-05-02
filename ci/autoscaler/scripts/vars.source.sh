#!/bin/bash
# Source this file please.
# Moved to ci/  *DO NOT MODIFY MANUALLY*

# NOTE: to turn on debug use DEBUG=true
# shellcheck disable=SC2155
if [ -z "${BASH_SOURCE[0]}" ]; then
  echo  "### Source this from inside a script only! "
  echo  "### ======================================="
  echo
  return
fi

write_error_state() {
  echo "Error failed execution of \"$1\" at line $2"
  local frame=0
  while true ; do
    caller $frame && break
    ((frame++));
  done
}

trap 'write_error_state "$BASH_COMMAND" "$LINENO"' ERR

debug=${DEBUG:-}
if [ -n "${debug}" ] && [ ! "${debug}" = "false" ]; then
  function debug(){ echo "  -> $1"; }
else
  function debug(){ :; }
fi

function warn(){
  echo " - WARN: $1"
}

function log(){
  echo " - $1"
}

function step(){
  echo "# $1"
}

script_dir="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dir=$(realpath -e "${script_dir}/../../..")

export PR_NUMBER=${PR_NUMBER:-$(gh pr view --json number --jq '.number' )}
debug "PR_NUMBER: '${PR_NUMBER}'"
user=${USER:-"test"}

export DEPLOYMENT_NAME="${DEPLOYMENT_NAME:-"autoscaler-${PR_NUMBER}"}"
[ "${DEPLOYMENT_NAME}" = "autoscaler-" ] && DEPLOYMENT_NAME="${user}"
debug "DEPLOYMENT_NAME: ${DEPLOYMENT_NAME}"
log "set up vars: DEPLOYMENT_NAME=${DEPLOYMENT_NAME}"
# shellcheck disable=SC2034
deployment_name="${DEPLOYMENT_NAME}"

export SYSTEM_DOMAIN="${SYSTEM_DOMAIN:-"autoscaler.app-runtime-interfaces.ci.cloudfoundry.org"}"
debug "SYSTEM_DOMAIN: ${SYSTEM_DOMAIN}"
# shellcheck disable=SC2034
system_domain="${SYSTEM_DOMAIN}"

BBL_STATE_PATH="${BBL_STATE_PATH:-$( realpath -e "${root_dir}/../app-autoscaler-env-bbl-state/bbl-state" 2> /dev/null || echo "${root_dir}/../bbl-state/bbl-state" )}"
BBL_STATE_PATH="$(realpath -e "${BBL_STATE_PATH}" || echo "ERR_invalid_state_path" )"
export BBL_STATE_PATH
debug  "BBL_STATE_PATH: ${BBL_STATE_PATH}"
# shellcheck disable=SC2034
bbl_state_path="${BBL_STATE_PATH}"

AUTOSCALER_DIR="${AUTOSCALER_DIR:-${root_dir}}"
export AUTOSCALER_DIR="$(realpath -e "${AUTOSCALER_DIR}" )"
debug "AUTOSCALER_DIR: ${AUTOSCALER_DIR}"
# shellcheck disable=SC2034
autoscaler_dir="${AUTOSCALER_DIR}"

CI_DIR="${CI_DIR:-$(realpath -e "${root_dir}/ci")}"
export CI_DIR="$(realpath -e "${CI_DIR}")"
debug "CI_DIR: ${CI_DIR}"
# shellcheck disable=SC2034
ci_dir="${CI_DIR}"

export SERVICE_NAME="${DEPLOYMENT_NAME}"
debug "SERVICE_NAME: ${SERVICE_NAME}"
# shellcheck disable=SC2034
service_name="%{SERVICE_NAME"

export SERVICE_BROKER_NAME="${DEPLOYMENT_NAME}servicebroker"
debug "SERVICE_BROKER_NAME: ${SERVICE_BROKER_NAME}"
# shellcheck disable=SC2034
service_broker_name="${SERVICE_BROKER_NAME}"

export NAME_PREFIX="${NAME_PREFIX:-"${DEPLOYMENT_NAME}-TESTS"}"
debug "NAME_PREFIX: ${NAME_PREFIX}"
# shellcheck disable=SC2034
name_prefix="${NAME_PREFIX}"

export SERVICE_OFFERING_ENABLED=${SERVICE_OFFERING_ENABLED:-true}
debug "SERVICE_OFFERING_ENABLED: ${SERVICE_OFFERING_ENABLED}"

export GINKGO_OPTS=${GINKGO_OPTS:-"--fail-fast"}

export PERFORMANCE_APP_COUNT="${PERFORMANCE_APP_COUNT:-50}"
debug "PERFORMANCE_APP_COUNT: ${PERFORMANCE_APP_COUNT}"

export PERFORMANCE_APP_PERCENTAGE_TO_SCALE="${PERFORMANCE_APP_PERCENTAGE_TO_SCALE:-30}"
debug "PERFORMANCE_APP_PERCENTAGE_TO_SCALE: ${PERFORMANCE_APP_PERCENTAGE_TO_SCALE}"

export PERFORMANCE_SETUP_WORKERS="${PERFORMANCE_SETUP_WORKERS:-20}"
debug "PERFORMANCE_SETUP_WORKERS: ${PERFORMANCE_SETUP_WORKERS}"

export PERFORMANCE_TEARDOWN=${PERFORMANCE_TEARDOWN:-true}
debug "PERFORMANCE_TEARDOWN: ${PERFORMANCE_TEARDOWN}"

export CPU_UPPER_THRESHOLD=${CPU_UPPER_THRESHOLD:-100}
debug "CPU_UPPER_THRESHOLD: ${CPU_UPPER_THRESHOLD}"
# shellcheck disable=SC2034
cpu_upper_threshold=${CPU_UPPER_THRESHOLD}
