#!/bin/bash
set -euo pipefail
set -x

pushd autoscaler-env-bbl-state/bbl-state
  bbl outputs | yq eval '.system_domain_dns_servers | join(" ")' -
popd


