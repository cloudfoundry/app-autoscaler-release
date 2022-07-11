#!/usr/bin/env bash
[ -n "${DEBUG}" ] && set -x
set -euo pipefail

mkdir -p 'generated-release'
previous_version="$(cat gh-release/tag)"
generated="$(realpath generated-release)"

function create_release() {
   set -e
   local VERSION=$1
   local generated=$2
   bosh create-release \
        --final \
        --version "$VERSION" \
        --tarball="app-autoscaler-v${VERSION}.tgz"

    RELEASE_TGZ="app-autoscaler-v${VERSION}.tgz"
    RELEASE_SHA256="$(sha256sum "${RELEASE_TGZ}" | head -n1 | awk '{print $1}')"
    mkdir -p "${generated}/artifacts"
    mv "app-autoscaler-v${VERSION}.tgz" "${generated}/artifacts/"
}

function create_tests() {
  set -e
  local VERSION=$1
  local generated=$2
  ACCEPTANCE_TESTS_FILE="${generated}/artifacts/app-autoscaler-acceptance-tests-v${VERSION}.tgz"
  tar --create --auto-compress --directory='./src' --file="${ACCEPTANCE_TESTS_FILE}" 'acceptance'
  ACCEPTANCE_SHA256=$(sha256sum "${ACCEPTANCE_TESTS_FILE}" | head -n1 | awk '{print $1}')
}

pushd 'app-autoscaler-release' > /dev/null
  # generate the private.yml file with the credentials
  cat > 'config/private.yml' <<EOF
---
blobstore:
  options:
    credentials_source: static
    json_key:
EOF
  echo 'Generating private.yml...'
  yq eval -i '.blobstore.options.json_key = strenv(UPLOADER_KEY)' config/private.yml


  LAST_COMMIT_SHA="$(git rev-parse HEAD)"
  echo "Generating release including commits up to: ${LAST_COMMIT_SHA}"
  pushd src/changelog > /dev/null
    go run main.go \
      --changelog-file "${generated}/changelog.md" \
      --last-commit-sha-id "${LAST_COMMIT_SHA}"\
      --prev-rel-tag "${previous_version}"\
      --version-file "${generated}/name"
  popd

  VERSION=$(cat "${generated}/name")
  export VERSION

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
    git commit -m "Updated release version to ${VERSION} in golangapiserver"

    create_release "${VERSION}" "${generated}"
    create_tests "${VERSION}" "${generated}"

    git add -A
    git status
    git commit -m "release v${VERSION}"
  else
    export RELEASE_SHA256="dummy-sha"
    export ACCEPTANCE_SHA256="dummy-sha"
  fi
  echo "${VERSION}" > "${generated}/tag"
  cat >> "${generated}/changelog.md" <<EOF

## Deployment

\`\`\`yaml
releases:
- name: app-autoscaler
  version: ${VERSION}
  url: https://storage.googleapis.com/app-autoscaler-releases/releases/app-autoscaler-v${VERSION}.tgz
  sha1: sha256:${RELEASE_SHA256}
- name: app-autoscaler-acceptance-tests
  version: ${VERSION}
  url: https://storage.googleapis.com/app-autoscaler-releases/releases/app-autoscaler-acceptance-tests-v${VERSION}.tgz
  sha1: sha256:${ACCEPTANCE_SHA256}
\`\`\`
EOF
  echo "---------- Changelog file ----------"
  cat "${generated}/changelog.md"
  echo "---------- end file ----------"
popd

cp -a app-autoscaler-release "${REPO_OUT}"
