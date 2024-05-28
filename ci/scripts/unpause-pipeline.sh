#!/usr/bin/env bash

target=$(yq -r ".targets | keys | .[0]" ~/.flyrc)

function unpause-job(){
  pipeline=$1
  jobs=$(fly -t "$target" jobs -p "$pipeline" --json | jq ".[] | select(.paused==true) | .name" -r )
  selected_job=$(gum choose --no-limit "$jobs" --header "Select jobs to unpause from pipeline $pipeline"
)

  for j in $selected_job; do
    fly -t "$target" unpause-job -j "$pipeline"/"$j"
  done
}

function unpause-pipeline(){
  payload=$(fly -t "$target" pipelines --json)

  pipelines=$(echo "$payload" | jq ".[] |.name" -r | sort)
  pipeline=$(gum choose "$pipelines" "all")

  if [[ "$pipeline" == "all" ]]; then
    for p in $pipelines; do
      fly -t "$target" unpause-pipeline -p "$p"
      unpause-job "$p"
    done
  else
    fly -t "$target" unpause-pipeline -p "$pipeline"
    unpause-job "$pipeline"
  fi

}


unpause-pipeline "${@:-}"
