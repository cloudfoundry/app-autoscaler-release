runtime-og-ci
=============

This repository provides all public scripts and pipeline deployments used
by the runtime_og team. To reproduce this pipeline, you can use your own
private configuration files for the `pipeline.yml` files as described below.

## autoscaler
-------------

This directory contains the concourse `pipeline.yml` for the autoscaler [pipeline](https://runtime-og.ci.cf-app.com/pipelines/autoscaler)
and all of the associated scripts. Tp use this manifest, you need to provide a private configuration file
for all of the template parameters.

NOTE: If you are recreating this pipeline, for personal use and do not have authority to update
tracker or push to github. The `pipeline.yml` file needs to have any `tracker` sections commented
out as well as the app-autoscaler private key

## runtime-og
-------------

This folder contains the concourse `pipeline.yml` for runtime-og [pipeline](https://runtime-og.ci.cf-app.com/pipelines/runtime-og)
and all of the associated scripts. To use this manifest one needs to provide a private configuration file
for all of the template credentials.

NOTE: if this pipeline is being recreated with a purpose of only running tests then you need to ensure
that all of the tracker and auto-bump task have been removed or commented out. As well as their associated
resources.

## ci-images
------------

This docker image is shared between the runtime-og and autoscaler pipeline. If a new image needs to be made,
please tag the image with the git commit. This way we can correlate which docker file built which image.
Also update the latest tag with the newest image. This way you can docker pull and always grab the most recent build.
