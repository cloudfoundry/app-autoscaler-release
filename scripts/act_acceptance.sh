#!/bin/bash

set -euo pipefail
# shellcheck disable=SC2155
export script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
# shellcheck disable=SC1091
source "${script_dir}/pr-vars.source.sh"
export SKIP_TEARDOWN=true

export SUITES="app"
export GINKGO_OPTS="--progress"

"${CI_DIR}/autoscaler/scripts/run-acceptance-tests.sh"
