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

function create_bosh_config(){
   # generate the private.yml file with the credentials
   config_file="${autoscaler_dir}/config/private.yml"
    cat > "$config_file" <<EOF
---
blobstore:
  options:
    credentials_source: static
    json_key:
EOF
    echo ' - Generating private.yml...'
    yq eval -i '.blobstore.options.json_key = strenv(UPLOADER_KEY)' "$config_file"
}


script_dir="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dir=$(realpath -e "${script_dir}/../../../..")


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


