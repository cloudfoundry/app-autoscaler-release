#!/usr/bin/env bash
set -euo pipefail

function create_release() {
   set -e
   local VERSION=$1
   local GENERATED=$2
   bosh create-release \
        --force \
        --version "$VERSION" \
        --tarball=app-autoscaler-v$VERSION.tgz

    RELEASE_TGZ=app-autoscaler-v$VERSION.tgz
    RELEASE_SHA1="$(sha1sum $RELEASE_TGZ | head -n1 | awk '{print $1}')"
    mkdir -p ${GENERATED}/artifacts
    mv app-autoscaler-v${VERSION}.tgz ${GENERATED}/artifacts/
}

function create_tests() {
  set -e
  local VERSION=$1
  local GENERATED=$2
  ACCEPTANCE_TESTS_FILE="${GENERATED}/artifacts/app-autoscaler-acceptance-tests-v${VERSION}.tgz"
  tar --create --auto-compress --directory='./src' \
  --file="${ACCEPTANCE_TESTS_FILE}" 'acceptance'
  ACCEPTANCE_SHA1=$(sha1sum "${ACCEPTANCE_TESTS_FILE}" | head -n1 | awk '{print $1}')
}

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
  echo "Generating private.yml..." 
  yq eval -i '.blobstore.options.json_key = strenv(UPLOADER_KEY)' config/private.yml


  pushd src/changelog
    RECOMMENDED_VERSION_FILE=${GENERATED}/name OUTPUT_FILE=${GENERATED}/changelog.md go run main.go
  popd

  export VERSION=$(cat ${GENERATED}/name)

  yq eval -i '.properties."autoscaler.apiserver.info.build".default = strenv(VERSION)' jobs/golangapiserver/spec

  echo "Displaying diff..."
  export GIT_PAGER=cat
  git diff
  
  if [ "${PERFORM_BOSH_RELEASE}" == "true" ]; then
    # FIXME these should be configurable variables
    if [[ -z $(git config --global user.email) ]]; then
      git config --global user.email "ci@cloudfoundry.org"
    fi

    # FIXME these should be configurable variables
    if [[ -z $(git config --global user.name) ]]; then
      git config --global user.name "CI Bot"
    fi

    git add jobs/golangapiserver/spec
    git commit -m "Updated release version to $VERSION in golangapiserver"

    create_release $VERSION $GENERATED

    git add -A
    git status
    git commit -m "release v${VERSION}"
  else
    export RELEASE_SHA1="dummy-sha"
    echo "SHA1=$RELEASE_SHA1"
  fi

  create_tests ${VERSION} ${GENERATED}

  echo "${VERSION}" > ${GENERATED}/tag

  cat >> ${GENERATED}/changelog.md <<EOF

## Deployment

\`\`\`yaml
releases:
- name: app-autoscaler
  version: ${VERSION}
  url: https://storage.googleapis.com/app-autoscaler-releases/releases/app-autoscaler-v${VERSION}.tgz
  sha1: ${RELEASE_SHA1}
- name: app-autoscaler-acceptance-tests
  version: ${VERSION}
  url: https://storage.googleapis.com/app-autoscaler-releases/releases/app-autoscaler-acceptance-tests-v${VERSION}.tgz
  sha1: ${ACCEPTANCE_SHA1}
\`\`\`
EOF

  cat ${GENERATED}/changelog.md
popd

cp -a app-autoscaler-release ${REPO_OUT}

