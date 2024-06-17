#!/bin/bash

set -euo pipefail

manifest_path="${MANIFEST_PATH:-../templates/app-autoscaler.yml}"
operation_dir_path="${OPERATION_DIR_PATH:-../example/operation}"
scalingengine_instance_group="${SCALINGENGINE_INSTANCE_GROUP:-scalingengine}"
scheduler_instance_group="${SCHEDULER_INSTANCE_GROUP:-scheduler}"
metricsforwarder_instance_group="${METICSFORWARDER_INSTANCE_GROUP:-metricsforwarder}"
operator_instance_group="${OPERATOR_INSTANCE_GROUP:-operator}"
eventgenerator_instance_group="${EVENTGENERATOR_INSTANCE_GROUP:-eventgenerator}"

# this is a really basic check to validate that the peristent disk value is set.
# FIXME we need a much better way of doing this.
ACTUAL=$(bosh int "$manifest_path" | yq e '.instance_groups[] | select(.jobs[].name == "postgres").persistent_disk_type' -)
if [ "${ACTUAL}" != "null" ]; then
	echo "FAILED: default has no persistent disk"
	exit 1
fi

echo "$operation_dir_path/postgres-persistent-disk.yml"
ACTUAL=$(bosh int -o "$operation_dir_path/postgres-persistent-disk.yml" "$manifest_path" | yq e '.instance_groups[] | select(.jobs[].name == "postgres").persistent_disk_type' -)
if [ "${ACTUAL}" != "10GB" ]; then
	echo "FAILED: Expected 10GB to be set as the persistent disk size"
	exit 1
fi


COMPONENTS="scheduler eventgenerator metricsforwarder operator scalingengine"
for COMPONENT in $COMPONENTS; do
  ACTUAL=$(bosh int "$manifest_path" | yq e ".variables[] | select(.name == \"autoscaler_${COMPONENT}_health_password\").type" -)
  if [ "${ACTUAL}" != "password" ]; then
	  echo "FAILED: autoscaler_${COMPONENT}_health_password should be set"
	  exit 1
  fi

done

for COMPONENT in $COMPONENTS; do
  echo "$operation_dir_path/disable-basicauth-on-health-endpoints.yml - ${COMPONENT}"
  ACTUAL=$(bosh int -o "$operation_dir_path/disable-basicauth-on-health-endpoints.yml" "$manifest_path" | yq e ".variables[] | select(.name == \"autoscaler_${COMPONENT}_health_password\").type" -)
  if [ "${ACTUAL}" != "" ]; then
	  echo "FAILED: autoscaler_${COMPONENT}_health_password should not be set"
	  exit 1
  fi
done


ACTUAL=$(bosh int "$manifest_path" | yq e ".instance_groups[] | select(.name == \"$scalingengine_instance_group\") | .jobs[] | select(.name == \"scalingengine\").properties.autoscaler.scalingengine.health.password" -)
if [ "${ACTUAL}" != "((autoscaler_scalingengine_health_password))" ]; then
	echo "FAILED: $scalingengine_instance_group/scalingengine health password should be set"
	exit 1
fi

ACTUAL=$(bosh int "$manifest_path" | yq e ".instance_groups[] | select(.name == \"$scheduler_instance_group\") | .jobs[] | select(.name == \"scheduler\").properties.autoscaler.scheduler.health.password" -)
if [ "${ACTUAL}" != "((autoscaler_scheduler_health_password))" ]; then
	echo "FAILED: $scheduler_instance_group/scheduler health password should be set"
	exit 1
fi

ACTUAL=$(bosh int "$manifest_path" | yq e ".instance_groups[] | select(.name == \"$operator_instance_group\") | .jobs[] | select(.name == \"operator\").properties.autoscaler.operator.health.password" -)
if [ "${ACTUAL}" != "((autoscaler_operator_health_password))" ]; then
	echo "FAILED: $operator_instance_group/operator health password should be set"
	exit 1
fi

ACTUAL=$(bosh int "$manifest_path" | yq e ".instance_groups[] | select(.name == \"$eventgenerator_instance_group\") | .jobs[] | select(.name == \"eventgenerator\").properties.autoscaler.eventgenerator.health.password" -)
if [ "${ACTUAL}" != "((autoscaler_eventgenerator_health_password))" ]; then
	echo "FAILED: asmetrics/eventgenerator health password should be set"
	exit 1
fi

ACTUAL=$(bosh int "$manifest_path" | yq e ".instance_groups[] | select(.name == \"$metricsforwarder_instance_group\") | .jobs[] | select(.name == \"metricsforwarder\").properties.autoscaler.metricsforwarder.health.password" -)
if [ "${ACTUAL}" != "((autoscaler_metricsforwarder_health_password))" ]; then
	echo "FAILED: $metricsforwarder_instance_group/metricsforwarder health password should be set"
	exit 1
fi







ACTUAL=$(bosh int -o "$operation_dir_path/disable-basicauth-on-health-endpoints.yml" "$manifest_path" | yq e ".instance_groups[] | select(.name == \"$scalingengine_instance_group\") | .jobs[] | select(.name == \"scalingengine\").properties.autoscaler.scalingengine.health.password" -)
if [ "${ACTUAL}" != "null" ]; then
	echo "FAILED: $scalingengine_instance_group/scalingengine health password should not be set"
	exit 1
fi

ACTUAL=$(bosh int -o "$operation_dir_path/disable-basicauth-on-health-endpoints.yml" "$manifest_path" | yq e ".instance_groups[] | select(.name == \"$scheduler_instance_group\") | .jobs[] | select(.name == \"scheduler\").properties.autoscaler.scheduler.health.password" -)
if [ "${ACTUAL}" != "null" ]; then
	echo "FAILED: $scheduler_instance_group/scheduler health password should not be set"
	exit 1
fi

ACTUAL=$(bosh int -o "$operation_dir_path/disable-basicauth-on-health-endpoints.yml" "$manifest_path" | yq e ".instance_groups[] | select(.name == \"$operator_instance_group\") | .jobs[] | select(.name == \"operator\").properties.autoscaler.operator.health.password" -)
if [ "${ACTUAL}" != "null" ]; then
	echo "FAILED: $operator_instance_group/operator health password should not be set"
	exit 1
fi

ACTUAL=$(bosh int -o "$operation_dir_path/disable-basicauth-on-health-endpoints.yml" "$manifest_path" | yq e ".instance_groups[] | select(.name == \"$eventgenerator_instance_group\") | .jobs[] | select(.name == \"eventgenerator\").properties.autoscaler.eventgenerator.health.password" -)
if [ "${ACTUAL}" != "null" ]; then
	echo "FAILED: $eventgenerator_instance_group/eventgenerator health password should not be set"
	exit 1
fi

ACTUAL=$(bosh int -o "$operation_dir_path/disable-basicauth-on-health-endpoints.yml" "$manifest_path" | yq e ".instance_groups[] | select(.name == \"$metricsforwarder_instance_group\") | .jobs[] | select(.name == \"metricsforwarder\").properties.autoscaler.metricsforwarder.health.password" -)
if [ "${ACTUAL}" != "null" ]; then
	echo "FAILED: $metricsforwarder_instance_group/metricsforwarder health password should not be set"
	exit 1
fi

