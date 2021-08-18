#!/bin/bash

set -euo pipefail

export PREVIOUS_VERSION="$(cat gh-release/tag)"

mkdir -p generated-release

pushd app-autoscaler-release
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


  pushd src/changelog
    RECOMMENDED_VERSION_FILE=../../../generated-release/name OUTPUT_FILE=../../../generated-release/changelog.md go run main.go
  popd

  VERSION=$(cat ../generated-release/name)

  # create bosh release with the specified version
  bosh create-release \
    --final \
    --version "$VERSION" \
    --tarball=app-autoscaler-v$VERSION.tgz
  
  RELEASE_TGZ=app-autoscaler-v$VERSION.tgz
  export SHA1=$(sha1sum $RELEASE_TGZ | head -n1 | awk '{print $1}')
  echo "SHA1=$SHA1"

  echo "${VERSION}" > ../generated-release/tag

  mkdir -p ../generated-release/artifacts
  mv app-autoscaler-v${VERSION}.tgz ../generated-release/artifacts/

  cat >> ../generated-release/changelog.md <<EOF

## Deployment

\`\`\`yaml
releases:
- name: app-autoscaler
  version: $VERSION
  url: https://storage.googleapis.com/app-autoscaler-releases/releases/app-autoscaler-v${VERSION}.tgz
  sha1: $SHA1
\`\`\`
EOF

  cat ../generated-release/changelog.md
  
  git status
popd
