#!/bin/bash

function start_consul_agent {
  # Start the consul agent so we can connect to a database url provided by consul dns
  if [ -f /var/vcap/jobs/consul_agent/bin/agent_ctl ]; then
    # If consul is already running, start exits 1
    set +e
    /var/vcap/jobs/consul_agent/bin/pre-start
    chpst -u vcap:vcap /var/vcap/jobs/consul_agent/bin/agent_ctl start
    set -e
  fi
}

function wait_consul_agent {
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
    echo "wait consul agent starting: $retry"
  done
}
