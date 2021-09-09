#!/bin/bash

set -euo pipefail

echo "default"
conftest test -p default -o github ../../../templates/app-autoscaler-deployment.yml

echo "postgres-persistent-disk-after"
bosh int -o ../../operation/postgres-persistent-disk.yml ../../../templates/app-autoscaler-deployment.yml | conftest test -o github -p postgres-persistent-disk-after -
