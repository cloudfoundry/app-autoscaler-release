#!/bin/bash
set -x
set -e

echo "$BOSH_CA" > "./bosh_ca"
bosh -e $BOSH_TARGET --ca-cert ./bosh_ca alias-env vbox 
export BOSH_CLIENT=$BOSH_USERNAME 
export BOSH_CLIENT_SECRET=$BOSH_PASSWORD
bosh -n -e vbox delete-deployment -d cf
cd cf-deployment
bosh -n -e vbox upload-stemcell https://bosh.io/d/stemcells/bosh-warden-boshlite-ubuntu-trusty-go_agent
bosh -n -e vbox update-cloud-config iaas-support/bosh-lite/cloud-config.yml
bosh -e vbox -d cf deploy -n cf-deployment.yml \
  -o operations/bosh-lite.yml \
  -o operations/use-compiled-releases.yml \
  --vars-store ../app-autoscaler-ci/autoscaler/deployment-vars.yml \
  -v system_domain=bosh-lite.com \
  -v cf_admin_password=$CF_ADMIN_PASSWORD \
  -v uaa_admin_client_secret=$CF_ADMIN_CLIENT_SECRET