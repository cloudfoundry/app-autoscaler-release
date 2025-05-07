#! /usr/bin/env bash

# shellcheck disable=SC2086

set -eu -o pipefail

# ==================== 🚧 To-do: start debugging ====================
echo "1. Name des Skripts (wie aufgerufen): $0"
echo "2. Tatsächlicher Pfad des Skripts: $BASH_SOURCE"
echo "3. Aufrufargumente: $@"
echo "4. Anzahl der Argumente: $#"
echo "5. Vollständiger Aufrufbefehl:"
ps -o args= -p $$
# ==================== 🚧 To-do: end debugging ====================

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
source "${script_dir}/common.sh"

export DB_HOST='localhost' # 🚧 To-do: Can we skip this?

ci_prepare_postgres_db # It is assumed that this test runs in isolation. Consequently the database
											 # to run the tests on is not already existing.
trap 'devbox services stop postgresql' EXIT # 🚧 To-do: Can we avoid the `--config`-parameter?
# trap 'devbox services --config='/code' stop postgresql' EXIT

CI='true' make --directory='./app-autoscaler-release' test
