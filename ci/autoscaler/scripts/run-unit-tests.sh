#! /usr/bin/env bash

# shellcheck disable=SC2086

set -eu -o pipefail


# 🚧 To-do: Check here and in the Makefile if we can skip that db-setup. It is strange that we run the db for unit-tests!

export DB_HOST="localhost"

# 🚀 Initialise and start the PostgreSQL DBMS-server
#
# devbox makes sure that the environment-variables PGHOST and PGDATA are set appropriately.
#
# It is assumed that this test runs in isolation. Consequently the database to run the tests on is
# not already existing.
initdb
# pg_ctl will not work as it is not aware of where to create the socket.
devbox services up postgresql --background
createuser tests-pg # --superuser  # As this will only be used in a CI-job, we can allow everything within the image.

# pg_ctlcluster "$(pg_lsclusters -j | jq --raw-output '.[0].version')" main start

# psql 'postgres://postgres@127.0.0.1:5432' --command='DROP DATABASE IF EXISTS autoscaler'
createdb tests-pg # Needed to be done like this, because 'tests-pg' does not have the required priviledges.
# psql 'postgres://postgres@127.0.0.1:5432' --command='CREATE DATABASE autoscaler'

CI='true' make --directory='./app-autoscaler-release' test

devbox services stop postgresql
