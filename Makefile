SHELL := /bin/bash
.SHELLFLAGS = -euo pipefail -c
MAKEFLAGS = -s
go_modules:= $(shell  find . -maxdepth 3 -name "*.mod" -exec dirname {} \; | sed 's|\./src/||' | sort)
all_modules:= $(go_modules) db scheduler
.SHELLFLAGS := -eu -o pipefail -c ${SHELLFLAGS}
MVN_OPTS="-Dmaven.test.skip=true"
OS:=$(shell . /etc/lsb-release &>/dev/null && echo $${DISTRIB_ID} ||  uname  )
db_type:=postgres
DBURL := $(shell case "${db_type}" in\
			 (postgres) printf "postgres://postgres:postgres@localhost/autoscaler?sslmode=disable"; ;; \
 			 (mysql) printf "root@tcp(localhost)/autoscaler?tls=false"; ;; esac)
MYSQL_TAG := 8
POSTGRES_TAG := 12
SUITES?=broker api app
AUTOSCALER_DIR?=$(shell pwd)
lint_config:=${AUTOSCALER_DIR}/.golangci.yaml
CI_DIR?=${AUTOSCALER_DIR}/ci
CI?=false
VERSION?=0.0.testing
DEST?=build

export BUILDIN_MODE?=false
export DEBUG?=false
export ACCEPTANCE_TESTS_FILE?=${DEST}/app-autoscaler-acceptance-tests-v${VERSION}.tgz

$(shell mkdir -p target)
$(shell mkdir -p build)

list-modules:
	@echo ${go_modules}

.PHONY: check-type
check-db_type:
	@case "${db_type}" in\
	 (mysql|postgres) echo " - using bd:${db_type}"; ;;\
	 (*) echo "ERROR: db_type needs to be one of mysql|postgres"; exit 1;;\
	 esac

.PHONY: init-db
init-db: check-db_type start-db db target/init-db-${db_type}
target/init-db-${db_type}:
	@./scripts/initialise_db.sh ${db_type}
	@touch $@

.PHONY: init
init: target/init
target/init:
	@make -C src/autoscaler buildtools
	@touch $@

.PHONY: clean-autoscaler clean-java clean-vendor
clean: clean-vendor clean-autoscaler clean-java clean-targets clean-scheduler clean-certs clean-bosh-release clean-node clean-build
	@make stop-db db_type=mysql
	@make stop-db db_type=postgres
clean-build:
	@rm -rf build | true
clean-java:
	@echo " - cleaning java resources"
	@cd src && mvn clean > /dev/null && cd ..
