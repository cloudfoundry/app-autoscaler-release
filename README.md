# Bosh Release for app-autoscaler service
(This release is under active development)

## Purpose

The purpose of this bosh release is to deploy and setup the [app-autoscaler](https://github.com/cloudfoundry-incubator/app-autoscaler) service.

## Usage

### Bosh Lite Deployment 

#### Deploy on bosh-lite with cf-release
Install and start [BOSH-Lite](https://github.com/cloudfoundry/bosh-lite), following its [README](https://github.com/cloudfoundry/bosh-lite/blob/master/README.md).
Modify the cloud-config and deployment manifest settings by modifying the files under /example directory.
Install [Spiff](https://github.com/cloudfoundry-incubator/spiff#installation)

Instructions to install on [bosh-lite](https://github.com/cloudfoundry/bosh-lite) below:

```sh
bosh target BOSH_DIRECTOR_HOST
git clone https://github.com/cloudfoundry-incubator/app-autoscaler-release
cd app-autoscaler-release
./scripts/update
```

> Deploy using BOSH v2 manifest

```sh
bosh update cloud-config ./example/cloud-config.yml
./scripts/generate-bosh-lite-manifest \
	-c <path to cf-release deployment manifest> \
	-p ./example/property-overrides.yml
./scripts/deploy
```

> Deploy using BOSH v1 manifest

```sh
./scripts/generate-bosh-lite-manifest \
	-c <path to cf-release deployment manifest> \
	-p ./example/property-overrides.yml \
	-v v1
./scripts/deploy
```

> ** cf-release deployment manifest should be cf-release/bosh-lite/deployments/cf.yml
>
> To use pre-existing postgres server(s), it is required to pass db-stubs while generating manifest. Otherwise an instance of default postgres server will be provided as part of app-autoscaler deployment.

```sh
./scripts/generate-bosh-lite-manifest \
	-c <path to cf-release deployment manifest> \
	-p ./example/property-overrides.yml \
	-d ./example/dbstubs/db-stub-external.yml \
```

#### Deploy on bosh-lite with cf-deployment
Install [Bosh-cli-v2](https://bosh.io/docs/cli-v2.html#install)

Install and start [BOSH-Deployment](https://github.com/cloudfoundry/bosh-deployment), following its [README](https://github.com/cloudfoundry/bosh-deployment/blob/master/README.md). 

Install [CF-deployment](https://github.com/cloudfoundry/cf-deployment/blob/master/cf-deployment.yml)

Create and upload release
```sh
git clone https://github.com/cloudfoundry-incubator/app-autoscaler-release
cd app-autoscaler-release
./scripts/update
bosh create-release
bosh -e YOUR_ENV upload-release
```
Deploy app-autoscaler
```sh
bosh -e YOUR_ENV -d app-autoscaler \
     deploy templates/app-autoscaler-deployment.yml \
     --vars-store=bosh-lite/deployments/vars/autoscaler-deployment-vars.yml \
     -v system_domain=bosh-lite.com \
     -v cf_admin_password=<cf admin password> \
     -v skip_ssl_validation=true
```

Alternatively you can use cf-deployment vars file to provide the cf_admin_password
```sh
bosh -e YOUR_ENV -d app-autoscaler \
     deploy templates/app-autoscaler-deployment.yml \
     --vars-store=bosh-lite/deployments/vars/autoscaler-deployment-vars.yml \
     -v system_domain=bosh-lite.com \
     -v skip_ssl_validation=true \
     --vars-file=<path to cf deployment vars file>
```
>** It's advised not to make skip_ssl_validation=true for non-development environment


## Register service 

Log in to Cloud Foundry with admin user, and use the following commands to register `app-autoscaler` service

```
cf create-service-broker autoscaler <brokerUserName> <brokerPassword> <brokerURL>
```

* `brokerUserName`: the user name to authenticate with service broker
* `brokerPassword`: the password to authenticate with service broker
* `borkerURL`: the URL of the service broker

All these parameters are configured in the bosh deployment. If you are using default values of deployment manifest, register the service with the commands below.

```
cf create-service-broker autoscaler username password https://autoscalerservicebroker.bosh-lite.com
```

## Acceptance test

Refer to [AutoScaler UAT guide](src/acceptance/README.md) to run acceptance test. 

## Use service

To use the service to auto-scale your applications, log in to Cloud Foundry with admin user, and use the following command to enable service access to all or specific orgs.
```
cf enable-service-access autoscaler [-o ORG]
```
The following commands don't require admin rights, but user needs to be Space Developer. Create the service instance, and then bind your application to the service instance with the policy as parameter.

```
cf create-service autoscaler  autoscaler-free-plan  <service_instance_name>
cf bind-service <app_name> <service_instance_name> -c <policy>
```

## Remove the service

Log in to Cloud Foundry with admin user, and use the following commands to remove all the service instances and the service broker of `app-autoscaler` from Cloud Foundry.

```
cf purge-service-offering autoscaler
cf delete-service-broker autoscaler
```
