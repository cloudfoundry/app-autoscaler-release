#! /usr/bin/env bash

set -euo pipefail

PROJECT="app-runtime-interfaces-wg"
ZONE="europe-west3-a"
#VM_LIST=("vm-4daadccc-3c5b-47c1-6d2c-dcadb15c93f2" "vm-5461ec15-36b0-4d36-667b-8fb9223fc405" "vm-5886f3a0-5e49-479b-67d4-61cbdb3b402b" "vm-5c753d79-08c7-4dee-6315-d408a03bec11",
#"vm-b7bcb8b2-f200-4686-6f24-96acb28c4125" "vm-7a3a08c5-557b-4602-68d1-9d8ffb944783" "vm-26e43861-2e13-47c2-472e-4ffc8f8f4fc6")

# autoscaler-performance deployment
VM_LIST=("vm-7d74e45e-7d2e-44ca-5663-06fd6b67c350" "vm-5886f3a0-5e49-479b-67d4-61cbdb3b402b" "vm-5c753d79-08c7-4dee-6315-d408a03bec11"
"vm-b7bcb8b2-f200-4686-6f24-96acb28c4125" "vm-7a3a08c5-557b-4602-68d1-9d8ffb944783" "vm-26e43861-2e13-47c2-472e-4ffc8f8f4fc6" "vm-92037a76-4882-4818-62cf-5855df61bbb6")

for VM in "${VM_LIST[@]}"
do
  echo "$(date): Starting $VM"
  gcloud compute instances resume "$VM" --zone="$ZONE" --project="$PROJECT"
done
