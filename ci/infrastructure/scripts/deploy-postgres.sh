#!/bin/bash
# shellcheck disable=SC2086
set -euo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${script_dir}/vars.source.sh"
source "${script_dir}/utils.source.sh"

bosh_deploy_opts=${BOSH_DEPLOY_OPTS:-}
deployment_name="${DEPLOYMENT_NAME:-postgres}"

slack_channel="${SLACK_CHANNEL:-cf-dev-autoscaler-alerts}"
slack_webhook="${SLACK_WEBHOOK}"

release_dir="${POSTGRES_DIR:-$(realpath -e ${root_dir}/../postgres-release)}"
repo_dir="${REPO_DIR:-$(realpath -e ${root_dir}/../postgres-repo)}"
deployment_manifest=${DEPLOYMENT_MANIFEST:-"${repo_dir}/templates/postgres.yml"}

release_ops="${repo_dir}/templates/operations"
ops_files=${OPS_FILES:-"${release_ops}/use_ssl.yml\
                       ${release_ops}/add_static_ips.yml\
                       "}


function add_var_to_bosh_deploy_opts() {
  local var_name=$1
  local var_value=$2
  bosh_deploy_opts="${bosh_deploy_opts} -v ${var_name}=${var_value}"
}

function deploy () {
  local ops_files_to_use=""
  validate_ops_files "${ops_files}"

  for OPS_FILE in ${ops_files}; do
    ops_files_to_use="${ops_files_to_use} -o ${OPS_FILE}"
  done

  add_var_to_bosh_deploy_opts "postgres_host_or_ip" "10.0.2.2"

  step "Deploying release with name '${deployment_name}' "
  log "using Ops files: '${ops_files_to_use}'"
  bosh -n -d "${deployment_name}" \
    deploy "${deployment_manifest}" \
    ${ops_files_to_use} \
    ${bosh_deploy_opts}
}

load_bbl_vars
find_or_upload_stemcell_from "${deployment_manifest}"

upload_release "${release_dir}"
deploy

