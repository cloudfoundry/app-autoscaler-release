#!/bin/sh -x

set -euo pipefail

ORG_PREFIX=ASATS

ORGS=$(cf orgs | grep -v name | grep ${ORG_PREFIX})
for ORG in $ORGS; do
	set +e
	cf delete-org $ORG -f
	if [ "$?" != "0" ]; then 
		cf target -o $ORG
		SERVICES=$(cf services | grep -v name | grep autoscaler | awk '{print $1}')
		for SERVICE in $SERVICES; do
			cf purge-service-instance $SERVICE -f
		done
		cf delete-org -f $ORG
	fi
	set -e
done
