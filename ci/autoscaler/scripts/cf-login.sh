#!/bin/bash
# shellcheck disable=SC2086
set -euo pipefail
script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${script_dir}/vars.source.sh"

if [[ ! -d ${bbl_state_path} ]]; then
  echo "FAILED: Did not find bbl-state folder at ${bbl_state_path}"
  echo "Make sure you have checked out the app-autoscaler-env-bbl-state repository next to the app-autoscaler-release repository to run this target or indicate its location via BBL_STATE_PATH";
  exit 1;
  fi

pushd "${bbl_state_path}" > /dev/null
  eval "$(bbl print-env)"
popd > /dev/null

cf api "https://api.${SYSTEM_DOMAIN}" --skip-ssl-validation

cf auth admin "$(credhub get -n /bosh-autoscaler/cf/cf_admin_password -q)"

cf autoscaling-api "https://autoscaler-${PR_NUMBER}.${SYSTEM_DOMAIN}"
