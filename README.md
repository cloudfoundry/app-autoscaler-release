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
	--v1
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

** please note `./script/generate-bosh-lite-manifest` uses GNU `getopt` to parse command options. GNU `getopt` is installed by default on Linux, but on Mac OS X and FreeBSD it needs to be installed separately. On Mac OS, use `brew install gnu-getopt`, or install [MacPorts](http://www.macports.org) and then do `sudo port install getopt` to install GNU getopt (usually into `/opt/local/bin`), and make sure that `/opt/local/bin` is in your shell path ahead of `/usr/bin`. On FreeBSD, install `misc/getopt`.

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
cf create-service-broker autoscaler username password https://servicebroker.service.cf.internal:6101
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