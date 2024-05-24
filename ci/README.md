# app-autoscaler-ci

This repository provides all public scripts and pipeline deployments used

By the app autoscaler team.  The public pipeline is hosted at: <https://concourse.app-runtime-interfaces.ci.cloudfoundry.org>.

To reproduce this pipeline, you can use your own private configuration files for the `pipeline.yml` files as described below.

## Autoscaler

This directory contains the concourse `pipeline.yml` for the autoscaler [pipeline](https://concourse.app-runtime-interfaces.ci.cloudfoundry.org/teams/app-autoscaler/pipelines/app-autoscaler-release)
and all of the associated scripts. To use this manifest, you need to provide a private configuration file

for all of the template parameters.

NOTE: If you are recreating this pipeline, for personal use and do not have authority to update
tracker or push to github. The `pipeline.yml` file needs to have any `tracker` sections commented
out as well as the app-autoscaler private key

## dockerfiles

These docker images in this repo are built and pushed with GitHub actions, they are hosted on ghcr.io

## Terrgrunt

This directory contains the terragrunt managed stacks of resouces in account app-runtime-interfaces-wg GCP project.

## Deploy pipeline

__Setup__

```
fly --target app-autoscaler-release login --team-name app-autoscaler --concourse-url https://concourse.app-runtime-interfaces.ci.cloudfoundry.org
push autoscaler
./set-pipeline.sh
popd
```

## Prometheus

This is deployed using the script [deploy-prometheus](infrastructure/scripts/deploy-prometheus.sh).
To deploy localy you will need:

- bosh ca certificate and place this it is `${HOME}/.ssh/bosh.ca.crt`.
- <https://github.com/bosh-prometheus/prometheus-boshrelease> cloned in ../
- <https://github.com/cloudfoundry/app-autoscaler-env-bbl-state> cloned in ../

Then you can run the script directly.

### setup

- The Slack channel is stored in the cf credhub under `/bosh-autoscaler/prometheus/alertmanager_slack_channel`
- The Slack Message can be customised in the [slack-receiver-template.yml](operations/slack-receiver-template.yml)
