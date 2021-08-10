#!/bin/bash
set -euo pipefail
set -x

pushd autoscaler-env-bbl-state/bbl-state
  bbl outputs | yq eval '.system_domain_dns_servers | join(" ")' -
popd


write_gcp_service_account_key() {
  set +x
  if [ -f "${BBL_GCP_SERVICE_ACCOUNT_KEY}" ]; then
    cp "${BBL_GCP_SERVICE_ACCOUNT_KEY}" /tmp/google_service_account.json
  else
    echo "${BBL_GCP_SERVICE_ACCOUNT_KEY}" > /tmp/google_service_account.json
  fi
  export BBL_GCP_SERVICE_ACCOUNT_KEY="/tmp/google_service_account.json"
  set -x
}

write_gcp_service_account_key
gcloud auth activate-service-account --key-file=${BBL_GCP_SERVICE_ACCOUNT_KEY}
gcloud dns record-sets list --project ${BBL_GCP_PROJECT_ID} --zone ${GCP_DNS_ZONE} --name ${GCP_DNS_NAME} --format=json | jq -r '.[].rrdatas | join(" ") '

