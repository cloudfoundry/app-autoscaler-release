#! /bin/bash
for release in $(yq ".releases[].url" templates/app-autoscaler-deployment.yml);
do
  bosh upload-release --fix "${release}"
done

