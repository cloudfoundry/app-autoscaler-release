#!/bin/bash

set -euo pipefail
# shellcheck disable=SC2155
export script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

#PR_NUMBER="801"
#SUITES="app"
# shellcheck disable=SC1091
source "${script_dir}/pr-vars.source.sh"
export SKIP_TEARDOWN=true
export GINKGO_OPTS="--progress --fail-fast -v "
echo "Running acceptance tests for PR: ${PR_NUMBER}"
export NODES=1
"${CI_DIR}/autoscaler/scripts/run-acceptance-tests.sh"
