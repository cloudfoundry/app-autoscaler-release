#!/usr/bin/env bash
set -euo pipefail

     tg_infra_params=( --terragrunt-exclude-dir ./app --terragrunt-exclude-dir ./backend --terragrunt-exclude-dir ./dr_create --terragrunt-source-update  )
   tg_backend_params=( --terragrunt-exclude-dir ./app --terragrunt-exclude-dir ./infra --terragrunt-exclude-dir ./dr_create --terragrunt-source-update  )
tg_dr_restore_params=( --terragrunt-config=credhub_sql_passwords.hcl --terragrunt-source-update )
       tg_app_params=( --terragrunt-exclude-dir ./infra --terragrunt-exclude-dir ./backend --terragrunt-exclude-dir ./dr_create --terragrunt-source-update  )


if [ ! -s "./config.yaml" ]; then
    echo "ERROR: Please 'cd' to your folder with config.yaml and run ./dr/dr_restore.sh"
    exit 1
fi

echo ">> Plan terragrunt for infra only"; echo
terragrunt run-all plan "${tg_infra_params[@]}"
echo

echo ">> Apply terragrunt for infra only"; echo
terragrunt run-all apply "${tg_infra_params[@]}"
echo
# -------------------------------------------------------------------

echo ">> Plan terragrunt for backend only [1/3]"; echo
terragrunt run-all plan "${tg_backend_params[@]}"
echo

echo ">> Apply terragrunt for backend only [1/3]"; echo
terragrunt run-all apply "${tg_backend_params[@]}"
echo

# -------------------------------------------------------------------
# Carvel might not learn new state during the recovery an we need to retrigger it
echo ">> Plan terragrunt for backend only [2/3] to trigger carvel kapp provider [1/2] "; echo
terragrunt run-all plan "${tg_backend_params[@]}"
echo

echo ">> Apply terragrunt for backend only [2/3] to trigger carvel kapp provider [1/2]"; echo
terragrunt run-all apply "${tg_backend_params[@]}"
echo

echo ">> Plan terragrunt for backend only [3/3] to trigger carvel kapp provider [2/2]"; echo
terragrunt run-all plan "${tg_backend_params[@]}"
echo

echo ">> Apply terragrunt for backend only [3/3] to trigger carvel kapp provider [2/2]"; echo
terragrunt run-all apply "${tg_backend_params[@]}"
echo

# -------------------------------------------------------------------

# Apply for terragrunt without run-all will show resources to create before applying, hence plan step is redudant
echo ">> Apply terragrunt (with resource show) to restore credhub encryption key key and populate new sql users passwords from secretgen"
echo -e "\033[0;35m>> Note: use 'yes' to accept\033[0m"
( cd ./dr_restore && terragrunt apply "${tg_dr_restore_params[@]}" )
echo

# -------------------------------------------------------------------
echo ">> >> Plan terragrunt for app only"
terragrunt run-all plan "${tg_app_params[@]}"
echo

echo ">> >> Apply terragrunt for app only"
terragrunt run-all apply "${tg_app_params[@]}"
echo

echo "-- DR recovery completed"