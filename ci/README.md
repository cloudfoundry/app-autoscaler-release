app-autoscaler-ci
=============

This repository provides all public scripts and pipeline deployments used
by the app autoscaler team.  The public pipeline is hosted at: https://bosh.ci.cloudfoundry.org/.

To reproduce this pipeline, you can use your own private configuration files for the `pipeline.yml` files as described below.

## autoscaler
-------------

This directory contains the concourse `pipeline.yml` for the autoscaler [pipeline](https://bosh.ci.cloudfoundry.org/pipelines/app-autoscaler)
and all of the associated scripts. To use this manifest, you need to provide a private configuration file
for all of the template parameters.

NOTE: If you are recreating this pipeline, for personal use and do not have authority to update
tracker or push to github. The `pipeline.yml` file needs to have any `tracker` sections commented
out as well as the app-autoscaler private key

## dockerfiles
------------

These docker images in this repo are built and pushed with GitHub actions, they are hosted on ghcr.io

## Deploy pipeline

__Setup__

```
fly --target autoscaler login --team-name app-autoscaler --concourse-url https://bosh.ci.cloudfoundry.org/
push autoscaler
./set-pipeline.sh
popd
```
