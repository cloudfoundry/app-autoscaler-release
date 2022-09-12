#! /usr/bin/env bash
#
# To run this script you need to have set up the target using
# fly -t "autoscaler" login -c "https://bosh.ci.cloudfoundry.org" -n "app-autoscaler"
#
# When running concourse locally: ` fly -t "local" login -c "http://localhost:8080" `
# Then  `TARGET=local set-pipeline.sh`
set -euo pipefail

SCRIPT_RELATIVE_DIR=$(dirname "${BASH_SOURCE[0]}")
pushd "${SCRIPT_RELATIVE_DIR}" > /dev/null
  TARGET="${TARGET:-autoscaler}"

  PIPELINE_NAME="app-autoscaler-release"

  fly -t "${TARGET}" set-pipeline --config="pipeline.yml" --pipeline="${PIPELINE_NAME}"
popd