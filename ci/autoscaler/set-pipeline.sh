#! /usr/bin/env bash
#
# To run this script you need to have set up the target using
# fly -t "autoscaler" login -c "https://bosh.ci.cloudfoundry.org" -n "app-autoscaler"
#
# When running concourse locally: ` fly -t "local" login -c "http://localhost:8080" `
# Then  `TARGET=local set-pipeline.sh`
set -euo pipefail

function set_pipeline(){
  local pipeline_name="$1"
  fly -t "${TARGET}" set-pipeline --config="pipeline.yml" --pipeline="${pipeline_name}" -v branch_name="${CURRENT_BRANCH}" -v trigger_acceptance=true
  fly -t autoscaler unpause-pipeline -p "${pipeline_name}"
}

function pause_job(){
  local job_name="$1"
  fly -t "${TARGET}" pause-job -j "${job_name}"
}

function main(){
  SCRIPT_RELATIVE_DIR=$(dirname "${BASH_SOURCE[0]}")
  pushd "${SCRIPT_RELATIVE_DIR}" > /dev/null
    TARGET="${TARGET:-autoscaler}"
    CURRENT_BRANCH="$(git symbolic-ref --short HEAD)"

    if [[ "$CURRENT_BRANCH" == "main" ]];then
      export PIPELINE_NAME="app-autoscaler-release"
      set_pipeline $PIPELINE_NAME
    else
      export PIPELINE_NAME="app-autoscaler-release-${CURRENT_BRANCH}"
      set_pipeline "$PIPELINE_NAME"
      pause_job "${PIPELINE_NAME}/release"
      pause_job "${PIPELINE_NAME}/acceptance"
      pause_job "${PIPELINE_NAME}/acceptance-buildin"
      pause_job "${PIPELINE_NAME}/acceptance-log-cache"
      pause_job "${PIPELINE_NAME}/upgrade-test"
    fi


  popd > /dev/null
}

main
