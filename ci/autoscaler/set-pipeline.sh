#!/bin/bash
# To run this script you need to have set up the target using
# fly -t "autoscaler" login -c "https://bosh.ci.cloudfoundry.org" -n "app-autoscaler"
# When running locally: ` fly -t "local" login -c "http://localhost:8080" `
# Then  `TARGET=local set-pipeline.sh`
set -euo pipefail

TARGET=${TARGET:-autoscaler}

PIPELINE_NAME=app-autoscaler-release

fly -t $TARGET set-pipeline --config=pipeline.yml --pipeline=$PIPELINE_NAME
