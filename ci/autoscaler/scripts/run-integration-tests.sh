#!/bin/bash
set -euo pipefail

pg_ctlcluster 10 main start

psql postgres://postgres@127.0.0.1:5432 -c 'DROP DATABASE IF EXISTS autoscaler'
psql postgres://postgres@127.0.0.1:5432 -c 'CREATE DATABASE autoscaler'

pushd app-autoscaler-release

  POSTGRES_OPTS='--username=postgres --url=jdbc:postgresql://127.0.0.1/autoscaler --driver=org.postgresql.Driver'

  make -C src/autoscaler buildtools
  ./src/scheduler/scripts/generate_unit_test_certs.sh
  ./scripts/generate_test_certs.sh

  pushd src
    mvn package --no-transfer-progress -Dmaven.test.skip=true -DskipTests
    echo "liquibase.hub.mode=off" > liquibase.properties

    java -cp 'db/target/lib/*' liquibase.integration.commandline.Main $POSTGRES_OPTS --changeLogFile=autoscaler/api/db/api.db.changelog.yml update
    java -cp 'db/target/lib/*' liquibase.integration.commandline.Main $POSTGRES_OPTS --changeLogFile=autoscaler/api/db/servicebroker.db.changelog.yaml update
    java -cp 'db/target/lib/*' liquibase.integration.commandline.Main $POSTGRES_OPTS --changeLogFile=scheduler/db/scheduler.changelog-master.yaml update
    java -cp 'db/target/lib/*' liquibase.integration.commandline.Main $POSTGRES_OPTS --changeLogFile=scheduler/db/quartz.changelog-master.yaml update
    java -cp 'db/target/lib/*' liquibase.integration.commandline.Main $POSTGRES_OPTS --changeLogFile=autoscaler/metricsserver/db/metricscollector.db.changelog.yml update
    java -cp 'db/target/lib/*' liquibase.integration.commandline.Main $POSTGRES_OPTS --changeLogFile=autoscaler/eventgenerator/db/dataaggregator.db.changelog.yml update
    java -cp 'db/target/lib/*' liquibase.integration.commandline.Main $POSTGRES_OPTS --changeLogFile=autoscaler/scalingengine/db/scalingengine.db.changelog.yml update
    java -cp 'db/target/lib/*' liquibase.integration.commandline.Main $POSTGRES_OPTS --changeLogFile=autoscaler/operator/db/operator.db.changelog.yml update
  popd

  export DBURL="postgres://postgres@localhost/autoscaler?sslmode=disable"

  make -C src/autoscaler build
  make -C src/autoscaler integration

popd
