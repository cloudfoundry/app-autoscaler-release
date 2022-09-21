#!/bin/bash
# shellcheck disable=SC2086
set -euo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

system_domain="${SYSTEM_DOMAIN:-autoscaler.app-runtime-interfaces.ci.cloudfoundry.org}"
bbl_state_path="${BBL_STATE_PATH:-bbl-state/bbl-state}"
deployment_name="${DEPLOYMENT_NAME:-prometheus}"
bosh_cert_ca_file="${BOSH_CERT_CA_FILE:-$(mktemp)}"
uaa_ssl_ca_file="${UAA_SSL_CA_FILE:-$(mktemp)}"
uaa_ssl_cert_file="${UAA_SSL_CERT_FILE:-$(mktemp)}"
uaa_ssl_key_file="${UAA_SSL_KEY_FILE:-$(mktemp)}"
prometheus_dir="${PROMETHEUS_DIR:-$script_dir/../../../../prometheus-boshrelease}"
deployment_manifest="${prometheus_dir}/manifests/prometheus.yml"
bosh_fix_releases="${BOSH_FIX_RELEASES:-false}"
ops_files=${OPS_FILES:-"${prometheus_dir}/manifests/operators/monitor-bosh.yml\
                        ${prometheus_dir}/manifests/operators/enable-bosh-uaa.yml\
                        ${prometheus_dir}/manifests/operators/configure-bosh-exporter-uaa-client-id.yml\
                        ${prometheus_dir}/manifests/operators/monitor-cf.yml\
                        ${prometheus_dir}/manifests/operators/enable-cf-route-registrar.yml\
                        ${prometheus_dir}/manifests/operators/enable-grafana-uaa.yml\
                        ${script_dir}/../../operations/prometheus-nats-tls.yaml"}

if [[ ! -d ${bbl_state_path} ]]; then
  echo "FAILED: Did not find bbl-state folder at ${bbl_state_path}"
  echo "Make sure you have checked out the app-autoscaler-env-bbl-state repository next to the app-autoscaler-release repository to run this target or indicate its location via BBL_STATE_PATH";
  exit 1;
  fi


pushd "${bbl_state_path}" > /dev/null
  eval "$(bbl print-env)"
popd > /dev/null

echo -e  "$BOSH_CA_CERT" > $bosh_cert_ca_file
echo "Bosh cert retrived: $bosh_cert_ca_file"

echo "# Deploying prometheus with name '${deployment_name}' "

UAA_CLIENTS_GRAFANA_SECRET=$(credhub get -n /bosh-autoscaler/cf/uaa_clients_grafana_secret -q)
UAA_CLIENTS_CF_EXPORTER_SECRET=$(credhub get -n /bosh-autoscaler/cf/uaa_clients_cf_exporter_secret -q)
UAA_CLIENTS_FIREHOSE_EXPORTER_SECRET=$(credhub get -n /bosh-autoscaler/cf/uaa_clients_firehose_exporter_secret -q)
PROMETHEUS_CLIENT=prometheus
PROMETHEUS_CLIENT_SECRET=$(yq e .prometheus_client_password $bbl_state_path/vars/director-vars-store.yml)

credhub get -n /bosh-autoscaler/cf/uaa_ssl -k ca          > $uaa_ssl_ca_file
credhub get -n /bosh-autoscaler/cf/uaa_ssl -k certificate > $uaa_ssl_cert_file
credhub get -n /bosh-autoscaler/cf/uaa_ssl -k private_key > $uaa_ssl_key_file

function deploy () {
  bosh_deploy_args=""

  if [[ $bosh_fix_releases == "true" ]]; then
    bosh_fix_releases="${BOSH_FIX_RELEASES:-true}"
    bosh_deploy_args="$bosh_deploy_args --fix-releases"
  fi

  echo " - Deploy args: '${bosh_deploy_args}'"

  echo "# creating Bosh deployment '${deployment_name}'"

  bosh -n -d "${deployment_name}" \
    deploy "${deployment_manifest}" \
    ${OPS_FILES_TO_USE} \
    ${bosh_deploy_args} \
    --var-file bosh_ca_cert="$bosh_cert_ca_file" \
    --var-file uaa_ssl.ca="$uaa_ssl_ca_file" \
    --var-file uaa_ssl.certificate="$uaa_ssl_cert_file" \
    --var-file uaa_ssl.private_key="$uaa_ssl_key_file" \
    -v bosh_url="$BOSH_ENVIRONMENT" \
    -v uaa_bosh_exporter_client_id="$PROMETHEUS_CLIENT" \
    -v uaa_bosh_exporter_client_secret="$PROMETHEUS_CLIENT_SECRET" \
    -v metrics_environment=oss \
    -v metron_deployment_name=cf \
    -v uaa_clients_cf_exporter_secret="$UAA_CLIENTS_CF_EXPORTER_SECRET" \
    -v uaa_clients_firehose_exporter_secret="$UAA_CLIENTS_FIREHOSE_EXPORTER_SECRET" \
    -v skip_ssl_verify=true \
    -v traffic_controller_external_port=4443 \
    -v system_domain="$system_domain" \
    -v cf_deployment_name=cf \
    -v uaa_clients_grafana_secret="$UAA_CLIENTS_GRAFANA_SECRET"


}


pushd "${prometheus_dir}" > /dev/null
  OPS_FILES_TO_USE=""
  for OPS_FILE in ${ops_files}; do
    if [ -f "${OPS_FILE}" ]; then
      OPS_FILES_TO_USE="${OPS_FILES_TO_USE} -o ${OPS_FILE}"
    else
      echo "ERROR: could not find ops file ${OPS_FILE} in ${PWD}"
      exit 1
    fi
  done

  echo " - Using Ops files: '${OPS_FILES_TO_USE}'"
  deploy
popd > /dev/null
