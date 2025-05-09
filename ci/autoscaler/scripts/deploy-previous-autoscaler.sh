#!/bin/bash
set -euo pipefail
script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${script_dir}/vars.source.sh"
source "${script_dir}/common.sh"

bosh_login "${BBL_STATE_PATH}"

RELEASE_URL="$(cat previous-stable-release/url)"
RELEASE_SHA="$(cat previous-stable-release/sha1)"
RELEASE_VERSION="$(cat previous-stable-release/version)"

echo "Downloading release '$RELEASE_VERSION'/${RELEASE_SHA} from '$RELEASE_URL'"
bosh upload-release --sha1 "${RELEASE_SHA}" "${RELEASE_URL}"
export RELEASE_VERSION
"${script_dir}/deploy-autoscaler.sh"
