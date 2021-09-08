#!/bin/bash
set -euo pipefail

pg_ctlcluster 10 main start

psql postgres://postgres@127.0.0.1:5432 -c 'DROP DATABASE IF EXISTS autoscaler'
psql postgres://postgres@127.0.0.1:5432 -c 'CREATE DATABASE autoscaler'

pushd app-autoscaler-release/src/app-autoscaler

  POSTGRES_OPTS='--username=postgres --url=jdbc:postgresql://127.0.0.1/autoscaler --driver=org.postgresql.Driver'

  make -C src/autoscaler buildtools
  ./scheduler/scripts/generate_unit_test_certs.sh
  ./scripts/generate_test_certs.sh

  mvn package --no-transfer-progress -Dmaven.test.skip=true -DskipTests

  java -cp 'db/target/lib/*' liquibase.integration.commandline.Main $POSTGRES_OPTS --changeLogFile=src/autoscaler/api/db/api.db.changelog.yml update
  java -cp 'db/target/lib/*' liquibase.integration.commandline.Main $POSTGRES_OPTS --changeLogFile=src/autoscaler/api/db/servicebroker.db.changelog.yaml update
  java -cp 'db/target/lib/*' liquibase.integration.commandline.Main $POSTGRES_OPTS --changeLogFile=scheduler/db/scheduler.changelog-master.yaml update
  java -cp 'db/target/lib/*' liquibase.integration.commandline.Main $POSTGRES_OPTS --changeLogFile=scheduler/db/quartz.changelog-master.yaml update
  java -cp 'db/target/lib/*' liquibase.integration.commandline.Main $POSTGRES_OPTS --changeLogFile=src/autoscaler/metricsserver/db/metricscollector.db.changelog.yml update
  java -cp 'db/target/lib/*' liquibase.integration.commandline.Main $POSTGRES_OPTS --changeLogFile=src/autoscaler/eventgenerator/db/dataaggregator.db.changelog.yml update
  java -cp 'db/target/lib/*' liquibase.integration.commandline.Main $POSTGRES_OPTS --changeLogFile=src/autoscaler/scalingengine/db/scalingengine.db.changelog.yml update
  java -cp 'db/target/lib/*' liquibase.integration.commandline.Main $POSTGRES_OPTS --changeLogFile=src/autoscaler/operator/db/operator.db.changelog.yml update

  export DBURL="postgres://postgres@localhost/autoscaler?sslmode=disable"

  make -C src/autoscaler build
  mvn package --no-transfer-progress -Dmaven.test.skip=true -DskipTests
  make -C src/autoscaler integration

popd
