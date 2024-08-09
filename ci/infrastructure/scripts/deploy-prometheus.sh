#!/bin/bash
# shellcheck disable=SC2086
set -euo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "${script_dir}/vars.source.sh"

bosh_deploy_opts=${BOSH_DEPLOY_OPTS:-}
bosh_upload_stemcell_opts="${BOSH_UPLOAD_STEMCELL_OPTS:-""}"
deployment_name="${DEPLOYMENT_NAME:-prometheus}"
bosh_cert_ca_file=${BOSH_CERT_CA_FILE:-"${HOME}/.ssh/bosh.ca.crt"}
uaa_ssl_ca_file="${UAA_SSL_CA_FILE:-$(mktemp)}"
uaa_ssl_cert_file="${UAA_SSL_CERT_FILE:-$(mktemp)}"
uaa_ssl_key_file="${UAA_SSL_KEY_FILE:-$(mktemp)}"
slack_channel="${SLACK_CHANNEL:-cf-dev-autoscaler-alerts}"
slack_webhook="${SLACK_WEBHOOK:-}"
prometheus_dir="${PROMETHEUS_DIR:-$(realpath -e ${root_dir}/../prometheus-boshrelease)}"
deployment_manifest=${DEPLOYMENT_MANIFEST:-"${prometheus_dir}/manifests/prometheus.yml"}
prometheus_ops="${prometheus_dir}/manifests/operators"
ops_files=${OPS_FILES:-"${prometheus_ops}/monitor-bosh.yml\
                        ${prometheus_ops}/enable-bosh-uaa.yml\
                        ${prometheus_ops}/configure-bosh-exporter-uaa-client-id.yml\
                        ${prometheus_ops}/monitor-cf.yml\
                        ${prometheus_ops}/enable-cf-route-registrar.yml\
                        ${prometheus_ops}/enable-grafana-uaa.yml\
                        ${prometheus_ops}/deprecated/enable-cf-loggregator-v2.yml\
                        ${prometheus_ops}/monitor-bosh-director.yml\
                        ${prometheus_ops}/alertmanager-slack-receiver.yml\
                        ${ci_dir}/operations/prometheus-customize-alerts.yml\
                        ${ci_dir}/operations/slack-receiver-template.yml\
                        ${ci_dir}/operations/prometheus-nats-tls.yml\
                        ${prometheus_ops}/alertmanager-group-by-alertname.yml\
                        "}

director_store="${bbl_state_path}/vars/director-vars-store.yml"
log "director_store = '${director_store}'"

pushd "${bbl_state_path}" > /dev/null
  eval "$(bbl print-env)"
popd > /dev/null

if [ ! -e "${bosh_cert_ca_file}" ]; then
  bosh_cert_ca_file=$(mktemp)
  echo -e "${BOSH_CA_CERT}" > $bosh_cert_ca_file
  log "bosh cert written: $bosh_cert_ca_file"
fi

step "Deploying prometheus with name '${deployment_name}' "

UAA_CLIENTS_GRAFANA_SECRET=$(credhub get -n /bosh-autoscaler/cf/uaa_clients_grafana_secret -q)
UAA_CLIENTS_CF_EXPORTER_SECRET=$(credhub get -n /bosh-autoscaler/cf/uaa_clients_cf_exporter_secret -q)
UAA_CLIENTS_FIREHOSE_EXPORTER_SECRET=$(credhub get -n /bosh-autoscaler/cf/uaa_clients_firehose_exporter_secret -q)
PROMETHEUS_CLIENT=prometheus
PROMETHEUS_CLIENT_SECRET=$(yq e '.prometheus_client_password' "${director_store}")

# Metrics for ${prometheus_dir}/manifests/operators/monitor-bosh-director.yml ops-file
BOSH_METRICS_SERVER_CLIENT_CA=$(yq e '.metrics_server_client_tls.ca' "${director_store}")
BOSH_METRICS_SERVER_CLIENT_CERT=$(yq e '.metrics_server_client_tls.certificate' "${director_store}")
BOSH_METRICS_SERVER_CLIENT_KEY=$(yq e '.metrics_server_client_tls.private_key' "${director_store}")
credhub set -n /bosh-autoscaler/prometheus/bosh_metrics_server_client_ca -t certificate -c "${BOSH_METRICS_SERVER_CLIENT_CA}" >/dev/null
credhub set -n /bosh-autoscaler/prometheus/bosh_metrics_server_client -t certificate -c "${BOSH_METRICS_SERVER_CLIENT_CERT}" -p "${BOSH_METRICS_SERVER_CLIENT_KEY}" -m /bosh-autoscaler/prometheus/bosh_metrics_server_client_ca >/dev/null

