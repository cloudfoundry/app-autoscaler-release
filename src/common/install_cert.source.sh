#check if the cert file exists, readable and that the trust store exists and is writeable
install_cert () {
  set -euo pipefail
  local trust_store_file=$1
  local password=$2
  local cert_file=$3
  local cert_alias=$4
  echo "install_cert trust_file:\"${trust_store_file}\"  cert_file:\"${cert_file}\" cert_alias:\"${cert_alias}\""
  if test -r "${cert_file}" -a -f "${cert_file}"
  then
    if test -f "${trust_store_file}" -a -w "${trust_store_file}"
    then
      #check to see if the alias exists
      if ! manage_truststore list "${trust_store_file}" "${password}" "${cert_file}" "${cert_alias}"; then
        echo "Installing ${cert_file} with alias ${cert_alias}"
        if ! manage_truststore importcert "${trust_store_file}" "${password}" "${cert_file}" "${cert_alias}"; then
          # implement import error logic
          echo "Failed to install certificate[1]." >&2
          exit 1
        fi
      else
        echo "Certificate already installed. Replacing it"
        if ! manage_truststore delete "${trust_store_file}" "${password}" "${cert_file}" "${cert_alias}"; then
          # implement import error logic
          echo "Failed to delete existing alias, will attempt to reinstall it" >&2
        fi
  
        if ! manage_truststore importcert "${trust_store_file}" "${password}" "${cert_file}" "${cert_alias}"; then
          # implement import error logic
          echo "Failed to install certificate[2]." >&2
          exit 1
        fi
      fi
    else
      echo "Installing ${cert_file} with alias ${cert_alias}"
      if ! manage_truststore importcert "${trust_store_file}" "${password}" "${cert_file}" "${cert_alias}"; then
        # implement import error logic
        echo "Failed to install certificate[3]." >&2
        exit 1
      fi
    fi
  else
    echo "Unable to read certificate file: ${cert_file}" >&2
    exit 1
  fi
}