#!/bin/bash

function detect_dns {
  set +e
  HOST=$1
  PORT=$2
  retry=0
  while [[ $retry -lt 10 ]]
  do
    nc -z $HOST $PORT
    if [[ $? -eq 0 ]]
    then
      break
    fi
    sleep 5
    let retry=$retry+1
    echo "wait dns service starting: $retry"
  done
  set -e
}