#!/bin/bash
set -euo pipefail
script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${script_dir}/common.sh"

pushd "${bbl_state_path}" > /dev/null
  eval "$(bbl print-env)"
popd > /dev/null



write_gcp_service_account_key() {
  if [ -f "${bbl_gcp_service_account_json}" ]; then
    cp "${bbl_gcp_service_account_json}" /tmp/google_service_account.json
  else
    echo "${bbl_gcp_service_account_json}" > /tmp/google_service_account.json
  fi
  bbl_gcp_service_account_json="/tmp/google_service_account.json"
}

write_gcp_service_account_key
gcloud auth activate-service-account --key-file=${bbl_gcp_service_account_json}
gcp_dns_values=$(gcloud dns record-sets list --project "${bbl_gcp_project_wg}" --zone "${gcp_dns_zone}" --name "${system_domain}" --format=json | jq -r '.[].rrdatas | join(" ") ')

if [ "${BBL_DNS_VALUES}" == "${gcp_dns_values}" ]; then
  echo "${BBL_DNS_VALUES} is correct"
else
  echo "dns zone:${GCP_DNS_ZONE} name:${system_domain} need to be updated to ${BBL_DNS_VALUES}"
  exit 1
fi
