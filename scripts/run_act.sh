#!/bin/bash
# shellcheck disable=SC2012
#
workflows="$(ls .github/workflows  | cut -d'.' -f1)"


select workflow in "${workflows[@]}"; do
  jobs=$(yq -r ".jobs | keys | .[]"  ".github/workflows/${workflow}.yaml")
  select job in "${jobs[@]}"; do
    act --workflows "./.github/workflows/${workflow}.yaml" --job "$job"
    break;
  done
done
