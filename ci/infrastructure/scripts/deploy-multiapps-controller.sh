#!/bin/bash
# shellcheck disable=SC2086
set -euo pipefail

set -x
script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${script_dir}/vars.source.sh"
source "${script_dir}/utils.source.sh"

function create_postgres_service() {
  postgres_username=pgadmin
  postgres_database_name=multiapps_controller
  postgres_hostname=$(credhub get -n /bosh-autoscaler/postgres/postgres_host_or_ip -q)
  postgres_password=$(credhub get -n /bosh-autoscaler/postgres/pgadmin_database_password -q)

  # delete existing service
  cf cups deploy-service-database -p "{ \"uri\": \"postgres://${postgres_username}:${postgres_password}@${postgres_hostname}:5524/${postgres_database_name}?ssl=false\", \"username\": \"${postgres_username}\", \"password\": \"${postgres_password}\" }" -t postgres
}


function deploy_multiapps_controller() {
  app_name=deploy-service
  mvn -Dmaven.test.skip=true -f multiapps-controller-repo/pom.xml clean install

  pushd multiapps-controller-repo/multiapps-controller-web/target/manifests
  cf push -f manifest.yml "${app_name}"
  popd
}

function add_postrgres_security_group() {
  pushd ${CI_DIR}/infrastructure/assets
    cf create-security-group multiapps-postgres-security-group multiapps-postgres-security-group.json
    cf update-security-group multiapps-postgres-security-group multiapps-postgres-security-group.json
    cf unbind-security-group multiapps-postgres-security-group ${cf_org} ${cf_space}
    cf bind-security-group multiapps-postgres-security-group ${cf_org} --space ${cf_space}
  popd
}

function cleanup_multiapps_controller() {
  cf delete -f multiapps-controller
  cf delete-service -f deploy-service-database
}

load_bbl_vars
cf_login
cleanup_multiapps_controller
create_postgres_service
add_postrgres_security_group
deploy_multiapps_controller
