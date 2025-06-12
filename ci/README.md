# app-autoscaler-ci
This repository provides all public scripts and pipeline deployments used by the app “autoscaler-team”. The public pipeline is hosted at: <https://concourse.app-runtime-interfaces.ci.cloudfoundry.org>

To reproduce this pipeline, you can use your own private configuration files for the `pipeline.yml` files as described below.

🚸 __Important__: Regarding the concourse-pipelines, please note that there is a dedicated credhub-instance that differs from the one that is used for the bosh-director to render manifests. It needs to contain the credhub-secrets that are referenced in the pipeline-definition. Their paths must be prefixed by `/concourse/<team-name>` (e.g. `/concourse/app-autoscaler`). The login to that instance can be done via calling the script [terragrunt/scripts/concourse/start-credhub-cli.sh](<https://github.com/cloudfoundry/app-runtime-interfaces-infrastructure/blob/main/terragrunt/scripts/concourse/start-credhub-cli.sh>) in the repository <https://github.com/cloudfoundry/app-runtime-interfaces-infrastructure>.

## Autoscaler
This directory contains the concourse `pipeline.yml` for the autoscaler-[pipeline](<https://concourse.app-runtime-interfaces.ci.cloudfoundry.org/teams/app-autoscaler/pipelines/app-autoscaler-release>) and all of the associated scripts. To use this manifest, you need to provide a private configuration file for all of the template parameters.

🪧 _NOTE_: If you are recreating this pipeline, for personal use and do not have authority to update tracker or push to github. The `pipeline.yml` file needs to have any `tracker` sections commented out as well as the app-autoscaler private key.

## Dockerfiles
These docker images in this repo are built and pushed with GitHub actions, they are hosted on <ghcr.io>.

## Terragrunt
This directory contains the terragrunt managed stacks of resources in the account app-runtime-interfaces-wg GCP project. It comes with its dedicated instructions in <./terragrunt/app-autoscaler/concourse/README.md> on how to set-up concourse-related resources.

## Deploy pipeline
__Setup__

```shell
make set-target
make set-autoscaler-pipeline
```

## Unpause pipeline and jobs
You will be prompted to select the specific jobs you want to unpause.
```shell
make unpause-pipeline
```

## Prometheus
This is deployed using the script [deploy-prometheus](<./infrastructure/scripts/deploy-prometheus.sh>). To deploy locally you will need:
 + bosh ca certificate and place this it is `${HOME}/.ssh/bosh.ca.crt`.
 + <https://github.com/bosh-prometheus/prometheus-boshrelease> cloned in ../
 + <https://github.com/cloudfoundry/app-autoscaler-env-bbl-state> cloned in ../

Then you can run the script directly.

### Setup
 + The Slack channel is stored in the cf credhub under `/bosh-autoscaler/prometheus/alertmanager_slack_channel`
 + The Slack Message can be customised in the [slack-receiver-template.yml](<./operations/slack-receiver-template.yml>)
