#!/bin/bash

set -euo pipefail

mkdir -p combined-ops/operations
mkdir -p combined-ops/operations/cf
mkdir -p combined-ops/operations/autoscaler
mkdir -p combined-ops/operations/prometheus/operators/cf

cp -r cf-deployment/operations/* combined-ops/operations/cf/
cp -r ci/ci/operations/* combined-ops/operations/autoscaler
cp -r prometheus-boshrelease/manifests/operators/cf/* combined-ops/operations/prometheus/operators/cf
