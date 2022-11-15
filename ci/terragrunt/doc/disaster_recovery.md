# Prerequisites

1. The backup of credhub encryption key has been stored in GCP Secret Manager - this part is handled automatically with `dr_creste` terragrunt part of the stack
2. The secret in GCP was not deleted/altered manually.
3. Credhub database exists or is available or recovered from a backup.


## DR scenario tested

* deleted entire deployment including 'concourse' namespace
* deleted all databases and database users with db recovered from backup
* GKE cluster destroyed



# Steps
Fully automated restore with:
```
cd <folder witg config.yaml>
../scripts/dr_restore.sh
```


---
# Troubleshooting
### Unexpected credhub encryption-key-in k8s secrets
Providing GKE cluster or application was removed recovery is not expecting credhub-encryption-key stored in kubernetes secrets. Please remove it from k8s since it will be restored from GCP Secret Manager.

```
╷
│ Error: secrets "credhub-encryption-key" already exists
│ 
│   with kubernetes_secret_v1.credhub_encryption_key,
│   on credhub_restore.tf line 6, in resource "kubernetes_secret_v1" "credhub_encryption_key":
│    6: resource "kubernetes_secret_v1" "credhub_encryption_key" {
│ 
╵
```

###  Carvel kapp is unwilling to apply backend changes

   In case carvel kapp is unwilling to apply backend changes you can taint it and re-provision.
  _WARNING_ proceed with caution if you use the backend in other projects on the cluster (ie. carvel secret gen). Shall this be a case secretgen should not be a part of managed concourse deployment anymore.

```
cd ./backend
terragrunt taint carvel_kapp.concourse_backend
terragrunt plan
terragrunt apply
```
Re run dr restore
```
cd ..
../scripts/dr_restore.sh
```