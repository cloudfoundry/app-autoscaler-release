#!/bin/bash
# shellcheck disable=SC2086

set -euo pipefail

export DB_HOST="localhost"

pg_ctlcluster "$(pg_lsclusters -j | jq -r '.[0].version')" main start

psql postgres://postgres@127.0.0.1:5432 -c 'DROP DATABASE IF EXISTS autoscaler'
psql postgres://postgres@127.0.0.1:5432 -c 'CREATE DATABASE autoscaler'

pushd app-autoscaler-release

  CI=true make test

popd
