#!/bin/bash

set -euo pipefail

git restore src/db/pom.xml
rm -f liquibase.properties

make clean

pushd src/db
  # set liquibase to 3.6.3
  mvn versions:use-dep-version -Dincludes=org.liquibase:liquibase-core -DdepVersion=3.6.3 -DforceVersion=true
popd

make init-db | grep -v '##'

pushd src/db
  # set liquibase to 4.x.x
  mvn versions:use-dep-version -Dincludes=org.liquibase:liquibase-core -DdepVersion=4.6.2 -DforceVersion=true
  mvn clean package -DskipTests
popd

rm target/init-db-postgres

echo "liquibase.hub.mode=off" > liquibase.properties

make init-db | grep -v '##'

psql postgres://postgres:postgres@localhost/autoscaler\?sslmode=disable -c "select id,author,liquibase,filename,exectype from databasechangelog;"
make clean
make init-db | grep -v '##'
psql postgres://postgres:postgres@localhost/autoscaler\?sslmode=disable -c "select id,author,liquibase,filename,exectype from databasechangelog;"
