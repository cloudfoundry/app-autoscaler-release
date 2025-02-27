#!/bin/bash

CERTS_DIR="/tmp/db_certs"
mkdir -p "$CERTS_DIR"

extract_service() {
  local json="$1"
  local tag="$2"
  echo "$json" | jq -r --arg tag "$tag" '
    .["user-provided"][] | select(.tags[] == $tag) | .credentials'
}

parse_uri() {
  local uri="$1"
  local field="$2"
  case "$field" in
    user) echo "$uri" | awk -F[:@] '{print substr($2, 3)}' ;;
    password) echo "$uri" | awk -F[:@] '{print $3}' ;;
    host) echo "$uri" | awk -F[@:/?] '{print $6}' ;;
    port) echo "$uri" | awk -F[@:/?] '{print $7}' ;;
    dbname) echo "$uri" | awk -F[@:/?] '{print $8}' ;;
  esac
}

persist_cert() {
  local content="$1"
  local file="$2"
  echo "$content" > "$file"
  chmod 600 "$file"
}

build_jdbc_url() {
  local host="$1" port="$2" dbname="$3" sslmode="$4"
  local client_cert="$5" client_key="$6" server_ca="$7"
  echo "jdbc:postgresql://$host:$port/$dbname?sslmode=$sslmode&sslcert=$client_cert&sslkey=$client_key&sslrootcert=$server_ca"
}

function convert_to_pk8() {
  local -r in_file="$1"
  local -r out_file="$2"
  openssl pkcs8 -topk8 -outform DER -in "$in_file" -out "$out_file" -nocrypt
  chgrp vcap "$out_file"
  chmod g+r "$out_file"
}


function run_liquibase() {
  local -r jdbcdburl="$1"
  local -r user="$2"
  local -r password="$3"
  local -r changelog="$4"

  local classpath=$(readlink -f /home/vcap/app/BOOT-INF/lib/* | tr '\n' ':')
  local java_bin="/home/vcap/app/.java-buildpack/open_jdk_jre/bin/java"

  "$java_bin" -cp "$classpath" liquibase.integration.commandline.Main \
    --url "$jdbcdburl" --username="$user" --password="$password" --driver=org.postgresql.Driver --logLevel=DEBUG --changeLogFile="$changelog" update
}

function main() {
  local json="$VCAP_SERVICES"
  local tag="policy_db"

  local service=$(extract_service "$json" "$tag")
  local uri=$(echo "$service" | jq -r '.uri')

  local host=$(parse_uri "$uri" "host")
  local port=$(parse_uri "$uri" "port")
  local dbname=$(parse_uri "$uri" "dbname")
  local sslmode="verify-ca"


  local client_cert="$CERTS_DIR/client-cert.pem"
  local client_key="$CERTS_DIR/client-key.pem"
  local client_pk8_key="$CERTS_DIR/client-key.pk8"

  local server_ca="$CERTS_DIR/server-ca.pem"


  persist_cert "$(echo "$service" | jq -r '.client_cert')" "$client_cert"
  persist_cert "$(echo "$service" | jq -r '.client_key')" "$client_key"
  persist_cert "$(echo "$service" | jq -r '.server_ca')" "$server_ca"

  convert_to_pk8 "$client_key" "$client_pk8_key"

  JDBCDBURL=$(build_jdbc_url "$host" "$port" "$dbname" "$sslmode" "$client_cert" "$client_pk8_key" "$server_ca")
  PASSWORD=$(parse_uri "$uri" "password")
  USER=$(parse_uri "$uri" "user")


  run_liquibase "$JDBCDBURL" "$USER" "$PASSWORD" "BOOT-INF/classes/api.db.changelog.yml"
  run_liquibase "$JDBCDBURL" "$USER" "$PASSWORD" "BOOT-INF/classes/servicebroker.db.changelog.yaml"
}


main

