#! /usr/bin/env bash
# shellcheck disable=SC2086
set -eu -o pipefail

make --directory="${1}" ${TARGETS}
