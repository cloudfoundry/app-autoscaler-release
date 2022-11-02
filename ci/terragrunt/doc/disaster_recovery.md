# Prerequisites

Scenario assumes terragrun was run after 1st deployment from `concourse-dr` folder.
```sh
cd concourse-dr
terragrunt plan --terragrunt-config=create.hcl
terragrunt apply --terragrunt-config=restore.hcl
```


## DR scenario tested

* deleted entire deployment including 'concourse' namespace
* deleted all databases and database users
* GKE cluster not destroyed

TODO:
* test recovery will all infrastructure destroyed but sql instance


# Steps
## 1. Restore SQL Instance from backup
* https://console.cloud.google.com/sql/instances/
  * Choose the desired database instance 
  * Restore desired backup version

## 2. Ensure infra and backend parts are up to date 
Please note the usage of brackets as these allow you to execute bash commands from subfolders and return to the current folder once finished.

```
( cd ./concourse && terragrunt run-all plan --terragrunt-exclude-dir ./app )
```
*Note: terragrunt plan only works if kubernetes cluster exists*

```
( cd ./concourse && terragrunt run-all apply --terragrunt-exclude-dir ./app )
```
*Note: terraform reporting missing databases at this point is an indication instance restoration needs to be run.*


## 3. Restore secrets
```
( cd concourse-dr && terragrunt apply --terragrunt-config=restore.hcl )
```

## 4. Deploy remaining components
From this moment onward `terragrunt` should be happy to run again as usual.
```
terragrunt run-all plan
terragrunt run-all apply
```

# Troubleshooting

###  Carvel kapp is unwilling to apply backend changes

   In case carvel kapp is unwilling to apply backend changes you can taint it and re-provision.
  _WARNING_ proceed with caution if you use the backend in other projects on the cluster (ie. carvel secret gen). Shall this be a case secretgen should not be a part of managed concourse deployment anymore.

```
cd ./concourse/backend
terragrunt taint carvel_kapp.concourse_backend
terragrunt plan
terragrunt apply
```

