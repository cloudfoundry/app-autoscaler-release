#!/bin/bash

set -euo pipefail

mkdir -p generated-release

pushd app-autoscaler-release
  # determine what the next release version should be
  VERSION=$(jx-release-version)
  echo "Next Version should be ${VERSION}"

  # generate the private.yml file with the credentials
  cat > config/private.yml <<EOF
---
blobstore:
  options:
    credentials_source: static
    json_key:
EOF

  yq eval -i '.blobstore.options.json_key = strenv(UPLOADER_KEY)' config/private.yml

  git status

  # create bosh release with the specified version
  bosh create-release \
    --final \
    --version "$VERSION" \
    --tarball=app-autoscaler-v$VERSION.tgz
  
  RELEASE_TGZ=app-autoscaler-v$VERSION.tgz
  export SHA1=$(sha1sum $RELEASE_TGZ | head -n1 | awk '{print $1}')
  echo "SHA1=$SHA1"

  export RELEASE_ROOT=../generated-release

  echo "v${VERSION}" > ${RELEASE_ROOT}/tag
  echo "v${VERSION}" > ${RELEASE_ROOT}/name
  mkdir -p ${RELEASE_ROOT}/artifacts
  mv app-autoscaler-v${VERSION}.tgz ${RELEASE_ROOT}/artifacts

  #mv ${REPO_ROOT}/ci/release_notes.md          ${RELEASE_ROOT}/notes.md
  #cat >> ${RELEASE_ROOT}/notes.md <<EOF
  
popd
