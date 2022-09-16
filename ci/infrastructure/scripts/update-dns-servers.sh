#!/bin/bash
set -euo pipefail

pushd bbl-state/bbl-state
  BBL_DNS_VALUES=$(bbl outputs | yq eval '.system_domain_dns_servers | join(" ")' -)
popd


write_gcp_service_account_key() {
  if [ -f "${BBL_GCP_SERVICE_ACCOUNT_KEY}" ]; then
    cp "${BBL_GCP_SERVICE_ACCOUNT_KEY}" /tmp/google_service_account.json
  else
    echo "${BBL_GCP_SERVICE_ACCOUNT_KEY}" > /tmp/google_service_account.json
  fi
  export BBL_GCP_SERVICE_ACCOUNT_KEY="/tmp/google_service_account.json"
}

write_gcp_service_account_key
gcloud auth activate-service-account --key-file=${BBL_GCP_SERVICE_ACCOUNT_KEY}
GCP_DNS_VALUES=$(gcloud dns record-sets list --project "${BBL_GCP_PROJECT_ID}" --zone "${GCP_DNS_ZONE}" --name "${GCP_DNS_NAME}" --format=json | jq -r '.[].rrdatas | join(" ") ')

if [ "${BBL_DNS_VALUES}" == "${GCP_DNS_VALUES}" ]; then
  echo "${BBL_DNS_VALUES} is correct"
else
  echo "dns zone:${GCP_DNS_ZONE} name:${GCP_DNS_NAME} need to be updated to ${BBL_DNS_VALUES}"
  exit 1
fi
