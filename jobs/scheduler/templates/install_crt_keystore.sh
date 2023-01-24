#!/bin/bash
set -euo pipefail

CERT_ALIAS=$1
CERT=$2
CERT_KEY=$3
KEY_STORE=$4

#define the certificate to import
CERT_FILE="/var/vcap/jobs/scheduler/config/certs/$CERT"

#define the private key to import
KEY_FILE="/var/vcap/jobs/scheduler/config/certs/$CERT_KEY"

#install ssl certificate into key store
openssl pkcs12 -export -name "$CERT_ALIAS" -in "$CERT_FILE" -inkey "$KEY_FILE" -out "$KEY_STORE" -password pass:123456
chgrp vcap "$KEY_STORE"
chmod g+r "$KEY_STORE"
