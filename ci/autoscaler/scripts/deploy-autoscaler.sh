#!/bin/bash
set -x
set -e

echo "$BOSH_CA" > "./bosh_ca"
bosh -e $BOSH_TARGET --ca-cert ./bosh_ca alias-env vbox 
export BOSH_CLIENT=$BOSH_USERNAME 
export BOSH_CLIENT_SECRET=$BOSH_PASSWORD

cd app-autoscaler-release
# ./scripts/update
# sed -i -e 's/vm_type: default/vm_type: minimal/g' ./templates/app-autoscaler-deployment.yml
bosh create-release --force
bosh -n -e vbox upload-release --rebase

bosh -e vbox -d app-autoscaler \
     deploy -n templates/app-autoscaler-deployment.yml \
     --vars-store=bosh-lite/deployments/vars/autoscaler-deployment-vars.yml \
     -v system_domain=bosh-lite.com \
     -v cf_admin_password=$CF_ADMIN_PASSWORD \
     -v cf_admin_client_secret=$CF_ADMIN_CLIENT_SECRET \
     -v skip_ssl_validation=true 