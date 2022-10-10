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

function step(){
  echo "# $1"
}

script_dir="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dir=$(realpath -e "${script_dir}/../../..")

BBL_STATE_PATH="${BBL_STATE_PATH:-$( realpath -e "${root_dir}/../app-autoscaler-env-bbl-state/bbl-state" 2> /dev/null || realpath -e "${root_dir}/../bbl-state/bbl-state" 2> /dev/null )}"
export BBL_STATE_PATH="$(realpath -e "${BBL_STATE_PATH}" )"
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

export SYSTEM_DOMAIN="${SYSTEM_DOMAIN:-"autoscaler.app-runtime-interfaces.ci.cloudfoundry.org"}"
debug "SYSTEM_DOMAIN: ${SYSTEM_DOMAIN}"
# shellcheck disable=SC2034
system_domain="${SYSTEM_DOMAIN}"

BOSH_USERNAME="${BOSH_USERNAME:-admin}"
export BOSH_USERNAME
debug "BOSH_USERNAME: ${BOSH_USERNAME}"
# shellcheck disable=SC2034
bosh_username="${BOSH_USERNAME}"