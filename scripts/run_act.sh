#!/bin/bash
# shellcheck disable=SC2012,SC2068
#
workflows="$(ls .github/workflows  | cut -d'.' -f1)"


select workflow in ${workflows[@]}; do
  jobs=$(yq -r ".jobs | keys | .[]"  ".github/workflows/${workflow}.yaml")
  select job in ${jobs[@]}; do
    act -r --workflows "./.github/workflows/${workflow}.yaml" --job "$job"
    break;
  done
  break;
done
