# App-autoscaler team and pipelines management for Concourse
Concourse-URL: <https://concourse.app-runtime-interfaces.ci.cloudfoundry.org>

## Dependencies
None. Terraform scripts are contained with terragrunt config.

## Requirements
Make sure that the concourse-target `app-autoscaler-release` for <https://concourse.app-runtime-interfaces.ci.cloudfoundry.org> is known to the fly-cli. You may call `concourse_login 'app-autoscaler-release'` which is defined in <../../../autoscaler/scripts/common.sh>.

Login to your fly target prior to executing terragrunt:
```shell
fly login --target='app-autoscaler-release'
```

## Usage

```sh
terragrunt plan
terragrunt apply
```
