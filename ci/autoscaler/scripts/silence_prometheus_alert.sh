#!/bin/bash

set -euo pipefail
script_dir="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${script_dir}/vars.source.sh"

DATE="date"
which gdate > /dev/null && DATE="gdate"

silence_time_mins=${SILENCE_TIME_MINS:-"45"}
alert_name=${ALERT_NAME:-"$1"}

pushd "${bbl_state_path}" > /dev/null
  eval "$(bbl print-env)"
popd > /dev/null

# shellcheck disable=SC2034
alert_manager=${ALERT_MANAGER:-"https://alertmanager.${system_domain}"}
alert_pass=${ALERT_PASS:-$(credhub get -n /bosh-autoscaler/prometheus/alertmanager_password -q)}
start_time=$(${DATE} --iso-8601=seconds --utc)
end_time=$(${DATE} -d "+ ${silence_time_mins} minutes" --iso-8601=seconds --utc)

step "silencing alert '${alert_name}' on deployment '${deployment_name}' for ${silence_time_mins} mins (${start_time} -> ${end_time})"

curl -k -s -L -X 'POST' \
  "${alert_manager}/api/v2/silences" \
  -u "admin:${alert_pass}" \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json; charset=utf-8' \
 --data-binary @- > /dev/null << EOF
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