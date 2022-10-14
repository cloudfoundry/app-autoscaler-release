#! /usr/bin/env bash
#
# To run this script you need to have set up the target using
# fly -t "autoscaler" login -c "https://bosh.ci.cloudfoundry.org" -n "app-autoscaler"
#
# When running concourse locally: ` fly -t "local" login -c "http://localhost:8080" `
# Then  `TARGET=local set-pipeline.sh`
set -euo pipefail

function set_pipeline(){
  fly -t "${TARGET}" set-pipeline --config="pipeline.yml" --pipeline="${PIPELINE_NAME}" -v branch_name="${CURRENT_BRANCH}"
}

SCRIPT_RELATIVE_DIR=$(dirname "${BASH_SOURCE[0]}")
pushd "${SCRIPT_RELATIVE_DIR}" > /dev/null
  TARGET="${TARGET:-autoscaler}"
  CURRENT_BRANCH="$(git symbolic-ref --short HEAD)"

  if [[ "$CURRENT_BRANCH" == "main" ]];then
    PIPELINE_NAME="app-autoscaler-release"
    set_pipeline
  else
    PIPELINE_NAME="app-autoscaler-release-${CURRENT_BRANCH}"
    set_pipeline
  fly -t "${TARGET}" pause-job -j "${PIPELINE_NAME}/release"
  fi

  fly -t autoscaler unpause-pipeline -p "${PIPELINE_NAME}"
popd > /dev/null
