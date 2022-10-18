#!/bin/bash
# shellcheck disable=SC2086
set -euo pipefail
script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${script_dir}/common.sh"

TARGET="${TARGET:-autoscaler}"

function get_pipelines(){
  fly -t "${TARGET}" pipelines --json | jq ".[].name" -r
}
function destroy_pipeline(){
  local pipeline_name="$1"
  fly -t "${TARGET}" destroy-pipeline -p  "${pipeline_name}"
}

for pipeline_name in $(get_pipelines |grep -v "app-autoscaler-release$"| grep -v "infrastructure"); do
		destroy_pipeline "${pipeline_name}"
done

