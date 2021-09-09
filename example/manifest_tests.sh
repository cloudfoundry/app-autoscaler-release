#!/bin/bash

set -euo pipefail

# this is a really basic check to validate that the peristent disk value is set.
# FIXME we need a much better way of doing this.
echo "no ops files"
ACTUAL=$(bosh int ../templates/app-autoscaler-deployment.yml | yq e '.instance_groups[] | select(.jobs[].name == "postgres").persistent_disk_type' -)
if [ "${ACTUAL}" != "null" ]; then
	echo "FAILED: default has no persistent disk"
	exit 1
fi

echo "operation/postgres-persistent-disk.yml"
ACTUAL=$(bosh int -o operation/postgres-persistent-disk.yml ../templates/app-autoscaler-deployment.yml | yq e '.instance_groups[] | select(.jobs[].name == "postgres").persistent_disk_type' -)
if [ "${ACTUAL}" != "10GB" ]; then
	echo "FAILED: Expected 10GB to be set as the persistent disk size"
	exit 1
fi


COMPONENTS="scheduler eventgenerator metricsforwarder metricsgateway metricsserver operator scalingengine"
for COMPONENT in $COMPONENTS; do
  echo "no ops files - ${COMPONENT}"
  ACTUAL=$(bosh int ../templates/app-autoscaler-deployment.yml | yq e ".variables[] | select(.name == \"autoscaler_${COMPONENT}_health_password\").type" -)
  if [ "${ACTUAL}" != "password" ]; then
	  echo "FAILED: autoscaler_${COMPONENT}_health_password should be set"
	  exit 1
  fi

done

for COMPONENT in $COMPONENTS; do
  echo "operation/disable-basicauth-on-health-endpoints.yml - ${COMPONENT}"
  ACTUAL=$(bosh int -o operation/disable-basicauth-on-health-endpoints.yml ../templates/app-autoscaler-deployment.yml | yq e ".variables[] | select(.name == \"autoscaler_${COMPONENT}_health_password\").type" -)
  if [ "${ACTUAL}" != "" ]; then
	  echo "FAILED: autoscaler_${COMPONENT}_health_password should not be set"
	  exit 1
  fi
done


ACTUAL=$(bosh int ../templates/app-autoscaler-deployment.yml | yq e '.instance_groups[] | select(.name == "asactors") | .jobs[] | select(.name == "scalingengine").properties.autoscaler.scalingengine.health.password' -)
if [ "${ACTUAL}" != "((autoscaler_scalingengine_health_password))" ]; then
	echo "FAILED: asactors/scalingengine health password should be set"
	exit 1
fi

ACTUAL=$(bosh int ../templates/app-autoscaler-deployment.yml | yq e '.instance_groups[] | select(.name == "asactors") | .jobs[] | select(.name == "scheduler").properties.autoscaler.scheduler.health.password' -)
if [ "${ACTUAL}" != "((autoscaler_scheduler_health_password))" ]; then
	echo "FAILED: asactors/operator health password should be set"
	exit 1
fi

ACTUAL=$(bosh int ../templates/app-autoscaler-deployment.yml | yq e '.instance_groups[] | select(.name == "asactors") | .jobs[] | select(.name == "operator").properties.autoscaler.operator.health.password' -)
if [ "${ACTUAL}" != "((autoscaler_operator_health_password))" ]; then
	echo "FAILED: asmetrics/scheduler health password should be set"
	exit 1
fi

ACTUAL=$(bosh int ../templates/app-autoscaler-deployment.yml | yq e '.instance_groups[] | select(.name == "asmetrics") | .jobs[] | select(.name == "metricsserver").properties.autoscaler.metricsserver.health.password' -)
if [ "${ACTUAL}" != "((autoscaler_metricsserver_health_password))" ]; then
	echo "FAILED: asmetrics/metricsserver health password should be set"
	exit 1
fi

ACTUAL=$(bosh int ../templates/app-autoscaler-deployment.yml | yq e '.instance_groups[] | select(.name == "asmetrics") | .jobs[] | select(.name == "eventgenerator").properties.autoscaler.eventgenerator.health.password' -)
if [ "${ACTUAL}" != "((autoscaler_eventgenerator_health_password))" ]; then
	echo "FAILED: asmetrics/eventgenerator health password should be set"
	exit 1
