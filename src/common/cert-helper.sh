#!/bin/bash

function convert_to_pk8() {
  local -r in_file="$1"
  local -r out_file="$2"
  openssl pkcs8 -topk8 -outform DER -in "$in_file" -out "$out_file" -nocrypt
  chgrp vcap "$out_file"
  chmod g+r "$out_file"
}