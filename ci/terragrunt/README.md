
## Requirements

### Required tools
* glcoud
* helm
* terraform
* terragrunt
* kapp
* ytt
* vendir

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

## 3. Create Github OAuth token and supply as a Google Secret
 1. Request creation of a Google Secret
    ```sh
      terragrunt run-all apply --target module.concourse-infra.google_secret_manager_secret.github_oauth
    ```

 2. Create Github OAuth token

This is necessary if you want to be able to authenticate with your GitHub profile. Log on to github.com and navigate to:
"Settings" -> "Developer settings" -> "OAuth Apps" -> "New OAuth App"

As "Homepage URL", enter the Concourse's base URL. As "Authorization callback URL", enter the Concourse URL followed
by `/sky/issuer/callback`.

3. Please create a version for google secret using gcloud command or webui, with a following key-value format

```yaml
id: paste your Client ID
secret: paste your Client secret
```


## 4. Apply terrgrunt for the entire stack
The following commands need to be run from within this directory “terragrunt/concourse”:
```sh
terragrunt run-all apply
```
*NOTE: it's not possible to `plan` for a fresh project due to the fact we can't test kubernetes resources against non existing cluster*

## 5. Save secrets needed for DR scenario
This part is not intended to be fully automated.
```sh
cd ../concourse-dr
terragrunt plan --terragrunt-config=create.hcl
terragrunt apply --terragrunt-config=restore.hcl
```


---
## Upgrade components managed by kapp and vendir (when needed)
Required actions:
* changing charts versions
* `vendir sync`
* please see readme in terraform-modules/backend

## Plan/apply terragrunt for changes to modules

```sh
terragrunt run-all plan --terragrunt-source-update
```

## Destroy the project
```
terragrunt run-all destroy
```

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
