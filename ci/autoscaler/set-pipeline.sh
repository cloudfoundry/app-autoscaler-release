#! /usr/bin/env bash
#
# To run this script you need to have set up the target using
# fly -t "autoscaler" login -c "https://bosh.ci.cloudfoundry.org" -n "app-autoscaler"
#
# When running concourse locally: ` fly -t "local" login -c "http://localhost:8080" `
# Then  `TARGET=local set-pipeline.sh`
set -euo pipefail

TARGET="${TARGET:-autoscaler}"

function set_pipeline(){
  local pipeline_name="$1"

  fly -t "${TARGET}" set-pipeline --config="pipeline.yml" --pipeline="${pipeline_name}" -v branch_name="${CURRENT_BRANCH}"
  fly -t autoscaler unpause-pipeline -p "${pipeline_name}"
}

function pause_job(){
  local job_name="$1"

  fly -t "${TARGET}" pause-job -j "${job_name}"
}

function unpause_job(){
  local job_name="$1"

  fly -t "${TARGET}" unpause-job -j "${job_name}"
}

function get_jobs(){
  local pipeline_name="$1"

  fly -t "${TARGET}" jobs --pipeline="${pipeline_name}" --json  | jq ".[].name" -r
}

function pause_jobs(){
  local pipeline_name="$1"

  for job in $(get_jobs "$pipeline_name"); do
    pause_job "${pipeline_name}/$job"
  done
}

function main(){
  SCRIPT_RELATIVE_DIR=$(dirname "${BASH_SOURCE[0]}")
  pushd "${SCRIPT_RELATIVE_DIR}" > /dev/null
    CURRENT_BRANCH="$(git symbolic-ref --short HEAD)"

    if [[ "$CURRENT_BRANCH" == "main" ]];then
      export PIPELINE_NAME="app-autoscaler-release"
      set_pipeline $PIPELINE_NAME
    else
      export PIPELINE_NAME="app-autoscaler-release-${CURRENT_BRANCH}"
      set_pipeline "$PIPELINE_NAME"
      pause_jobs "$PIPELINE_NAME"
      unpause_job "$PIPELINE_NAME/set-pipeline"
    fi

  popd > /dev/null
}

[ "${BASH_SOURCE[0]}" == "${0}" ] && main "$@"
