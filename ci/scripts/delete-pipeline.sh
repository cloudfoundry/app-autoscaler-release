#!/usr/bin/env bash
# shellcheck disable=SC2086
#

target="app-autoscaler-release"

function delete-pipeline(){
  payload=$(fly -t "$target" pipelines --json)

  pipelines=$(echo "$payload" | jq ".[] |.name" -r | sort)
  # ignore shellcheck warning
  pipeline=$(gum choose $pipelines )

  if [ ! -z "$pipeline" ]; then
    fly -t "$target" destroy-pipeline -p "$pipeline"
  fi
}

function check-login(){
  if ! fly -t "$target" status;  then
    echo
    echo "fly -t $target login"
    echo
    exit 1
  fi
}

check-login
delete-pipeline "${@:-}"
