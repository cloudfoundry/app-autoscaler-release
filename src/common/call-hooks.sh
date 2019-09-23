#!/bin/bash
job=$1
phase=$2
JOB_DIR=/var/vcap/jobs/${job}
LOG_DIR=/var/vcap/sys/log/${job}
HOOK_LOG_OUT=${LOG_DIR}/hooks.stdout.log
HOOK_LOG_ERR=${LOG_DIR}/hooks.stderr.log

echo "Invoking ${phase} hook"
echo "------------ STARTING ${phase} at `date` --------------" >> ${HOOK_LOG_OUT} 2>>${HOOK_LOG_ERR}
${JOB_DIR}/bin/hooks/${phase}.sh >> ${HOOK_LOG_OUT} 2>>${HOOK_LOG_ERR}
retcode=$?
echo "------------ COMPLETED ${phase} at `date` with return code $retcode--------------" >> ${HOOK_LOG_OUT} 2>>${HOOK_LOG_ERR}
echo " ${phase} hook invoked with return code $retcode"
exit 0
