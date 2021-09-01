#!/bin/bash

set -euo pipefail

# this is a really basic check to validate that the peristent disk value is set.
# FIXME we need a much better way of doing this.
echo "no ops files"
ACTUAL=$(bosh int ../templates/app-autoscaler-deployment.yml | yq e '.instance_groups[] | select(.jobs[].name == "postgres").persistent_disk_type' -)
if [ "${ACTUAL}" != "null" ]; then
	echo "FAILED: default has no persistent disk"
	exit 1
fi

echo "operation/postgres-persistent-disk.yml"
ACTUAL=$(bosh int -o operation/postgres-persistent-disk.yml ../templates/app-autoscaler-deployment.yml | yq e '.instance_groups[] | select(.jobs[].name == "postgres").persistent_disk_type' -)
if [ "${ACTUAL}" != "10GB" ]; then
	echo "FAILED: Expected 10GB to be set as the persistent disk size"
	exit 1
fi
