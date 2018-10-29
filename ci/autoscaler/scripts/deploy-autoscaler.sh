#!/bin/bash
set -e

iptables -I INPUT -p tcp --dport 8443 -j ACCEPT &&iptables -I OUTPUT -p tcp --dport 8443 -j ACCEPT
iptables -t nat -I OUTPUT -p tcp -d 192.168.50.6 --dport 8443 -j DNAT --to-destination ${BOSH_TARGET}:8443

echo "$BOSH_CA" > "./bosh_ca"
bosh -e $BOSH_TARGET --ca-cert ./bosh_ca alias-env vbox 
export BOSH_CLIENT=$BOSH_USERNAME 
export BOSH_CLIENT_SECRET=$BOSH_PASSWORD

cd app-autoscaler-release
bosh create-release --force
bosh -n -e vbox upload-release --rebase
sed -i 's/          http_request_timeout: 5000/          http_request_timeout: 30000/g' templates/app-autoscaler-deployment.yml
bosh -n -e vbox -d app-autoscaler \
     deploy -n templates/app-autoscaler-deployment.yml \
     --vars-store=bosh-lite/deployments/vars/autoscaler-deployment-vars.yml \
     -o example/operation/bosh-dns.yml \
     -v system_domain=bosh-lite.com \
     -v cf_admin_password=$CF_ADMIN_PASSWORD \
     -v autoscaler_service_broker_password=autoscaler_service_broker_password \
     -v skip_ssl_validation=true
