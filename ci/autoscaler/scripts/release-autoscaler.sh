#!/bin/bash

set -euo pipefail

export PREVIOUS_VERSION="$(cat gh-release/tag)"

mkdir -p generated-release
export GENERATED=$(realpath generated-release)

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

  export SUBMODULE_CURRENT_SHA=$(git ls-tree HEAD src/app-autoscaler | awk '{print $3}')
  echo "Autoscaler SHA = $SUBMODULE_CURRENT_SHA"

  pushd src/changelog
    RECOMMENDED_VERSION_FILE=${GENERATED}/name OUTPUT_FILE=${GENERATED}/changelog.md go run main.go
  popd

  VERSION=$(cat ${GENERATED}/name)

  if [ "${PERFORM_BOSH_RELEASE}" == "true" ]; then
    # create bosh release with the specified version
    bosh create-release \
      --final \
      --version "$VERSION" \
      --tarball=app-autoscaler-v$VERSION.tgz
  
    RELEASE_TGZ=app-autoscaler-v$VERSION.tgz
    export SHA1=$(sha1sum $RELEASE_TGZ | head -n1 | awk '{print $1}')
    echo "SHA1=$SHA1"

    mkdir -p ${GENERATED}/artifacts
    mv app-autoscaler-v${VERSION}.tgz ${GENERATED}/artifacts/

    if [[ -z $(git config --global user.email) ]]; then
      git config --global user.email "ci@cloudfoundry.org"
    fi

    if [[ -z $(git config --global user.name) ]]; then
      git config --global user.name "CI Bot"
    fi

    git add -A
    git status
    git commit -m "release v${VERSION}"
  else
    export SHA1="dummy-sha"
    echo "SHA1=$SHA1"
  fi

  echo "${VERSION}" > ${GENERATED}/tag

  cat >> ${GENERATED}/changelog.md <<EOF

## Deployment

\`\`\`yaml
releases:
- name: app-autoscaler
  version: $VERSION
  url: https://storage.googleapis.com/app-autoscaler-releases/releases/app-autoscaler-v${VERSION}.tgz
  sha1: $SHA1
\`\`\`
EOF

  cat ${GENERATED}/changelog.md
  
  git status
popd

cp -a app-autoscaler-release ${REPO_OUT}

ls -la ${REPO_OUT}
