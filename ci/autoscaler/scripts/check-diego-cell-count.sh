#!/bin/bash
# shellcheck disable=SC2086
set -euo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

bbl_state_path="${BBL_STATE_PATH:-bbl-state/bbl-state}"
actual_cells=$( bosh -d cf manifest | yq ".instance_groups | map(select(.name == \"diego-cell\")) | .[] | .instances")

pushd "${bbl_state_path}" > /dev/null
  eval "$(bbl print-env)"
popd > /dev/null

[[ $expected_cells != $actual_cells ]] && exit 1
