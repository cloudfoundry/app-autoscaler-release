#!/bin/bash
#! /usr/bin/env bash

set -euo pipefail
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
bbl_state_path="${BBL_STATE_PATH:-bbl-state/bbl-state}"

pushd "${bbl_state_path}" > /dev/null
  eval "$(bbl print-env)"
popd > /dev/null

[ -n "${DEBUG}" ] && set -x

RELEASE_URL="$(cat previous-stable-release/url)"
RELEASE_SHA="$(cat previous-stable-release/sha1)"

bosh upload-release --sha1 "$RELEASE_SHA" "$RELEASE_URL"
export RELEASE_SHA
"${script_dir}/deploy-autoscaler.sh"