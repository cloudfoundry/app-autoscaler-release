# App Runtime Interfaces - Concourse on GCP Kubernetes
---
# Background
Based on [cloudfoundry/bosh-community-stemcell-ci-infra](https://github.com/cloudfoundry/bosh-community-stemcell-ci-infra)
## Requirements

### Required tools
We use asdf with the version stored in `.tools-versions` file
* glcoud
* helm
* terraform
* terragrunt
* kapp
* ytt
* vendir
* yq

The required tools are defined in the .tool-versions file
```
# add required plugin as defined in asdf .tool-versions
for p in $(cat .tool-versions | awk '{ print $1 }'); do asdf plugin add $p &&  asdf install $p; done

# verify the install tools
asdf current
```

### Permissions

Users who are required to perform operations need to be added in the Role `WG CI Manage` via IAM in the Google Cloud console.

# Prerequisites for a fresh project
## 1. Configuration
Adjust `config.yml`

For your project:
* Copy terragrunt code from `terragrunt/concourse-wg-ci'
* Use git resource for terraform modules (or copy `terraform-modules` folder)
## 2. Logon to your GCP account
```
gcloud auth login && gcloud auth application-default login
```

## 3. Create Github OAuth App and supply as a Google Secret
This is necessary if you want to be able to authenticate with your GitHub profile.
 1. Create Github OAuth App

    Log on to github.com https://github.com/settings/developers -> Click "New OAuth App"

    As "Homepage URL", enter the Concourse's base URL beginning with **https://**.

    As "Authorization callback URL", enter the Concourse URL followed by `/sky/issuer/`callback` also beginning with **https**://**.


 2. Create Google Secret - use a correct naming convention
The created secret name will be used by terraform scripts and needs to conform to the following convention `${gke_name}-concourse-github-oauth`. By default, `gke_name` is `wg-ci` (working group CI system).

    ```sh
    cd <folder with config.yaml>
    ```

    ```sh
    secret_id="$(yq .gke_name config.yaml)-concourse-github-oauth"
    secret_region="$(yq .region config.yaml)"
    project="$(yq .project config.yaml)"
    gcloud secrets create ${secret_id} \
     --replication-policy="user-managed" \
     --locations=${secret_region}\
     --project=${project}
    ```


3. Please create a version for google secret using gcloud command or webui, with a following key-value format

    ```yaml
    id: your Client ID
    secret: your Client secret
    ```
   *gcloud* cli example
    ```sh
    cd <folder with config.yaml>
    ```
    ```sh
    secret_id="$(yq .gke_name config.yaml)-concourse-github-oauth"
    project="$(yq .project config.yaml)"
    id="your Client ID"
    secret="your Client secret"
    echo -n "id: ${id}\nsecret: ${secret}" | \
    gcloud secrets versions add ${secret_id} --data-file=- --project=${project}
    ```

    For more information please refer to [gcloud documentation](https://cloud.google.com/secret-manager/docs/creating-and-accessing-secrets).
## 4. Apply terrgrunt for the entire stack
The following command needs to be run from within your root directory (containing `config.yaml` file).

*NOTE: it's not possible to `plan` for a fresh project due to the fact we can't test kubernetes resources against non-existing cluster*
```sh
terragrunt run-all apply
```

---
# Recommendations
## Cloud SQL Instance deletion protection
Terraform hashicorp provider includes a deletion protection flag however in some cases it's misleading as it's not setting it on Google Cloud.
To avoid confusion we do not set it in the code and recommend altering your production SQL Instance to protect from the deletion on the cloud side.

https://console.cloud.google.com/sql/instances/ -> select instance name -> edit ->  Data Protection -> tick: Enable delete protection

# Developer notes
Please see [developer notes](doc/developer_notes.md) about `vendir sync` and developing modules with `terragrunt`.

# Notes and known limitations

---
## Carvel kapp terraform provider not available for Apple M1
https://github.com/vmware-tanzu/terraform-provider-carvel/issues/30#issuecomment-1311465417

## Destroy the project
Since we protect a backup of credhub encryption key (stored in GCP Secret Manager) to fully destroy the project it needs to be removed from terraform state first.

```
cd <folder with config.yaml>/dr_create

terragrunt state rm google_secret_manager_secret_version.credhub_encryption_key
terragrunt state rm google_secret_manager_secret.credhub_encryption_key
```

**WARNING: to complete deletion, remove the secret from GCP Secret manager -- please be aware doing so will _permantently_ prevent DR recovery**

```
gcloud secrets delete <gke_name>-credhub-encryption-key --project=<your project name>
```

To destroy:
```
terragrunt run-all destroy
```

Delete terraform state gcp bucket from GCP console or via `gsutil`

## Plan/apply terragrunt for a specific component of the stack

```sh
cd concourse/app
terragrunt plan
terragrunt apply
```



## How to obtain GKE credentials for your terminal
Terraform code is fetching GKE credentials automatically. In case you need to access the cluster with `kubectl` (or other kube clients) or to connect to Credhub instance (via `scripts/start-credhub-cli.sh`)

```sh
gcloud container clusters list
# Example output:
# NAME   LOCATION        MASTER_VERSION   MASTER_IP     MACHINE_TYPE   NODE_VERSION     NUM_NODES  STATUS
# wg-ci  europe-west3-a  1.23.8-gke.1900  34.159.31.85  e2-standard-4  1.23.8-gke.1900  3          RUNNING

gcloud container clusters get-credentials wg-ci --zone europe-west3-a
# Example output:
# Fetching cluster endpoint and auth data.
# kubeconfig entry generated for wg-ci.

kubectl config current-context
# Example output:
# gke_app-runtime-interfaces-wg_europe-west3-a_wg-ci
```

## DR scenario
Please see [DR scenario](doc/disaster_recovery.md) for fully automated recovery procedure.

### DR credhub encryption check

The dr_create module will check for the existence and integrity of the Credhub encryption key. Following errors may appear if the user does not execute dr-create
1. Crehub encryption key does not exist in google secret manager or has no version
   ```
   │ Error: Error retrieving available secret manager secret versions: googleapi: Error 404: Secret [projects/899763165748/secrets/wg-ci-test-credhub-encryption-key] not found or has no versions.
   │
   │   with data.google_secret_manager_secret_version.credhub_encryption_key,
   │   on credhub_dr_check.tf line 2, in data "google_secret_manager_secret_version" "credhub_encryption_key":
   │    2: data "google_secret_manager_secret_version" "credhub_encryption_key" {
   │
   ```
2. Credhub encryption keys stored in google secrets manager is different to the one stored in kubernetes secret 
    ```
    │ Error: Call to unknown function
    │ 
    │   on .terraform/modules/assertion_encryption_key_identical/.tf line 6, in locals:
    │    6:   content = var.condition ? "" : SEE_ABOVE_ERROR_MESSAGE(true ? null : "ERROR: ${var.error_message}")
    │     ├────────────────
    │     │ var.error_message is "*** Encryption keys in GCP Secret Manager and kubernetes secrets do not match ***"
    │ 
    │ There is no function named "SEE_ABOVE_ERROR_MESSAGE".
    ```

## Secrets rotation
* Currently not implemented
