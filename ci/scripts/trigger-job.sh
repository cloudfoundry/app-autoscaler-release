
#!/usr/bin/env bash
# shellcheck disable=SC2086
#

set -euo pipefail

function trigger-job(){
  target="app-autoscaler-release"
  pipelines=$(fly -t "$target" pipelines --json | jq ".[] |.name" -r | sort)
  pipeline=$(gum choose $pipelines)
  jobs=$(fly -t "$target" jobs -p "$pipeline" --json | jq ".[] | .name" -r)
  job=$(gum choose $jobs)
  fly -t "$target" trigger-job -j "$pipeline/$job"
}

trigger-job
