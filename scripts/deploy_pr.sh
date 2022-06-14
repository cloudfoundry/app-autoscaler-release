#!/bin/bash

set -euo pipefail
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
pushd "${script_dir}" > /dev/null
source "./deploy_pr.sh"
"${CI_DIR}/autoscaler/scripts/deploy-autoscaler.sh"
"${CI_DIR}/autoscaler/scripts/register-broker.sh"
