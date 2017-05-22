# Bosh Release for app-autoscaler service
(This release is under active development)

## Purpose

The purpose of this bosh release is to deploy and setup the [app-autoscaler](https://github.com/cloudfoundry-incubator/app-autoscaler) service.

## Usage

### Bosh Lite Deployment 
Install and start [BOSH-Lite](https://github.com/cloudfoundry/bosh-lite), following its   [README](https://github.com/cloudfoundry/bosh-lite/blob/master/README.md).
Modify the cloud-config and deployment manifest settings by modifying the files under /example directory.
Install [Spiff](https://github.com/cloudfoundry-incubator/spiff#installation)

Instructions to install on [bosh-lite](https://github.com/cloudfoundry/bosh-lite) below:

```sh
bosh target BOSH_DIRECTOR_HOST
bosh update cloud-config <PATH_TO_CLOUD_CONFIG>
git clone https://github.com/cloudfoundry-incubator/app-autoscaler-release
cd app-autoscaler-release
./scripts/update
./scripts/generate-bosh-lite-manifest \
	-c <path to cf-release deployment manifest> \
	-p ./example/property-overrides.yml
./scripts/deploy
```


> ** cf-release deployment manifest should be cf-release/bosh-lite/deployments/cf.yml
>
> ** To generate BOSH V1 manifest template use --v1 flag with generate-bosh-lite-manifest. By default BOSH V2 manifest will be generated.
>
> To use pre-existing postgres server(s), it is required to pass db-stubs while generating manifest. Otherwise an instance of default postgres server will be provided as part of app-autoscaler deployment.

```sh
./scripts/generate-bosh-lite-manifest \
	-c <path to cf-release deployment manifest> \
	-p ./example/property-overrides.yml \
	-d ./example/dbstubs/db-stub-external.yml \
```

## Register service 

Log in Cloud Foundry with admin user, and use the following commands to register `app-autoscaler` service

```
cf create-service-broker autoscaler <brokerUserName> <brokerPassword> <brokerURL>
cf enable-service-access autoscaler
```

* `brokerUserName`: the user name to authenticate with service broker
* `brokerPassword`: the password to authenticate with service broker
* `borkerURL`: the URL of the service broker

All these parameters are configured in the bosh deployment. If you are using default values of deployment manifest, register the service with the commands below.

```
cf create-service-broker autoscaler username password https://servicebroker.service.cf.internal:6101
cf enable-service-access autoscaler

```

## Acceptance test

Refer to [AutoScaler UAT guide](src/acceptance/README.md) to run acceptance test. 

## Use service

To use the service to auto-scale your application, firstly create the service, and then bind to you application with policy as parameter. 

```
cf create-service autoscaler  autoscaler-free-plan  <service_instance_name>
cf bind-service <app_name> <service_instance_name> -c <policy>
```

## Remove the service

Log in Cloud Foundry with admin user, and use the following commands to remove all the service instances and the service broker of `app-autoscaler` from Cloud Foundry.

```
cf purge-service-offering autoscaler
cf delete-service-broker autoscaler
```