fi

ACTUAL=$(bosh int ../templates/app-autoscaler-deployment.yml | yq e '.instance_groups[] | select(.name == "asnozzle") | .jobs[] | select(.name == "metricsgateway").properties.autoscaler.metricsgateway.health.password' -)
if [ "${ACTUAL}" != "((autoscaler_metricsgateway_health_password))" ]; then
	echo "FAILED: asnozzle/metricsgateway health password should be set"
	exit 1
fi

ACTUAL=$(bosh int ../templates/app-autoscaler-deployment.yml | yq e '.instance_groups[] | select(.name == "asapi") | .jobs[] | select(.name == "metricsforwarder").properties.autoscaler.metricsforwarder.health.password' -)
if [ "${ACTUAL}" != "((autoscaler_metricsforwarder_health_password))" ]; then
	echo "FAILED: asapi/metricsforwarder health password should be set"
	exit 1
fi







ACTUAL=$(bosh int -o operation/disable-basicauth-on-health-endpoints.yml ../templates/app-autoscaler-deployment.yml | yq e '.instance_groups[] | select(.name == "asactors") | .jobs[] | select(.name == "scalingengine").properties.autoscaler.scalingengine.health.password' -)
if [ "${ACTUAL}" != "null" ]; then
	echo "FAILED: asactors/scalingengine health password should not be set"
	exit 1
fi

ACTUAL=$(bosh int -o operation/disable-basicauth-on-health-endpoints.yml ../templates/app-autoscaler-deployment.yml | yq e '.instance_groups[] | select(.name == "asactors") | .jobs[] | select(.name == "scheduler").properties.autoscaler.scheduler.health.password' -)
if [ "${ACTUAL}" != "null" ]; then
	echo "FAILED: asactors/scheduler health password should not be set"
	exit 1
fi

ACTUAL=$(bosh int -o operation/disable-basicauth-on-health-endpoints.yml ../templates/app-autoscaler-deployment.yml | yq e '.instance_groups[] | select(.name == "asactors") | .jobs[] | select(.name == "operator").properties.autoscaler.operator.health.password' -)
if [ "${ACTUAL}" != "null" ]; then
	echo "FAILED: asactors/operator health password should not be set"
	exit 1
fi

ACTUAL=$(bosh int -o operation/disable-basicauth-on-health-endpoints.yml ../templates/app-autoscaler-deployment.yml | yq e '.instance_groups[] | select(.name == "asmetrics") | .jobs[] | select(.name == "metricsserver").properties.autoscaler.metricsserver.health.password' -)
if [ "${ACTUAL}" != "null" ]; then
	echo "FAILED: asmetrics/metricsserver health password should not be set"
	exit 1
fi

ACTUAL=$(bosh int -o operation/disable-basicauth-on-health-endpoints.yml ../templates/app-autoscaler-deployment.yml | yq e '.instance_groups[] | select(.name == "asmetrics") | .jobs[] | select(.name == "eventgenerator").properties.autoscaler.eventgenerator.health.password' -)
if [ "${ACTUAL}" != "null" ]; then
	echo "FAILED: asmetrics/eventgenerator health password should not be set"
	exit 1
fi

ACTUAL=$(bosh int -o operation/disable-basicauth-on-health-endpoints.yml ../templates/app-autoscaler-deployment.yml | yq e '.instance_groups[] | select(.name == "asnozzle") | .jobs[] | select(.name == "metricsgateway").properties.autoscaler.metricsgateway.health.password' -)
if [ "${ACTUAL}" != "null" ]; then
	echo "FAILED: asnozzle/metricsgatway health password should not be set"
	exit 1
fi

ACTUAL=$(bosh int -o operation/disable-basicauth-on-health-endpoints.yml ../templates/app-autoscaler-deployment.yml | yq e '.instance_groups[] | select(.name == "asapi") | .jobs[] | select(.name == "metricsforwarder").properties.autoscaler.metricsforwarder.health.password' -)
if [ "${ACTUAL}" != "null" ]; then
	echo "FAILED: asapi/metricsforwarder health password should not be set"
	exit 1
fi

