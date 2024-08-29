#! /usr/bin/env bash
#
# To run this script you need to have set up the target using
# fly login -t app-autoscaler-release -c https://concourse.app-runtime-interfaces.ci.cloudfoundry.org -n app-autoscaler
#
# When running concourse locally: ` fly -t "local" login -c "http://localhost:8080" `
# Then  `TARGET=local set-pipeline.sh`
set -euo pipefail

if "$(which gh)" > /dev/null; then
  echo "gh cli found"
  gh --version
else
  echo "no gh cli found!"
  exit 1
fi


export PR_NUMBER=${PR_NUMBER:-$(gh pr view --json number --jq '.number' )}

fly_args=""

add_var() {
    fly_args="${fly_args} -v ${1}=${2}"
}

TARGET="${TARGET:-app-autoscaler-release}"

function set_pipeline(){
  local pipeline_name="$1"
  add_var branch_name "${CURRENT_BRANCH}"
  if [[ -z $PR_NUMBER ]]; then
    add_var pr_number "${PR_NUMBER}"
    add_var acceptance_deployment_name          "acceptance"
    add_var acceptance_deployment_name_logcache_metron "acceptance-lc"
    add_var acceptance_deployment_name_logcache_syslog "acceptance-lc-sl"
    add_var acceptance_deployment_name_logcache_syslog_cf "acceptance-lc-sl-cf"
  else
    add_var pr_number "${PR_NUMBER}"
    add_var acceptance_deployment_name          "${PR_NUMBER}-acceptance"
    add_var acceptance_deployment_name_logcache_metron "${PR_NUMBER}-acceptance-lc"
    add_var acceptance_deployment_name_logcache_syslog "${PR_NUMBER}-acceptance-lc-sl"
    add_var acceptance_deployment_name_logcache_syslog_cf "${PR_NUMBER}-acceptance-lc-sl-cf"
  fi

  # shellcheck disable=SC2086
  fly -t "${TARGET}" set-pipeline --config="pipeline.yml" --pipeline="${pipeline_name}" $fly_args
  fly -t "${TARGET}" unpause-pipeline -p "${pipeline_name}"
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
      local current_branch_without_slashes
      current_branch_without_slashes="$(echo "${CURRENT_BRANCH}" | sed 's/\//-/g')" # Concourse can't handle slashes in pipeline names

      export PIPELINE_NAME="app-autoscaler-release-${current_branch_without_slashes}"
      set_pipeline "$PIPELINE_NAME"
      pause_jobs "$PIPELINE_NAME"
    fi

  popd > /dev/null
}

[ "${BASH_SOURCE[0]}" == "${0}" ] && main "$@"
