#! /usr/bin/env bash

# shellcheck disable=SC2086

set -eu -o pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
source "${script_dir}/common.sh"

ci_prepare_postgres_db # It is assumed that this test runs in isolation. Consequently the database
											 # to run the tests on is not already existing.
trap 'devbox services stop postgresql' EXIT # 🚧 To-do: Can we avoid the `--config`-parameter?
# trap 'devbox services --config='/code' stop postgresql' EXIT

CI='true' make --directory='app-autoscaler-release' integration
