# Versioning with vendir
```
cd ./files
```
Update `vendir.yml`
```
vendir sync
```
Commit changes to the git repo.

# Note on UAA
File `files/config/uaa/_ytt_lib/uaa/k8s/templates/deployment.star` has been altered manually and removes ` "-DSECRETS_DIR={}".format(secrets_dir),` line from the original template.

This parameter when present will prevent uaa pod to populate `UAA_POSTGRES_HOST` env variable