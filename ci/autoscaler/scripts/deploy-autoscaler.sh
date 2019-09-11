#!/bin/bash
set -e

apt-get -y update
apt-get -y install dnsmasq
echo -e "\n\naddress=/.bosh-lite.com/${BOSH_TARGET}" >> /etc/dnsmasq.conf
echo 'starting dnsmasq'
dnsmasq
cp /etc/resolv.conf resolve.tmp
sed -i '1 i\nameserver 127.0.0.1' resolve.tmp
cp resolve.tmp /etc/resolv.conf

iptables -I INPUT -p tcp --dport 8443 -j ACCEPT &&iptables -I OUTPUT -p tcp --dport 8443 -j ACCEPT
iptables -t nat -I OUTPUT -p tcp -d 192.168.50.6 --dport 8443 -j DNAT --to-destination ${BOSH_TARGET}:8443

echo "${BOSH_CA}" > "./bosh_ca"
bosh -e ${BOSH_TARGET} --ca-cert ./bosh_ca alias-env vbox 
export BOSH_CLIENT=${BOSH_USERNAME} 
export BOSH_CLIENT_SECRET=${BOSH_PASSWORD}
cd app-autoscaler-release

set +e
autoscalerExists=$(bosh -e vbox releases | grep -c app-autoscaler)
if [[ $autoscalerExists == 1 ]];then
    deployedCommitHash=$(bosh -e vbox releases | grep app-autoscaler | awk -F ' ' '{print $3}' | sed 's/\+//g')
    currentCommitHash=$(git log -1 --pretty=format:"%H")
    theSame=$(echo ${currentCommitHash} | grep -c ${deployedCommitHash})
    if [[ $theSame == 1 ]];then
        echo "the app-autoscaler deployed ${deployedCommitHash} and the current ${currentCommitHash} are the same"
        exit 0
    fi
fi
set -e

uaac target https://uaa.bosh-lite.com --skip-ssl-validation
uaac token client get admin -s admin-secret
set +e
exist=$(uaac client get autoscaler_client_id | grep -c NotFound)
set -e
if [[ $exist == 0 ]];then
	uaac client update "autoscaler_client_id" \
	    --authorities "cloud_controller.read,cloud_controller.admin,uaa.resource,routing.routes.write,routing.routes.read,routing.router_groups.read"
else
	uaac client add "autoscaler_client_id" \
	--authorized_grant_types "client_credentials" \
	--authorities "cloud_controller.read,cloud_controller.admin,uaa.resource,routing.routes.write,routing.routes.read,routing.router_groups.read" \
	--secret "autoscaler_client_secret"
fi

cat >buildin.yml <<-EOF
- type: replace
  path: /instance_groups/name=asapi/jobs/name=apiserver/properties/autoscaler/api_server/service_offering_enabled
  value: false
EOF

bosh create-release --force \
&& bosh -e vbox upload-release \
&& bosh -e vbox -n -d app-autoscaler \
     deploy templates/app-autoscaler-deployment.yml \
     --vars-store=bosh-lite/deployments/vars/autoscaler-deployment-vars.yml \
     -o ./buildin.yml \
     -l ../app-autoscaler-ci/autoscaler/deployment-vars.yml \
     -v system_domain=bosh-lite.com \
     -v cf_client_id=autoscaler_client_id \
     -v cf_client_secret=autoscaler_client_secret \
     -v skip_ssl_validation=true
