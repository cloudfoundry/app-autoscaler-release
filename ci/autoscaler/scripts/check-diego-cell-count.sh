#!/bin/bash
# shellcheck disable=SC2086
set -euo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

bbl_state_path="${BBL_STATE_PATH:-bbl-state/bbl-state}"

pushd "${bbl_state_path}" > /dev/null
  eval "$(bbl print-env)"
popd > /dev/null

actual_cells=$( bosh -d cf manifest | yq ".instance_groups | map(select(.name == \"diego-cell\")) | .[] | .instances")

echo "Expeted diego cell count: '$expected_cells'"
echo "Actual diego cell count: '$actual_cells'"

[[ "$expected_cells" != "$actual_cells" ]] && exit 1
