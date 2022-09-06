#!/bin/bash

# enable deep monitoring by relying on the rule set-up
export DT_MONITOR="true"

# set the minimum number of file handles to a sensible 16k(base2)
minimum_file_handles=16384

mkdir -p /var/vcap/sys/log
script_name=$(basename "$0")
exec > >(tee -a >(logger -p user.info -t "vcap.${script_name}.stdout") | awk -W interactive '{ system("echo -n [$(date +\"%Y-%m-%d %H:%M:%S%z\")]"); print " " $0 }' >> "/var/vcap/sys/log/${script_name}.log")
exec 2> >(tee -a >(logger -p user.error -t "vcap.${script_name}.stderr") | awk -W interactive '{ system("echo -n [$(date +\"%Y-%m-%d %H:%M:%S%z\")]"); print " " $0 }' >> "/var/vcap/sys/log/${script_name}.err.log")

pid_guard() {
  echo "------------ STARTING ${script_name} at $(date) --------------" | tee /dev/stderr
  pidfile=$1
  name=$2

  if [ -f "$pidfile" ]; then
    pid="$(head -1 "${pidfile}")"
    echo "pidno ${pid}"
    if [ -n "${pid}" ] && [ -e "/proc/$pid" ] && grep -q "/var/vcap/packages/$name" "/proc/$pid/cmdline"; then
      echo "${name} is already running, please stop it first"
      exit 1
    fi

    echo "Removing stale pidfile..."
    rm "${pidfile}"
  fi
}

wait_pidfile() {
  pidfile=$1
  try_kill=$2
  timeout=${3:-0}
  force=${4:-0}
  countdown=$(( timeout * 10 ))

  if [ -f "$pidfile" ]; then
    pid=$(head -1 "$pidfile")

    if [ -z "$pid" ]; then
      echo "Unable to get pid from $pidfile"
      exit 1
    fi

    if [ -e "/proc/${pid}" ]; then
      if [ "$try_kill" = "1" ]; then
        echo "Killing $pidfile: $pid "
        kill "${pid}"
      fi
      while [ -e "/proc/${pid}" ]; do
        sleep 0.1
        [ "$countdown" != '0' ] && [ $(( countdown % 10 )) = '0' ] && echo -n .
        if [ "${timeout}" -gt 0 ]; then
          if [ $countdown -eq 0 ]; then
            if [ "$force" = "1" ]; then
              echo -ne "\nKill timed out, using kill -9 on $pid... "
              kill -9 "$pid"
              sleep 0.5
            fi
            break
          else
            countdown=$(( countdown - 1 ))
          fi
        fi
      done
      if [ -e "/proc/$pid" ]; then
        echo "Timed Out"
      else
        echo "Stopped"
      fi
    else
      echo "Process $pid is not running"
    fi

    rm -f "$pidfile"
  else
    echo "Pidfile $pidfile doesn't exist"
  fi
}

kill_and_wait() {
  pidfile=$1
  # Monit default timeout for start/stop is 30s
  # Append 'with timeout {n} seconds' to monit start/stop program configs
  timeout=${2:-25}
  force=${3:-1}

  wait_pidfile "${pidfile}" 1 "${timeout}" "${force}"
}

set_ulimits(){
  #Set the minimum file handle limit to $minimum_file_handles
  file_limit=$(ulimit -n)
  if [ "${file_limit}" == "unlimited"  ]; then
     return
  elif [ "${file_limit}" -lt "${minimum_file_handles}" ]; then
    ulimit -n ${minimum_file_handles}
  fi
}