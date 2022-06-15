#!/bin/bash

set -euo pipefail
export script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
source "$script_dir/pr-vars.source.sh"
export SKIP_TEARDOWN=true

export SUITES="app"
export GINKGO_OPTS="--progress"

"${CI_DIR}/autoscaler/scripts/run-acceptance-tests.sh"
