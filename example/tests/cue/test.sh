#!/bin/bash

set -euo pipefail

echo "postgres-persistent-disk.before.cue"
bosh int ../../../templates/app-autoscaler-deployment.yml \
    | cue vet postgres-persistent-disk.before.cue yaml: -

echo "postgres-persistent-disk.after.cue"
bosh int -o ../../operation/postgres-persistent-disk.yml ../../../templates/app-autoscaler-deployment.yml \
    | cue vet postgres-persistent-disk.after.cue yaml: -

