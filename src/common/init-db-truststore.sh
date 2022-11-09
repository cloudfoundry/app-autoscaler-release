#!/bin/bash
set -euo pipefail
export JDK_HOME=/var/vcap/packages/openjdk-17

manage_truststore () {
    set -euo pipefail
    local operation=$1
    local trust_store_file=$2
    local password=$3
    local cert_file=$4
    local cert_alias=$5
    $JDK_HOME/bin/keytool "-$operation" -file "${cert_file}" -keystore "${trust_store_file}" -storeType pkcs12 -storepass "${password}" -noprompt -alias "${cert_alias}" >/dev/null 2>&1
}

cert_alias=$1
cert_file=$2

#create directory for trust store
mkdir -p "/var/vcap/data/certs/${cert_alias}"

source "/var/vcap/packages/common/install_cert.source.sh"
install_cert "/var/vcap/data/certs/${cert_alias}/cacerts" "123456" "${cert_file}" "${cert_alias}"

## END CERTIFICATE INSTALLATION

