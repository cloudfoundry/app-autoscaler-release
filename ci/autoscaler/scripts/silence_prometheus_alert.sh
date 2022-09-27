#!/bin/bash
set -euo pipefail
script_dir="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
#shellcheck disable=SC1091
source "${script_dir}/pr-vars.source.sh"

system_domain="${SYSTEM_DOMAIN:-autoscaler.app-runtime-interfaces.ci.cloudfoundry.org}"
deployment_name="${DEPLOYMENT_NAME:-app-autoscaler}"
bbl_state_path="${BBL_STATE_PATH:-bbl-state/bbl-state}"
bbl_state_path="${BBL_STATE_PATH:-bbl-state/bbl-state}"


silence_time=${SILENCE_TIME:-"20M"}
alert_name=${ALERT_NAME:-"BOSHJobExtendedUnhealthy"}

pushd "${bbl_state_path}" > /dev/null
  eval "$(bbl print-env)"
popd > /dev/null

# shellcheck disable=SC2034
alert_manager=${ALERT_MANAGER:-"https://alertmanager.${system_domain}"}
alert_pass=${ALERT_PASS:-$(credhub get -n /bosh-autoscaler/prometheus/alertmanager_password -q)}
start_time=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
end_time=$(date "-v+${silence_time}" -u +"%Y-%m-%dT%H:%M:%SZ")
curl -k -s -f -L -X 'POST' \
  "${alert_manager}/api/v2/silences" \
  -u "admin:${alert_pass}" \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json; charset=utf-8' \
 --data-binary @- << EOF
{
  "status": {
    "state": "active"
  },
  "createdBy": "Deployment Script ${0}",
  "comment": "Automagically added for the deployment of PR ${PR_NUMBER}",
  "matchers": [
    {
      "isEqual": true,
      "isRegex": false,
      "name": "alertname",
      "value": "${alert_name}"
    },
    {
      "isEqual": true,
      "isRegex": false,
      "name": "bosh_deployment",
      "value": "${deployment_name}"
    },
    {
      "isEqual": true,
      "isRegex": false,
      "name": "bosh_name",
      "value": "bosh-autoscaler"
    },
    {
      "isEqual": true,
      "isRegex": false,
      "name": "environment",
      "value": "oss"
    }
  ],
  "startsAt": "${start_time}",
  "endsAt": "${end_time}"
}
EOF
