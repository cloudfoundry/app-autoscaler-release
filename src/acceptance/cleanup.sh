#!/bin/sh -x

set -euo pipefail

ORG_PREFIX=ASATS

ORGS=$(cf orgs | grep -v name | grep ${ORG_PREFIX})
for ORG in $ORGS; do
	set +e
	SERVICE_INSTANCE=$(cf delete-org $ORG -f 2>&1 | grep "Service instance" | awk '{print $3}' | sed 's/://g')
	set -e
	if [ "$SERVICE_INSTANCE" != "" ]; then
		cf target -o $ORG
		cf purge-service-instance -f $SERVICE_INSTANCE
		cf delete-org -f $ORG
	fi
done
