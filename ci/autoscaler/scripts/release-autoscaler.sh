#!/bin/bash

set -euo pipefail


pushd app-autoscaler-release
  # determine what the next release version should be
  NEXT_VERSION=$(jx-release-version)
  echo "Next Version should be ${NEXT_VERSION}"

  # generate the private.yml file with the credentials
  cat > config/private.yml <<EOF
---
blobstore:
  options:
    credentials_source: static
    json_key:
EOF

  yq eval -n '.blobstore.options.json_key = strenv(UPLOADER_KEY)' config/private.yml

  # REMOVE ME
  cat config/private.yml
  git status

  # create bosh release with the specified version
  bosh create-release --final --version "$NEXT_VERSION" --tarball=releases/app-autoscaler-v"$NEXT_VERSION".tgz
popd

# create the GitHub release (from the correct sha & branch)

# upload release notes

# fail whilst this is a work in progress
exit 1