credhub get -n /bosh-autoscaler/cf/uaa_ssl -k ca          > ${uaa_ssl_ca_file}
credhub get -n /bosh-autoscaler/cf/uaa_ssl -k certificate > ${uaa_ssl_cert_file}
credhub get -n /bosh-autoscaler/cf/uaa_ssl -k private_key > ${uaa_ssl_key_file}

credhub set -n /bosh-autoscaler/prometheus/alertmanager_slack_channel -t value -v "${slack_channel}"
credhub set -n /bosh-autoscaler/prometheus/alertmanager_slack_api_url -t value -v "${slack_webhook}"


function find_or_upload_stemcell(){
  # Determine if we need to upload a stemcell at this point.
  stemcell_os=$(yq eval '.stemcells[] | select(.alias == "default").os' ${deployment_manifest})
  stemcell_version=$(yq eval '.stemcells[] | select(.alias == "default").version' ${deployment_manifest})
  stemcell_name="bosh-google-kvm-${stemcell_os}-go_agent"

  if ! bosh stemcells | grep "${stemcell_name}" >/dev/null; then
    URL="https://bosh.io/d/stemcells/${stemcell_name}"
    if [ "${stemcell_version}" != "latest" ]; then
	    URL="${URL}?v=${stemcell_version}"
    fi
    wget "$URL" -O stemcell.tgz
    bosh -n upload-stemcell $bosh_upload_stemcell_opts stemcell.tgz
  fi
}

function deploy () {
  OPS_FILES_TO_USE=""
  for OPS_FILE in ${ops_files}; do
    if [ -f "${OPS_FILE}" ]; then
      OPS_FILES_TO_USE="${OPS_FILES_TO_USE} -o ${OPS_FILE}"
    else
      echo "ERROR: could not find ops file ${OPS_FILE} in ${PWD}"
      exit 1
    fi
  done


  step "creating Bosh deployment '${deployment_name}'"
  log "using Ops files: '${OPS_FILES_TO_USE}'"
  log "deploy args: '${bosh_deploy_opts}'"

  # TODO: For Debugging: Do a `bosh interpolate` first?

  bosh -n -d "${deployment_name}" \
    deploy "${deployment_manifest}" \
    ${OPS_FILES_TO_USE} \
    ${bosh_deploy_opts} \
    --var-file bosh_ca_cert="$bosh_cert_ca_file" \
    --var-file uaa_ssl.ca="$uaa_ssl_ca_file" \
    --var-file uaa_ssl.certificate="$uaa_ssl_cert_file" \
    --var-file uaa_ssl.private_key="$uaa_ssl_key_file" \
    -v bosh_url="$BOSH_ENVIRONMENT" \
    -v bosh_ip="$(echo $BOSH_ENVIRONMENT | cut -d'/' -f3 |cut -d':' -f1)" \
    -v uaa_bosh_exporter_client_id="$PROMETHEUS_CLIENT" \
    -v uaa_bosh_exporter_client_secret="$PROMETHEUS_CLIENT_SECRET" \
    -v metrics_environment=oss \
    -v metron_deployment_name=cf \
    -v uaa_clients_cf_exporter_secret="$UAA_CLIENTS_CF_EXPORTER_SECRET" \
    -v uaa_clients_firehose_exporter_secret="$UAA_CLIENTS_FIREHOSE_EXPORTER_SECRET" \
    -v skip_ssl_verify=true \
    -v traffic_controller_external_port=4443 \
    -v system_domain="${system_domain}" \
    -v cf_deployment_name=cf \
    -v uaa_clients_grafana_secret="$UAA_CLIENTS_GRAFANA_SECRET"
}

find_or_upload_stemcell
deploy