clean-targets:
	@echo " - cleaning build target files"
	@rm target/* &> /dev/null || echo "  . Already clean"
clean-vendor:
	@echo " - cleaning vendored go"
	@find . -name "vendor" -type d -depth -exec rm -rf {} \;
clean-autoscaler:
	@make -C src/autoscaler clean
clean-scheduler:
	@echo " - cleaning scheduler test resources"
	@rm -rf src/scheduler/src/test/resources/certs
clean-certs:
	@echo " - cleaning test certs"
	@rm -f testcerts/*
clean-node:
	@echo " - cleaning node modules"
	@rm -rf src/acceptance/assets/app/nodeApp/node_modules
clean-bosh-release:
	@echo " - cleaning bosh dev releases"
	@rm -rf dev_releases
	@rm -rf .dev_builds

.PHONY: build build-test build-tests build-all $(all_modules)
build: init  $(all_modules)
build-tests: build-test
build-test: init $(addprefix test_,$(go_modules))
build-all: build build-test
db: target/db
target/db:
	@echo "# building $@"
	@cd src && mvn --no-transfer-progress package -pl db ${MVN_OPTS} && cd ..
	@touch $@
scheduler: init
	@echo "# building $@"
	@cd src && mvn --no-transfer-progress package -pl scheduler ${MVN_OPTS} && cd ..
autoscaler: init
	@make -C src/autoscaler build
changeloglockcleaner: init
	@make -C src/changeloglockcleaner build
changelog: init
	@make -C src/changelog build
$(addprefix test_,$(go_modules)):
	@echo "# Compiling '$(patsubst test_%,%,$@)' tests"
	@make -C src/$(patsubst test_%,%,$@) build_tests


.PHONY: test-certs
test-certs: target/autoscaler_test_certs target/scheduler_test_certs
target/autoscaler_test_certs:
	@./scripts/generate_test_certs.sh
	@touch $@
target/scheduler_test_certs:
	@./src/scheduler/scripts/generate_unit_test_certs.sh
	@touch $@


.PHONY: test test-autoscaler test-scheduler test-changelog test-changeloglockcleaner
test: test-autoscaler test-scheduler test-changelog test-changeloglockcleaner test-acceptance-unit
test-autoscaler: check-db_type init init-db test-certs
	@echo " - using DBURL=${DBURL} OPTS=${OPTS}"
	@make -C src/$(patsubst test-%,%,$@) test DBURL="${DBURL}" OPTS="${OPTS}"
test-autoscaler-suite: check-db_type init init-db test-certs
	@echo " - using DBURL=${DBURL} TEST=${TEST} OPTS=${OPTS}"
	@make -C src/autoscaler testsuite TEST=${TEST} DBURL="${DBURL}" OPTS="${OPTS}"
test-scheduler: check-db_type init init-db test-certs
	@cd src && mvn test --no-transfer-progress -Dspring.profiles.include=${db_type} && cd ..
test-changelog: init
	@make -C src/changelog test
test-changeloglockcleaner: init init-db test-certs
	@make -C src/changeloglockcleaner test DBURL="${DBURL}"
test-acceptance-unit:
	@make -C src/acceptance test-unit


.PHONY: start-db
start-db: check-db_type target/start-db-${db_type}_CI_${CI} waitfor_${db_type}_CI_${CI}
	@echo " SUCCESS"

.PHONY: waitfor_postgres_CI_false waitfor_postgres_CI_true
target/start-db-postgres_CI_false:
	@if [ ! "$(shell docker ps -q -f name="^${db_type}")" ]; then \
		if [ "$(shell docker ps -aq -f status=exited -f name="^${db_type}")" ]; then \
			docker rm ${db_type}; \
		fi;\
		echo " - starting docker for ${db_type}";\
		docker run -p 5432:5432 --name postgres \
			-e POSTGRES_PASSWORD=postgres \
			-e POSTGRES_USER=postgres \
			-e POSTGRES_DB=autoscaler \
			--health-cmd pg_isready \
			--health-interval 1s \
			--health-timeout 2s \
			--health-retries 10 \
			-d \
			postgres:${POSTGRES_TAG} >/dev/null;\
	else echo " - $@ already up'"; fi;
	@touch $@
target/start-db-postgres_CI_true:
	@echo " - $@ already up'"
waitfor_postgres_CI_false:
	@echo -n " - waiting for ${db_type} ."
	@COUNTER=0; until $$(docker exec postgres pg_isready &>/dev/null) || [ $$COUNTER -gt 10 ]; do echo -n "."; sleep 1; let COUNTER+=1; done;\
 	if [ $$COUNTER -gt 10 ]; then echo; echo "Error: timed out waiting for postgres. Try \"make clean\" first." >&2 ; exit 1; fi
waitfor_postgres_CI_true:
	@echo " - no ci postgres checks"

.PHONY: waitfor_mysql_CI_false waitfor_mysql_CI_true
target/start-db-mysql_CI_false:
	@if [  ! "$(shell docker ps -q -f name="^${db_type}")" ]; then \
		if [ "$(shell docker ps -aq -f status=exited -f name="^${db_type}")" ]; then \
			docker rm ${db_type}; \
		fi;\
		echo " - starting docker for ${db_type}";\
		docker pull mysql:${MYSQL_TAG}; \
		docker run -p 3306:3306  --name mysql \
			-e MYSQL_ALLOW_EMPTY_PASSWORD=true \
			-e MYSQL_DATABASE=autoscaler \
			-d \
			mysql:${MYSQL_TAG} \
			>/dev/null;\
	else echo " - $@ already up"; fi;
	@touch $@
target/start-db-mysql_CI_true:
	@echo " - $@ already up'"
waitfor_mysql_CI_false:
	@echo -n " - waiting for ${db_type} ."
	@until docker exec mysql mysqladmin ping &>/dev/null ; do echo -n "."; sleep 1; done
	@echo " SUCCESS"
	@echo -n " - Waiting for table creation ."
	@until [[ ! -z `docker exec mysql mysql -qfsBe "SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME='autoscaler'" 2> /dev/null` ]]; do echo -n "."; sleep 1; done
waitfor_mysql_CI_true:
	@echo -n " - Waiting for table creation "
	@which mysql >/dev/null &&\
	 {\
	   T=0;\
	   until [[ ! -z "$(shell mysql -u "root" -h "127.0.0.1"  --port=3306 -qfsBe "SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME='autoscaler'" 2> /dev/null)" ]]\
	     || [[ $${T} -gt 30 ]];\
	   do echo -n "."; sleep 1; T=$$((T+1)); done;\
	 }
	@[ ! -z "$(shell mysql -u "root" -h "127.0.0.1" --port=3306 -qfsBe "SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME='autoscaler'"  2> /dev/null)" ]\
	  || { echo "ERROR: Mysql timed out creating database"; exit 1; }


.PHONY: stop-db
stop-db: check-db_type
	@echo " - Stopping ${db_type}"
	@rm target/start-db-${db_type} &> /dev/null || echo " - Seems the make target was deleted stopping anyway!"
	@docker rm -f ${db_type} &> /dev/null || echo " - we could not stop and remove docker named '${db_type}'"

.PHONY: integration
integration: build init-db test-certs
	@echo " - using DBURL=${DBURL} OPTS=${OPTS}"
	make -C src/autoscaler integration DBURL="${DBURL}" OPTS="${OPTS}"


.PHONY:lint $(addprefix lint_,$(go_modules))
lint: golangci-lint_check $(addprefix lint_,$(go_modules)) eslint rubocop

golangci-lint_check:
	@current_version=$(shell golangci-lint version | cut -d " " -f 4);\
	current_major_version=$(shell golangci-lint version | cut -d " " -f 4| sed -E 's/v*([0-9]+\.[0-9]+)\..*/\1/');\
	expected_version=$(shell cat src/autoscaler/go.mod | grep golangci-lint  | cut -d " " -f 2 | sed -E 's/v([0-9]+\.[0-9]+)\..*/\1/');\
	if [ "$${current_major_version}" != "$${expected_version}" ]; then \
        echo "ERROR: Expected to have golangci-lint version '$${expected_version}.x' but we have $${current_version}";\
        exit 1;\
    fi

