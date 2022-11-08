
## Requirements

### Required tools
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

Users who are required to perform operations need to be added in the Role `WG CI Manage` via IAM in Google Cloud console.

# Prerequisites for a fresh project
## 1. Configuration
Adjust `config.yml`

## 2. Logon to your GCP account
```
gcloud auth login && gcloud auth application-default login
```

## 3. Create Github OAuth App and supply as a Google Secret
This is necessary if you want to be able to authenticate with your GitHub profile.
 1. Create Github OAuth App 

    Log on to github.com https://github.com/settings/developers -> Click "New OAuth App"

    As "Homepage URL", enter the Concourse's base URL beginning with **https://**. 
 
    As "Authorization callback URL", enter the Concourse URL followed by `/sky/issuer/callback` also beginnign with **https://**.


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
The following command needs to be run from within your root `concourse` directory (containing `config.yaml` file).

*NOTE: it's not possible to `plan` for a fresh project due to the fact we can't test kubernetes resources against non existing cluster*
```sh
terragrunt run-all apply
```


## 5. Save secrets needed for DR scenario
This part is not intended to be fully automated.
```sh
cd ../concourse/dr
terragrunt plan --terragrunt-config=create.hcl
terragrunt apply --terragrunt-config=create.hcl
```


---

## Destroy the project
```
terragrunt run-all destroy
```

## Plan/apply terragrunt for a specific component of the stack
e
```sh
cd concourse/app
terragrunt plan
terragrunt apply
```

## Plan/apply terragrunt for changes to modules
Update your terragrunt cache folders when terraform source modules code would change
```sh
terragrunt run-all plan --terragrunt-source-update
```

## Upgrade components managed by kapp and vendir (when needed)
Required actions:
* changing charts versions
* `vendir sync`
* please see readme in terraform-modules/backend

## How to obtain GKE credentials for your terminal

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
Please see [DR scenario readme](doc/disaster_recovery.md)

## Secret rotation
* Quark Secrets have been dropped.
* TBD process with Carvel Secret Manager
