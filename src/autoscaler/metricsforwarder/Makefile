DEPLOYMENT_NAME := autoscaler-2913
VM_NAME := $(shell bosh -d autoscaler-2913 vms --json | jq ".Tables | .[0] | .Rows | .[] | .instance" -r | grep metricsforwarder)

pull-files-for-poc:
	mkdir -p assets/certs/policy_db assets/certs/storedprocedure_db assets/certs/loggr-syslog-agent

	@echo "Pulling files from $(VM_NAME)..."

	bosh -d $(DEPLOYMENT_NAME) scp $(VM_NAME):/var/vcap/jobs/metricsforwarder/config/metricsforwarder.yml assets/metricsforwarder.yml

	bosh -d $(DEPLOYMENT_NAME) scp $(VM_NAME):/var/vcap/jobs/metricsforwarder/config/certs/policy_db/crt    assets/certs/policy_db/.
	bosh -d $(DEPLOYMENT_NAME) scp $(VM_NAME):/var/vcap/jobs/metricsforwarder/config/certs/policy_db/ca.crt assets/certs/policy_db/.
	bosh -d $(DEPLOYMENT_NAME) scp $(VM_NAME):/var/vcap/jobs/metricsforwarder/config/certs/policy_db/key    assets/certs/policy_db/.

	bosh -d $(DEPLOYMENT_NAME) scp $(VM_NAME):/var/vcap/jobs/metricsforwarder/config/certs/storedprocedure_db/crt	 assets/certs/storedprocedure_db/.
	bosh -d $(DEPLOYMENT_NAME) scp $(VM_NAME):/var/vcap/jobs/metricsforwarder/config/certs/storedprocedure_db/ca.crt assets/certs/storedprocedure_db/.
	bosh -d $(DEPLOYMENT_NAME) scp $(VM_NAME):/var/vcap/jobs/metricsforwarder/config/certs/storedprocedure_db/key    assets/certs/storedprocedure_db/.

	# loggr-syslog-agent certs
	bosh -d $(DEPLOYMENT_NAME) scp $(VM_NAME):/var/vcap/jobs/loggr-syslog-agent/config/certs/syslog_agent.crt   assets/certs/loggr-syslog-agent/.
	bosh -d $(DEPLOYMENT_NAME) scp $(VM_NAME):/var/vcap/jobs/loggr-syslog-agent/config/certs/syslog_agent.key   assets/certs/loggr-syslog-agent/.
	bosh -d $(DEPLOYMENT_NAME) scp $(VM_NAME):/var/vcap/jobs/loggr-syslog-agent/config/certs/loggregator_ca.crt assets/certs/loggr-syslog-agent/.

	cp assets/metricsforwarder.yml metricsforwarder.yml
	sed -i ''  's|\/var\/vcap\/jobs\/metricsforwarder\/config|\/home\/vcap\/app/assets|g' metricsforwarder.yml