rubocop:
	@echo " - ruby scripts"
	@bundle install
	@bundle exec rubocop ./spec ./packages

.PHONY: eslint
eslint:
	@echo " - linting testApp"
	@cd src/acceptance/assets/app/nodeApp && npm install && npm run lint

$(addprefix lint_,$(go_modules)): lint_%:
	@echo " - linting: $(patsubst lint_%,%,$@)"
	@pushd src/$(patsubst lint_%,%,$@) >/dev/null && golangci-lint run --path-prefix=src/$(patsubst lint_%,%,$@) --config ${lint_config} ${OPTS}

.PHONY: spec-test
spec-test:
	bundle install
	bundle exec rspec

.PHONY: release
bosh-release: mod-tidy vendor scheduler db build/autoscaler-test.tgz
build/autoscaler-test.tgz:
	@echo " - building bosh release into build/autoscaler-test.tgz"
	@mkdir -p build
	@bosh create-release --force --timestamp-version --tarball=build/autoscaler-test.tgz

.PHONY: vendor-app
vendor-app: target/vendor-app
target/vendor-app:
	@echo " - installing node modules to package node app"
	@cd src/acceptance/assets/app/nodeApp > /dev/null\
	 && npm install --production\
	 && npm prune --production
	@touch $@

.PHONY: acceptance-release
acceptance-release: mod-tidy vendor vendor-app
	@echo " - building acceptance test release '${VERSION}' to dir: '${DEST}' "
	@mkdir -p ${DEST}
	@tar --create --auto-compress --directory="src" --file="${ACCEPTANCE_TESTS_FILE}" 'acceptance'
.PHONY: mod-tidy
mod-tidy:
	@for folder in $$(find . -maxdepth 3 -name "go.mod" -exec dirname {} \;);\
	do\
	   cd $${folder}; echo " - go mod tidying '$${folder}'"; go mod tidy; cd - >/dev/null;\
	done

.PHONY: mod-download
mod-download:
	@for folder in $$(find . -maxdepth 3 -name "go.mod" -exec dirname {} \;);\
	do\
	   cd $${folder}; echo " - go mod download '$${folder}'"; go mod download; cd - >/dev/null;\
	done

