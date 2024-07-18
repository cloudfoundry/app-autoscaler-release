#!/bin/bash
set -euo pipefail
export JDK_HOME=/var/vcap/packages/openjdk-22

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
store_file=$3

# shellcheck source=src/common/install_cert.source.sh
source "/var/vcap/packages/common/install_cert.source.sh"
install_cert "${store_file}" "123456" "${cert_file}" "${cert_alias}"
chgrp vcap "$store_file"
chmod g+r "$store_file"

## END CERTIFICATE INSTALLATION
