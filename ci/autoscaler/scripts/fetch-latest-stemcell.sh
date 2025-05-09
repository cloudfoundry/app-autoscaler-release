#!/bin/bash

set -euo pipefail
script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${script_dir}/common.sh"


function main(){
  bosh_login
  bosh upload-stemcell --sha1 "sha256:$(cat gcp-jammy-stemcell/sha256)" "$(cat gcp-jammy-stemcell/url)"
}

[ "${BASH_SOURCE[0]}" == "${0}" ] && main "$@"
