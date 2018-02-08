#!/bin/bash
set -x
set -e

service postgresql start

rm -rf /usr/local/var/postgres
mkdir -p /usr/local/var/postgres
chown postgres /usr/local/var/postgres
nohup su - postgres -c "/usr/lib/postgresql/$PG_MAJOR/bin/pg_ctl init -D /usr/local/var/postgres && /usr/lib/postgresql/$PG_MAJOR/bin/postgres -D /usr/local/var/postgres" > pg.log 2>&1 &
echo "sleep 30 seconds waiting for postgres is ready"
sleep 30

psql postgres://postgres@127.0.0.1:5432 -c 'DROP DATABASE IF EXISTS autoscaler'
psql postgres://postgres@127.0.0.1:5432 -c 'CREATE DATABASE autoscaler'

cd app-autoscaler-release/src/app-autoscaler

POSTGRES_OPTS='--username=postgres --url=jdbc:postgresql://127.0.0.1/autoscaler --driver=org.postgresql.Driver'

mvn package
java -cp 'db/target/lib/*' liquibase.integration.commandline.Main $POSTGRES_OPTS --changeLogFile=api/db/api.db.changelog.yml update
java -cp 'db/target/lib/*' liquibase.integration.commandline.Main $POSTGRES_OPTS --changeLogFile=servicebroker/db/servicebroker.db.changelog.json update
java -cp 'db/target/lib/*' liquibase.integration.commandline.Main $POSTGRES_OPTS --changeLogFile=scheduler/db/scheduler.changelog-master.yaml update
java -cp 'db/target/lib/*' liquibase.integration.commandline.Main $POSTGRES_OPTS --changeLogFile=scheduler/db/quartz.changelog-master.yaml update
java -cp 'db/target/lib/*' liquibase.integration.commandline.Main $POSTGRES_OPTS --changeLogFile=src/autoscaler/metricscollector/db/metricscollector.db.changelog.yml update
java -cp 'db/target/lib/*' liquibase.integration.commandline.Main $POSTGRES_OPTS --changeLogFile=src/autoscaler/eventgenerator/db/dataaggregator.db.changelog.yml update
java -cp 'db/target/lib/*' liquibase.integration.commandline.Main $POSTGRES_OPTS --changeLogFile=src/autoscaler/scalingengine/db/scalingengine.db.changelog.yml update

npm set progress=false

pushd api
npm install
npm test
popd

pushd servicebroker
npm install
npm test
popd

export GOPATH=$PWD
export PATH=$GOPATH/bin:$PATH
go install github.com/onsi/ginkgo/ginkgo

curl -L -o ./consul-0.7.5.zip "https://releases.hashicorp.com/consul/0.7.5/consul_0.7.5_linux_amd64.zip"
unzip ./consul-0.7.5.zip -d $GOPATH/bin
rm ./consul-0.7.5.zip

DBURL=postgres://postgres@localhost/autoscaler?sslmode=disable ginkgo -r -race -randomizeAllSpecs src/autoscaler

pushd scheduler
mvn test
popd
