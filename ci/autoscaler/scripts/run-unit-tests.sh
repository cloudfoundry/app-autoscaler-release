#! /usr/bin/env bash

# shellcheck disable=SC2086

set -eu -o pipefail

export DB_HOST='localhost' # 🚧 To-do: Can we skip this?
pg_ctlcluster "$(pg_lsclusters -j | jq -r '.[0].version')" main start
psql postgres://postgres@127.0.0.1:5432 -c 'DROP DATABASE IF EXISTS autoscaler'
psql postgres://postgres@127.0.0.1:5432 -c 'CREATE DATABASE autoscaler'

CI='true' make --directory='./app-autoscaler-release' test
