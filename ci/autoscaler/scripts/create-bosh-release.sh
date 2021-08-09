#!/bin/bash
set -euo pipefail

if [ -n "$PRE_RELEASE_SCRIPT" ]; then
  pushd release
    $PRE_RELEASE_SCRIPT
  popd
fi

pushd release
  RELEASE_VERSION="0.0.0+$(git rev-parse --short HEAD)"
  bosh create-release \
	  --force \
	  --version=${RELEASE_VERSION} \
	  --tarball=../generated-release/app-autoscaler-${RELEASE_VERSION}.tgz
popd

ls -lah generated-release/app-autoscaler-${RELEASE_VERSION}.tgz
