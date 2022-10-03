#!/bin/bash -x
# Source this file please.
# Moved to ci/  *DO NOT MODIFY MANUALLY*
# shellcheck disable=SC2155
if [ -z "${BASH_SOURCE[0]}" ]; then
  echo "### Source this from inside a script only! "
  echo "### ======================================="
  echo
  return
fi

script_dir="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dir=$(realpath -e "${script_dir}/../../..")
echo ">> sourcing pr-vars.source.sh from ${script_dir}"

export PR_NUMBER=${PR_NUMBER:-44}
if [[  ${PR_NUMBER} == 44 ]]; then echo -e "?? WARN: no PR_NUMBER is set, will use the default\n"; fi
echo ">> PR_NUMBER: ${PR_NUMBER}"
export BBL_STATE_PATH="${BBL_STATE_PATH:-$( realpath -e "${root_dir}/../app-autoscaler-env-bbl-state/bbl-state" 2> /dev/null || realpath -e "${root_dir}/../bbl-state/bbl-state" 2> /dev/null )}"
echo  ">> BBL_STATE_PATH: ${BBL_STATE_PATH}"

export SYSTEM_DOMAIN="${SYSTEM_DOMAIN:-"autoscaler.app-runtime-interfaces.ci.cloudfoundry.org"}"
echo ">> SYSTEM_DOMAIN: ${SYSTEM_DOMAIN}"

export AUTOSCALER_DIR="${AUTOSCALER_DIR:-${root_dir}}"
echo ">> AUTOSCALER_DIR: ${AUTOSCALER_DIR}"

export CI_DIR="${CI_DIR:-$(realpath -e "${root_dir}/ci")}"
echo ">> CI_DIR: ${CI_DIR}"

export SKIP_SSL_VALIDATION=${SKIP_SSL_VALIDATION:-'true'}
echo  ">> SKIP_SSL_VALIDATION: ${SKIP_SSL_VALIDATION}"

export SUITES="${SUITES:-"api app broker"}"
echo ">> SUITES: ${SUITES}"

export NODES="${NODES:-3}"
echo ">> NODES: ${NODES}"

export BUILDIN_MODE=${BUILDIN_MODE:-"false"}
echo ">> BUILDIN_MODE: ${BUILDIN_MODE}"

export DEPLOYMENT_NAME="${DEPLOYMENT_NAME:-"autoscaler-${PR_NUMBER}"}"
echo ">> DEPLOYMENT_NAME: ${DEPLOYMENT_NAME}"

export SERVICE_NAME="${DEPLOYMENT_NAME}"
echo ">> SERVICE_NAME: ${SERVICE_NAME}"

export SERVICE_BROKER_NAME="${DEPLOYMENT_NAME}servicebroker"
echo ">> SERVICE_BROKER_NAME: ${SERVICE_BROKER_NAME}"

export NAME_PREFIX="${NAME_PREFIX:-"${DEPLOYMENT_NAME}-TESTS"}"
echo ">> NAME_PREFIX: ${NAME_PREFIX}"

export SERVICE_OFFERING_ENABLED=${SERVICE_OFFERING_ENABLED:-true}
echo ">> SERVICE_OFFERING_ENABLED: ${SERVICE_OFFERING_ENABLED}"

export GINKGO_OPTS=${GINKGO_OPTS:-"--fail-fast"}
