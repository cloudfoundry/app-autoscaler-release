#!/bin/bash

set -euo pipefail

echo "postgres-persistent-disk-before"
bosh int ../../../templates/app-autoscaler-deployment.yml > generated.yml
conftest test -o github -p postgres-persistent-disk-before generated.yml
rm generated.yml

echo "postgres-persistent-disk-after"
bosh int -o ../../operation/postgres-persistent-disk.yml ../../../templates/app-autoscaler-deployment.yml > generated.yml
conftest test -o github -p postgres-persistent-disk-after generated.yml
rm generated.yml
