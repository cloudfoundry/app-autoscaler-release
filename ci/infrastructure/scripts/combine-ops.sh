#!/bin/bash

set -euo pipefail

mkdir -p combined-ops/operations
mkdir -p combined-ops/operations/cf
mkdir -p combined-ops/operations/autoscaler
mkdir -p combined-ops/operations/prometheus

cp -r cf-deployment/operations/* combined-ops/operations/cf/ \
cp -r app-autoscaler-release/ci/operations/* combined-ops/autoscaler/operations
cp -r prometheus-deployment/manifests/operators/* combined-ops/prometheus/operations
