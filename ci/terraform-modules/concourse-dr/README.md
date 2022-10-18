# DR notes

Scenario assumes terraform was run after 1st deployment from `concourse-dr/create` folder. 

## DR scenario tested

* deleted entire deployment including 'concourse' namespace
* deleted all databases and database users
* GKE cluster not destroyed
TODO:
* test recovery with fresh k8s cluster

### DR
#### Prerequisites
Secrets were backed up with terraform in `./concourse-dr/create`

#### Steps
1. Restore SQL Instance from backup
* https://console.cloud.google.com/sql/instances/concourse/backups -> Restore desired backup version

2. Ensure infra part is up to scratch. Should report a namespace will be created.
Please note the usage of brackets as these allow you to execute bash commands from subfolders and return to the current folder once finished.
    ```
    ( cd ./concourse-infra && terragrunt apply )
    ```
    *Note: terraform reporting missing databases at this point is an indication instance restoration needs to be run.*

2. Recreate 'backend' part.
    ```
    ( cd ./concourse-backend && terragrunt apply )
    ```

    * **MAINTENANCE:** In case carvel kapp is unwilling to apply you can taint and re-provision. 
      _WARNING_ proceed with caution if you use the backend in other projects on the cluster (ie. carvel secret gen). Shall this be a case secretgen should not be a part of managed concourse deployment anymore.
      ```
      cd ./concourse-backend
      terraform taint carvel_kapp.concourse_backend
      terraform plan
      terraform apply
      ```

4. Restore secrets
   ```
   ( cd concourse-dr/restore && terraform apply )
   ```
   *NOTE:* due to the dr recovery workflow need this part is not managed with terragrunt.

5. Deploy remaining components
From this moment onward `terragrunt` should be happy to run again as usual.
    ```
    terragrunt run-all plan
    terragrunt run-all apply
    ```
  

