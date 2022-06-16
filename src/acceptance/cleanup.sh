#!/bin/bash

set -euo pipefail

function delete_org(){
  local ORG=$1
  if ! cf delete-org "$ORG" -f; then
    cf target -o "$ORG"
    SERVICES=$(cf services | grep "${SERVICE_PREFIX}" |  awk 'NR>1 { print $1}')
    for SERVICE in $SERVICES; do
      cf purge-service-instance "$SERVICE" -f || echo "ERROR: purge-service-instance '$SERVICE' failed"
    done
    cf delete-org -f "$ORG" || echo "ERROR: delete-org '$ORG' failed"
  fi
  echo "Deleted $ORG"
}

deployment_name="${DEPLOYMENT_NAME:-app-autoscaler}"
ORG_PREFIX="ASATS|ASUP|CUST_MET|${deployment_name}-TESTS"
SERVICE_PREFIX=autoscaler

ORGS=$(cf orgs |  awk 'NR>3{ print $1}' | grep -E "${ORG_PREFIX}" || true)
echo "Deleting orgs: '${ORGS}'"

for ORG in $ORGS; do
	# shellcheck disable=SC2181
  delete_org "$ORG" &
done

wait