#!/bin/bash
set -euo pipefail
set -x

pushd autoscaler-env-bbl-state
  terraform output -state=bbl-state/vars/terraform.tfstate -json | jq -r '.system_domain_dns_servers.value'
popd


