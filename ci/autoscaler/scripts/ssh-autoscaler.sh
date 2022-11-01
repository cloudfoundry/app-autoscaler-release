#!/bin/bash
# shellcheck disable=SC2086
set -euo pipefail
script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${script_dir}/common.sh"
deployment_name="autoscaler-${PR_NUMBER}"
bosh_login
vms=$(bosh vms -d $deployment_name --json | jq '.Tables[0].Rows | .[] | .instance'  -r)

select vm in ${vms[@]}; do
		bosh ssh -d $deployment_name $vm
  break;
done
