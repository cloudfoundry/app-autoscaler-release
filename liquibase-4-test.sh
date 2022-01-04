#!/bin/bash

set -euo pipefail

git restore src/db/pom.xml
rm -f liquibase.properties

make clean
# set liquibase to 3.6.3
sed -i 's/3.10.3/3.6.3/' src/db/pom.xml
make init-db | grep -v '##'
# set liquibase to 4.x.x

sed -i 's/3.6.3/4.6.2/' src/db/pom.xml
pushd src/db
  mvn clean package -DskipTests
popd

rm target/init-db-postgres

echo "liquibase.hub.mode=off" > liquibase.properties

make init-db | grep -v '##'

psql postgres://postgres:postgres@localhost/autoscaler\?sslmode=disable -c "select id,author,liquibase,filename,exectype from databasechangelog;"
make clean
make init-db | grep -v '##'
psql postgres://postgres:postgres@localhost/autoscaler\?sslmode=disable -c "select id,author,liquibase,filename,exectype from databasechangelog;"
