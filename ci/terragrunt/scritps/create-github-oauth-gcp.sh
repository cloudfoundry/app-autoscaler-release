#!/usr/bin/env bash
set -euo pipefail

# Enter you credentials here
# your Client ID and Secret
id="some_id"
secret="some_secret"

if [ -z "${id}" ] || [ -z "${secret}" ]; then
  echo "ERROR: Please enter your credentials on the top of this script"
  exit 1
fi

if [ ! -s "./config.yaml" ]; then
    echo "ERROR: Please 'cd' to your folder with config.yaml and run ./dr/dr_restore.sh"
    exit 1
fi

secret_id="$(yq .gke_name config.yaml)-concourse-github-oauth"
secret_region="$(yq .region config.yaml)"
project="$(yq .project config.yaml)"

echo "Creating the gcp secret ${secret_id} in project ${project} within region ${secret_region}..."

gcloud secrets create "${secret_id}" \
 --replication-policy="user-managed" \
 --locations="${secret_region}" \
 --project="${project}"

echo "Creating secret version..."

printf "id: %s\nsecret: %s" ${id} ${secret} | \
gcloud secrets versions add "${secret_id}" --data-file=- --project="${project}"

echo "Done"
