#!/bin/bash
set -e

iptables -I INPUT -p tcp --dport 8443 -j ACCEPT && iptables -I OUTPUT -p tcp --dport 8443 -j ACCEPT
iptables -t nat -I OUTPUT -p tcp -d 192.168.50.6 --dport 8443 -j DNAT --to-destination ${BOSH_TARGET}:8443

echo "$BOSH_CA" > "./bosh_ca"
bosh -e $BOSH_TARGET --ca-cert ./bosh_ca alias-env vbox 
export BOSH_CLIENT=${BOSH_USERNAME}
export BOSH_CLIENT_SECRET=${BOSH_PASSWORD}
# bosh -n -e vbox delete-deployment -d cf
cd cf-deployment
# git reset --hard 28451111afdbacec8e3f451e98347c0cf8d11cb0
set +e
STEMCELL_VERSION=$(yq read ./cf-deployment.yml stemcells[0].version)
STEMCELL_OS=$(yq read ./cf-deployment.yml stemcells[0].os)
STEMCELL_URL="https://bosh.io/d/stemcells/bosh-warden-boshlite-${STEMCELL_OS}-go_agent?v=${STEMCELL_VERSION}"
STEMCELL_EXISTS=$(bosh -e vbox stemcells | grep -c ${STEMCELL_VERSION})
set -e
if [[ $STEMCELL_EXISTS == 0 ]];then
	bosh -n -e $BOSH_TARGET upload-stemcell ${STEMCELL_URL}
fi
bosh -n -e $BOSH_TARGET update-cloud-config iaas-support/bosh-lite/cloud-config.yml

bosh -n -e $BOSH_TARGET -d cf deploy cf-deployment.yml \
-o operations/bosh-lite.yml \
-o operations/use-compiled-releases.yml \
--vars-store ../app-autoscaler-ci/autoscaler/deployment-vars.yml \
-v system_domain=bosh-lite.com \
-v cf_admin_password=cf_admin_password \
-v uaa_admin_client_secret=admin-secret


