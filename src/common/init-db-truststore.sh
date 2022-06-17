#!/bin/bash

## BEGIN CERTIFICATE INSTALLATION
#define JDK_HOME
export JDK_HOME=/var/vcap/packages/openjdk-11

CERT_ALIAS=$1
CERT_FILE=$2

#define the key store
TRUST_STORE_FILE="/var/vcap/data/certs/${CERT_ALIAS}/cacerts"

#define the password
PASSWORD=123456

#create directory for trust store
mkdir -p "/var/vcap/data/certs/${CERT_ALIAS}"

manage_truststore () {
    operation=$1
    $JDK_HOME/bin/keytool "-$operation" -file "${CERT_FILE}" -keystore "${TRUST_STORE_FILE}" -storeType pkcs12 -storepass "${PASSWORD}" -noprompt -alias "${CERT_ALIAS}" >/dev/null 2>&1
}

#check if the cert file exists, readable and that the trust store exists and is writeable
if test -r "${CERT_FILE}" -a -f "${CERT_FILE}"
then
  if test -f "${TRUST_STORE_FILE}" -a -w "${TRUST_STORE_FILE}"
  then
    #check to see if the alias exists

    if ! manage_truststore list; then
      echo "Installing $CERT_FILE with alias $CERT_ALIAS"
      if ! manage_truststore importcert; then
        # implement import error logic
        echo "Failed to install certificate[1]."
      fi
    else
      echo "Certificate already installed. Replacing it"
      if ! manage_truststore delete; then
        # implement import error logic
        echo "Failed to delete existing alias, will attempt to reinstall it"
      fi

      if ! manage_truststore importcert; then
        # implement import error logic
        echo "Failed to install certificate[2]."
      fi
    fi
  else
    echo "Installing $CERT_FILE with alias $CERT_ALIAS"

    if ! manage_truststore importcert; then
      # implement import error logic
      echo "Failed to install certificate[3]."
    fi
  fi
else
  echo "Unable to read certificate file: $CERT_FILE"
fi
## END CERTIFICATE INSTALLATION

