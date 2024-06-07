#!/usr/bin/env bash
# shellcheck disable=SC2086
#

set -euo pipefail
set -x
#uses gum to select pipeline

target="app-autoscaler-release"
# connect to a concourse container throgh intercept
function intercept-job(){
  payload=$(fly -t "$target" pipelines --json)
  pipelines=$(echo "$payload" | jq ".[] |.name" -r | sort)
  # ignore shellcheck warning
  pipeline=$(gum choose $pipelines)
  jobs=$(fly -t "$target" jobs -p "$pipeline" --json | jq ".[] | .name" -r)
  job=$(gum choose $jobs)


  fly -t "$target" intercept -j "$pipeline/$job"
}

intercept-job
