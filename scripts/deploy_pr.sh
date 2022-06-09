#!/bin/bash

set -euo pipefail
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
root_dir=${script_dir}/../
current_hash=$(git log -1 --pretty=format:"%H")
version=${DEPLOYMENT_VERSION:-${current_hash}}
deploy_name=${DEPLOYMENT_NAME:-app-autoscaler-42}
system_domain=autoscaler.ci.cloudfoundry.org
bbl_state_path="${root_dir}/../app-autoscaler-env-bbl-state/bbl-state/"

cd "${bbl_state_path}"
  eval "$(bbl print-env)"
cd -

password=$(credhub get -n /bosh-autoscaler/cf/cf_admin_password -q)

#credhub delete -p "/bosh-autoscaler/${deploy_name}"

bosh -n -d ${deploy_name}\
 deploy --dry-run --no-redact\
 "${root_dir}/templates/app-autoscaler-deployment.yml"\
 -o "${root_dir}/example/operation/instance-identity-cert-from-cf.yml"\
 -o "${root_dir}/example/operation/enable-nats-tls.yml"\
 -o "${root_dir}/example/operation/loggregator-certs-from-cf.yml"\
 -o "${root_dir}/example/operation/add-extra-plan.yml"\
 -o "${root_dir}/example/operation/set-release-version.yml"\
 -o "${root_dir}/example/operation/enable-name-based-deployments.yml"\
 -v system_domain="${system_domain}"\
 -v deployment_name="${deploy_name}"\
 -v app_autoscaler_version="${version}"\
 -v admin_password="${password}"\
 -v cf_client_id=autoscaler_client_id\
 -v cf_client_secret=autoscaler_client_secret\
 -v skip_ssl_validation=true