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

## Acceptance test

Refer to [AutoScaler UAT guide](src/acceptance/README.md) to run acceptance test. 
>
> To use pre-existing postgres server(s), it is required to pass db-stubs while generating manifest. Otherwise an instance of default postgres server will be provided as part of app-autoscaler deployment.

```sh
./scripts/generate-bosh-lite-manifest \
	-c <path to cf-release deployment manifest> \
	-p ./example/property-overrides.yml \
	-d ./example/dbstubs/db-stub-external.yml \
```