.PHONY: vendor
vendor:
	@for folder in $$(find . -maxdepth 3 -name "go.mod" -exec dirname {} \;);\
	do\
	   cd $${folder}; echo " - go mod vendor '$${folder}'"; go mod vendor; cd - >/dev/null;\
	done

.PHONY: fakes
fakes:
	@make -C src/autoscaler fakes

# https://github.com/golang/tools/blob/master/gopls/doc/workspace.md
.PHONY: workspace
workspace:
	[ -e go.work ] || go work init
	go work use $(addprefix ./src/,$(go_modules))

.PHONY: uuac
uaac:
	which uaac || gem install cf-uaac

.PHONY: deploy-autoscaler deploy-register-cf deploy-autoscaler-bosh
deploy-autoscaler: mod-tidy vendor uaac db scheduler deploy-autoscaler-bosh deploy-register-cf
deploy-register-cf:
	echo " - registering broker with cf"
	[ "$${BUILDIN_MODE}" == "false" ] && { ${CI_DIR}/autoscaler/scripts/register-broker.sh; } || echo " - Not registering broker due to buildin mode enabled"
deploy-autoscaler-bosh:
	echo " - deploying autoscaler"
	${CI_DIR}/autoscaler/scripts/deploy-autoscaler.sh

deploy-prometheus:
	@export DEPLOYMENT_NAME=prometheus;\
	export BBL_STATE_PATH=$${BBL_STATE_PATH:-$(shell realpath "../app-autoscaler-env-bbl-state/bbl-state/")};\
	${CI_DIR}/infrastructure/scripts/deploy-prometheus.sh;

.PHONY: acceptance-tests
acceptance-tests: vendor-app
	${CI_DIR}/autoscaler/scripts/run-acceptance-tests.sh;

.PHONY: deploy-cleanup
deploy-cleanup:
	@echo " - Cleaning up deployment '${DEPLOYMENT_NAME}'";\
	${CI_DIR}/autoscaler/scripts/cleanup-autoscaler.sh;


.PHONY: cleanup-concourse
cleanup-concourse:
	@${CI_DIR}/autoscaler/scripts/cleanup-concourse.sh

.PHONY: cf-login
cf-login:
	@${CI_DIR}/autoscaler/scripts/cf-login.sh

.PHONY: ssh-autoscaler
ssh-autoscaler:
	@${CI_DIR}/autoscaler/scripts/ssh-autoscaler.sh

.PHONY: setup-performance
setup-performance:
	export GINKGO_OPTS="";\
	export SKIP_TEARDOWN=true;\
	export NODES=1;\
	export DEPLOYMENT_NAME="autoscaler-performance";\
	export SUITES="setup_performance";\
	${CI_DIR}/autoscaler/scripts/run-acceptance-tests.sh;\

.PHONY: run-performance
run-performance:
	export GINKGO_OPTS="";\
	export SKIP_TEARDOWN=true;\
	export NODES=1;\
	export DEPLOYMENT_NAME="autoscaler-performance";\
	export SUITES="run_performance";\
	${CI_DIR}/autoscaler/scripts/run-acceptance-tests.sh;\

.PHONY: run-act
run-act:
	${AUTOSCALER_DIR}/scripts/run_act.sh;\

package-specs: mod-tidy vendor
	@echo " - Updating the package specs"
	@./scripts/sync-package-specs


## Prometheus Alerts
.PHONY: silence-alerts
silence-alerts:
	export SILENCE_TIME_MINS=480;\
	echo " - Silencing deployment '${DEPLOYMENT_NAME} 8 hours'"
	${CI_DIR}/autoscaler/scripts/silence_prometheus_alert.sh BOSHJobProcessExtendedUnhealthy ;\
	${CI_DIR}/autoscaler/scripts/silence_prometheus_alert.sh BOSHJobProcessUnhealthy ;\
	${CI_DIR}/autoscaler/scripts/silence_prometheus_alert.sh BOSHJobExtendedUnhealthy ;\
	${CI_DIR}/autoscaler/scripts/silence_prometheus_alert.sh BOSHJobProcessUnhealthy ;\
	${CI_DIR}/autoscaler/scripts/silence_prometheus_alert.sh BOSHJobEphemeralDiskPredictWillFill ;\
	${CI_DIR}/autoscaler/scripts/silence_prometheus_alert.sh BOSHJobUnhealthy ;
