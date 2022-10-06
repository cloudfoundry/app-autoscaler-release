#!/bin/bash -x
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

debug=${DEBUG:-}
if [ -n "${debug}" ]; then
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

script_dir="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dir=$(realpath -e "${script_dir}/../../..")

export PR_NUMBER=${PR_NUMBER:-44}
if [[  ${PR_NUMBER} == 44 ]]; then warn "no PR_NUMBER is set, will use the default"; fi
debug "PR_NUMBER: ${PR_NUMBER}"

export DEPLOYMENT_NAME="${DEPLOYMENT_NAME:-"autoscaler-${PR_NUMBER}"}"
debug "DEPLOYMENT_NAME: ${DEPLOYMENT_NAME}"
log "set up vars: DEPLOYMENT_NAME=${DEPLOYMENT_NAME}"

export SYSTEM_DOMAIN="${SYSTEM_DOMAIN:-"autoscaler.app-runtime-interfaces.ci.cloudfoundry.org"}"
debug "SYSTEM_DOMAIN: ${SYSTEM_DOMAIN}"

BBL_STATE_PATH="${BBL_STATE_PATH:-$( realpath -e "${root_dir}/../app-autoscaler-env-bbl-state/bbl-state" 2> /dev/null || realpath -e "${root_dir}/../bbl-state/bbl-state" 2> /dev/null )}"
export BBL_STATE_PATH="$(realpath -e "${BBL_STATE_PATH}" )"
debug  "BBL_STATE_PATH: ${BBL_STATE_PATH}"

AUTOSCALER_DIR="${AUTOSCALER_DIR:-${root_dir}}"
export AUTOSCALER_DIR="$(realpath -e "${AUTOSCALER_DIR}" )"
debug "AUTOSCALER_DIR: ${AUTOSCALER_DIR}"

CI_DIR="${CI_DIR:-$(realpath -e "${root_dir}/ci")}"
export CI_DIR="$(realpath -e "${CI_DIR}")"
debug "CI_DIR: ${CI_DIR}"

export SKIP_SSL_VALIDATION=${SKIP_SSL_VALIDATION:-'true'}
debug  "SKIP_SSL_VALIDATION: ${SKIP_SSL_VALIDATION}"

export SUITES="${SUITES:-"api app broker"}"
debug "SUITES: ${SUITES}"

export NODES="${NODES:-3}"
debug "NODES: ${NODES}"

export BUILDIN_MODE=${BUILDIN_MODE:-"false"}
debug "BUILDIN_MODE: ${BUILDIN_MODE}"

export SERVICE_NAME="${DEPLOYMENT_NAME}"
debug "SERVICE_NAME: ${SERVICE_NAME}"

export SERVICE_BROKER_NAME="${DEPLOYMENT_NAME}servicebroker"
debug "SERVICE_BROKER_NAME: ${SERVICE_BROKER_NAME}"

export NAME_PREFIX="${NAME_PREFIX:-"${DEPLOYMENT_NAME}-TESTS"}"
debug "NAME_PREFIX: ${NAME_PREFIX}"

export SERVICE_OFFERING_ENABLED=${SERVICE_OFFERING_ENABLED:-true}
debug "SERVICE_OFFERING_ENABLED: ${SERVICE_OFFERING_ENABLED}"

export GINKGO_OPTS=${GINKGO_OPTS:-"--fail-fast"}

