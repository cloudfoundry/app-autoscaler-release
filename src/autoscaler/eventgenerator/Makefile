.PHONY: fetch-config
fetch-config: start-metricsforwarder-vm
	# how to define variables in deployment name
	mkdir -p assets/certs/policy_db assets/certs/storedprocedure_db assets/certs/syslog_client

	echo "POSTGRES IP: $(POSTGRES_IP)"
	echo "LOG_CACHE IP: $(LOG_CACHE_IP)"

	@echo "Pulling metricforwarder config from $(METIRCSFORWARDER_VM)..."
	bosh -d $(DEPLOYMENT_NAME) scp $(METIRCSFORWARDER_VM):/var/vcap/jobs/metricsforwarder/config/metricsforwarder.yml assets/metricsforwarder.yml

	@echo "Pulling policy db certs from $(METIRCSFORWARDER_VM)..."
	bosh -d $(DEPLOYMENT_NAME) scp $(METIRCSFORWARDER_VM):/var/vcap/jobs/metricsforwarder/config/certs/policy_db/ca.crt assets/certs/policy_db/.
	bosh -d $(DEPLOYMENT_NAME) scp $(METIRCSFORWARDER_VM):/var/vcap/jobs/metricsforwarder/config/certs/policy_db/crt    assets/certs/policy_db/.
	bosh -d $(DEPLOYMENT_NAME) scp $(METIRCSFORWARDER_VM):/var/vcap/jobs/metricsforwarder/config/certs/policy_db/key    assets/certs/policy_db/.

	@echo "Pulling storeprocedure db certs from $(METIRCSFORWARDER_VM)..."
	bosh -d $(DEPLOYMENT_NAME) scp $(METIRCSFORWARDER_VM):/var/vcap/jobs/metricsforwarder/config/certs/storedprocedure_db/ca.crt assets/certs/storedprocedure_db/.
	bosh -d $(DEPLOYMENT_NAME) scp $(METIRCSFORWARDER_VM):/var/vcap/jobs/metricsforwarder/config/certs/storedprocedure_db/crt	 assets/certs/storedprocedure_db/.
	bosh -d $(DEPLOYMENT_NAME) scp $(METIRCSFORWARDER_VM):/var/vcap/jobs/metricsforwarder/config/certs/storedprocedure_db/key    assets/certs/storedprocedure_db/.

	@echo "Pulling syslog-client certs from $(METIRCSFORWARDER_VM)..."
	bosh -d $(DEPLOYMENT_NAME) scp $(METIRCSFORWARDER_VM):/var/vcap/jobs/metricsforwarder/config/certs/syslog_client/ca.crt		assets/certs/syslog_client/.
	bosh -d $(DEPLOYMENT_NAME) scp $(METIRCSFORWARDER_VM):/var/vcap/jobs/metricsforwarder/config/certs/syslog_client/client.crt assets/certs/syslog_client/.
	bosh -d $(DEPLOYMENT_NAME) scp $(METIRCSFORWARDER_VM):/var/vcap/jobs/metricsforwarder/config/certs/syslog_client/client.key assets/certs/syslog_client/.

	@echo "Build metricsforwarder config yaml"
	cp assets/metricsforwarder.yml metricsforwarder.yml

	# remove SERVER TLS CONFIG so that it starts a http server
	sed -i'' -e 's|\/var\/vcap\/jobs\/metricsforwarder\/config|\/home\/vcap\/app/assets|g' metricsforwarder.yml
	sed -i'' -e 's|$(DEPLOYMENT_NAME).autoscalerpostgres.service.cf.internal|$(POSTGRES_IP)|g' metricsforwarder.yml
