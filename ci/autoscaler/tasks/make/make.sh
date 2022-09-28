#!/bin/bash
# shellcheck disable=SC2086
set -euo pipefail

make -C $1 ${TARGETS}
