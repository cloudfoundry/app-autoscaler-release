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
root_dir=$(realpath -e "${script_dir}/../../../..")

export PR_NUMBER=${PR_NUMBER:-$(gh pr view --json number --jq '.number' || echo 44)}
[ "${PR_NUMBER}" == "44" ] && warn "no PR_NUMBER is set, will use the default (44)";
debug "PR_NUMBER: ${PR_NUMBER}"

export SYSTEM_DOMAIN="${SYSTEM_DOMAIN:-"autoscaler.app-runtime-interfaces.ci.cloudfoundry.org"}"
debug "SYSTEM_DOMAIN: ${SYSTEM_DOMAIN}"
# shellcheck disable=SC2034
system_domain="${SYSTEM_DOMAIN}"

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


