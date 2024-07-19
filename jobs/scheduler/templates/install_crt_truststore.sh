#!/bin/bash
set -euo pipefail
export JDK_HOME=/var/vcap/packages/openjdk-22

function manage_truststore() {
  set -euo pipefail
  local operation=$1
  local trust_store_file=$2
  local password=$3
  local cert_file=$4
  local cert_alias=$5

  # shellcheck disable=SC2155
  local number_of_certs=$(grep -c 'END CERTIFICATE' "${cert_file}")
  #file prefix for the split
  file_prefix_split=cert_
  # extract the certificate chain file is split into multiple files
  csplit "${cert_file}" '/-----END CERTIFICATE-----/1' -f "${file_prefix_split}" -q

  # loop over the files to put the certificates in the store
  for ((i=number_of_certs-1; i>=0;i--))
  do
    alias="${cert_alias}$i"
    if [ $i -eq 0 ]; then
      alias=${cert_alias}
    fi
    keytool "-${operation}" -file "${file_prefix_split}0$i" -keystore "${trust_store_file}" -storeType pkcs12 -storepass "${password}" -noprompt -alias "${alias}" >/dev/null 2>&1
  done
}


cert_alias=$1
cert_file=$2
store_file=$3

# shellcheck disable=SC1091
source "/var/vcap/packages/common/install_cert.source.sh"
install_cert "${store_file}" "123456" "/var/vcap/jobs/scheduler/config/certs/${cert_file}" "${cert_alias}"
chgrp vcap "$store_file"
chmod g+r "$store_file"
