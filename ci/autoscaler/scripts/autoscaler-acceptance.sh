#!/bin/bash
set -x -e

apt-get -y update
apt-get -y install dnsmasq
echo -e "\n\naddress=/.bosh-lite.com/$BOSH_TARGET" >> /etc/dnsmasq.conf
echo 'starting dnsmasq'
dnsmasq

mkdir bin
pushd bin
  curl -L 'https://cli.run.pivotal.io/stable?release=linux64-binary&source=github-rel' | tar xz
popd
export PATH=$PWD/bin:$PATH
cp /etc/resolv.conf ~/resolv.conf.new
sed -i '1 i\nameserver 127.0.0.1' ~/resolv.conf.new
cp -f ~/resolv.conf.new /etc/resolv.conf
# sed -i '1 i\nameserver 127.0.0.1' /etc/resolv.conf
cf api https://api.bosh-lite.com:443 --skip-ssl-validation
cf auth admin $CF_ADMIN_PASSWORD
# cf login -a https://api.bosh-lite.com -u admin -p admin --skip-ssl-validation

set +e
cf delete-service-broker -f autoscaler
set -e

cf create-service-broker autoscaler username password https://autoscalerservicebroker.bosh-lite.com
cf enable-service-access autoscaler

export GOPATH=$PWD/app-autoscaler-release
pushd app-autoscaler-release/src/acceptance
cat > acceptance_config.json <<EOF
{
  "api": "api.bosh-lite.com",
  "admin_user": "admin",
  "admin_password": "$CF_ADMIN_PASSWORD",
  "apps_domain": "bosh-lite.com",
  "skip_ssl_validation": true,
  "use_http": false,

  "service_name": "autoscaler",
  "service_plan": "autoscaler-free-plan",
  "aggregate_interval": 120
}
EOF
CONFIG=$PWD/acceptance_config.json ./bin/test_default -trace

popd
