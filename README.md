# Bosh Release for app-autoscaler service
(This release is under active development)

## Purpose

The purpose of this bosh release is to deploy and setup the [app-autoscaler](https://github.com/cloudfoundry-incubator/app-autoscaler) service.

## Usage

Modify the cloud-config and deployment manifest settings by modifying the files under /example directory.
Install [Spiff](https://github.com/cloudfoundry-incubator/spiff#installation)
Instructions to install on [bosh-lite](https://github.com/cloudfoundry/bosh-lite) below:

```
bosh target BOSH_DIRECTOR_HOST
git clone https://github.com/cloudfoundry-incubator/app-autoscaler-release
cd app-autoscaler-release
./scripts/update
./scripts/sync-package-specs
./scripts/generate-bosh-lite-manifest
./scripts/deploy
```