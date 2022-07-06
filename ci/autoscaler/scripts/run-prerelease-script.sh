#!/bin/bash

set -euo pipefail

pushd release
  exec "${SCRIPT_NAME}"
popd
