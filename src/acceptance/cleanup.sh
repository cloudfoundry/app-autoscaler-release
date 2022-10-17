#!/bin/bash

set -euo pipefail

function delete_org(){
  local ORG=$1

  if ! cf delete-org "$ORG" -f; then
    cf target -o "$ORG"
    for SERVICE in $(cf services | grep -e "autoscaler" |  awk 'NR>1 { print $1}'); do
      cf purge-service-instance "$SERVICE" -f || echo "ERROR: purge-service-instance '$SERVICE' failed"
    done
    cf delete-org -f "$ORG" || echo "ERROR: delete-org '$ORG' failed"
  fi
  echo " - deleted $ORG"
}

org_prefix=${NAME_PREFIX:-"ASATS|ASUP|CUST_MET"}

ORGS=$(cf orgs |  awk 'NR>3{ print $1}' | grep -E "${org_prefix}" || true)
echo "# deleting orgs: '${ORGS}'"

for ORG in $ORGS; do
	# shellcheck disable=SC2181
  delete_org "$ORG" &
done


if [ -n "${org_prefix}" ]
then
  for user in $(cf curl /v3/users | jq -r '.resources[].username' | grep "${org_prefix}-" )
  do
    echo " - deleting left over user '${user}'"
    cf delete-user -f "$user" &
  done
fi
wait
