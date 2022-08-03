#!/bin/bash
set -euo pipefail
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

pushd "$script_dir" >/dev/null
open_prs="$(gh pr list --state open --json "number" --jq ".[].number" | tr '\n' '|')"
open_prs=${open_prs%?}
echo "Open prs:'${open_prs}'"
closed_prs=$(bosh deployments | grep app-autoscaler | awk '{ print $1 }' | grep -vE "${open_prs}"  | sed -e 's/app-autoscaler-//' | tr '\n' ' ')
echo "Closed but still deployed PRs:'${closed_prs}'"

# shellcheck disable=SC2034
# shellcheck disable=SC2013
for PR_NUMBER in ${closed_prs}
do
  echo "Cleaning up PR:${PR_NUMBER}"
  # shellcheck disable=SC1091
  source ./pr-vars.source.sh
  "${CI_DIR}/autoscaler/scripts/cleanup-autoscaler.sh"
done
popd "${script_dir}" >/dev/null