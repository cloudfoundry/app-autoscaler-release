#!/usr/bin/env bash
set -eu

#  Run this script to start a credhub-cli session inside the kubernetes cluster.
#
#  Example commands to interact with credhub:
#
#     credhub set -t password -n '/concourse/main/s3_access_password' -w 'supersecret'
#     credhub generate -t password -n '/concourse/main/s3_access_password' -l 128 -S

# Dockerfile to create credhub-cli image:
#
#   FROM ubuntu:20.04
#   ADD https://github.com/cloudfoundry-incubator/credhub-cli/releases/download/2.9.0/credhub-linux-2.9.0.tgz /tmp/
#   RUN tar -xzf /tmp/credhub-linux-2.9.0.tgz -C /usr/local/bin \
#       && rm /tmp/credhub-linux-2.9.0.tgz


credhub_server="https://credhub.concourse.svc.cluster.local:9000"
credhub_ca_cert="$(kubectl --namespace concourse get secret credhub-root-ca -o json | jq -r .data.certificate | base64 --decode)"
credhub_client="credhub_admin_client"
credhub_secret="$(kubectl --namespace concourse get secret credhub-admin-client-credentials -o json | jq  -r .data.password | base64 --decode)"

kubectl run credhub-cli-$(openssl rand -hex 4) \
        --rm -i -t \
        --restart=Never \
        --image=yatzek/credhub-cli:2.9.0 \
        --env="CREDHUB_SERVER=$credhub_server" \
        --env="CREDHUB_CA_CERT=$credhub_ca_cert" \
        --env="CREDHUB_CLIENT=$credhub_client" \
        --env="CREDHUB_SECRET=$credhub_secret" \
        -- /bin/bash
