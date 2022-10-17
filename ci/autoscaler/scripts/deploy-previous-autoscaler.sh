#!/bin/bash
set -euo pipefail
script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${script_dir}/vars.source.sh"

pushd "${bbl_state_path}" > /dev/null
  eval "$(bbl print-env)"
popd > /dev/null

RELEASE_URL="$(cat previous-stable-release/url)"
RELEASE_SHA="$(cat previous-stable-release/sha1)"
RELEASE_VERSION="$(cat previous-stable-release/version)"

echo "Downloading release '$RELEASE_VERSION'/${RELEASE_SHA} from '$RELEASE_URL'"
bosh upload-release --sha1 "${RELEASE_SHA}" "${RELEASE_URL}"
export RELEASE_VERSION
"${script_dir}/deploy-autoscaler.sh"