#! /usr/bin/env bash

# shellcheck disable=SC2086

set -eu -o pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
source "${script_dir}/common.sh"

ci_prepare_postgres_db # It is assumed that this test runs in isolation. Consequently the database
											 # to run the tests on is not already existing.
trap 'devbox services stop postgresql' EXIT

CI='true' make --directory='app-autoscaler-release' integration


# pg_ctlcluster "$(pg_lsclusters -j | jq --raw-output '.[0].version')" main start

# psql 'postgres://postgres@127.0.0.1:5432' --command='DROP DATABASE IF EXISTS autoscaler'
# psql 'postgres://postgres@127.0.0.1:5432' --command='CREATE DATABASE autoscaler'

# pushd app-autoscaler-release
#		CI='true' make integration
# popd
