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
