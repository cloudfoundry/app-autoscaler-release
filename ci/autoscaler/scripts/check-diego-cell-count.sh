#!/bin/bash
# shellcheck disable=SC2086
set -euo pipefail
expected_cells=${EXPECTED_CELLS:-2}

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${script_dir}/vars.source.sh"

pushd "${bbl_state_path}" > /dev/null
  eval "$(bbl print-env)"
popd > /dev/null

actual_cells=$( bosh -d cf manifest | yq ".instance_groups | map(select(.name == \"diego-cell\")) | .[] | .instances")

echo "Expeted diego cell count: '${expected_cells}'"
echo "Actual diego cell count: '${actual_cells}'"

if [[  "${expected_cells}" == "${actual_cells}" ]]; then
  echo "Expected cell count match"
  exit 0
else
  echo "Expected cell count does not match"
  exit 1
fi

