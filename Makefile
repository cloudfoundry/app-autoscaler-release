SHELL := /bin/bash
go_modules:= acceptance autoscaler changelog changeloglockcleaner
all_modules:= $(go_modules) db scheduler
lint_config:=${PWD}/.golangci.yaml
.SHELLFLAGS := -eu -o pipefail -c ${SHELLFLAGS}
MVN_OPTS="-Dmaven.test.skip=true"
OS:=$(shell . /etc/lsb-release &>/dev/null && echo $${DISTRIB_ID} ||  uname  )
db_type:=postgres
DBURL := $(shell case "${db_type}" in\
			 (postgres) printf "postgres://postgres:postgres@localhost/autoscaler?sslmode=disable"; ;; \
 			 (mysql) printf "root@tcp(localhost)/autoscaler?tls=false"; ;; esac)
MYSQL_TAG := 8
POSTGRES_TAG := 12

CI?=false
$(shell mkdir -p target)

.PHONY: check-type
check-db_type:
	@case "${db_type}" in\
	 (mysql|postgres) echo "Using bd:${db_type}"; ;;\
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
clean: clean-vendor clean-autoscaler clean-java clean-targets
	@make stop-db db_type=mysql
	@make stop-db db_type=postgres
	@make clean-targets
clean-java:
	@echo " - cleaning java resources"
	@cd src && mvn clean > /dev/null && cd ..
clean-targets:
	@echo " - cleaning build target files"
	@rm target/* &> /dev/null || echo "  - Already clean"
clean-vendor:
	@echo " - cleaning vendored go"
	@find . -name "vendor" -type d -exec rm -rf {} \;
clean-autoscaler:
	@make -C src/autoscaler clean

.PHONY: build build-test build-all $(all_modules)
build: init  $(all_modules)
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
test: test-autoscaler test-scheduler test-changelog test-changeloglockcleaner test-acceptance
test-autoscaler: check-db_type init init-db test-certs
	@echo " - using DBURL=${DBURL}"
	@make -C src/$(patsubst test-%,%,$@) test DBURL="${DBURL}"
test-autoscaler-suite: check-db_type init init-db test-certs
	@echo " - using DBURL=${DBURL} TEST=${TEST}"
	@echo " - using TEST=${TEST}"
	@make -C src/autoscaler testsuite TEST=${TEST} DBURL="${DBURL}"
test-scheduler: check-db_type init init-db test-certs
	@cd src && mvn test --no-transfer-progress -Dspring.profiles.include=${db_type} && cd ..
test-changelog: init
	@make -C src/changelog test
test-changeloglockcleaner: init init-db test-certs
	@make -C src/changeloglockcleaner test DBURL="${DBURL}"
test-acceptance:
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
	@until docker exec postgres pg_isready &>/dev/null ; do echo -n "."; sleep 1; done
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
	make -C src/autoscaler integration DBURL="${DBURL}"

.PHONY: golangci-lint lint $(addprefix lint_,$(go_modules))
lint: golangci-lint_check golangci-lint $(addprefix lint_,$(go_modules))

golangci-lint_check:
	@current_version=$(shell golangci-lint version | cut -d " " -f 4);\
	current_major_version=$(shell golangci-lint version | cut -d " " -f 4| sed -E 's/v*([0-9]+\.[0-9]+)\..*/\1/');\
	expected_version=$(shell cat src/autoscaler/go.mod | grep golangci-lint  | cut -d " " -f 2 | sed -E 's/v([0-9]+\.[0-9]+)\..*/\1/');\
	if [ "$${current_major_version}" != "$${expected_version}" ]; then \
        echo "ERROR: Expected to have golangci-lint version '$${expected_version}.x' but we have $${current_version}";\
        exit 1;\
    fi

golangci-lint:
	@make -C src/autoscaler golangci-lint

rubocop:
	bundle exec rubocop -a

$(addprefix lint_,$(go_modules)): lint_%:
	@echo " - linting: $(patsubst lint_%,%,$@)"
	@pushd src/$(patsubst lint_%,%,$@) >/dev/null && golangci-lint --config ${lint_config} run ${OPTS}

spec-test:
	bundle exec rspec
release:
	./scripts/update
	bosh create-release --force --timestamp-version --tarball=${name}-${version}.tgz

mod-tidy:
	@for folder in $$(find . -name "go.mod" -exec dirname {} \;);\
	do\
	   cd $${folder}; echo "- go mod tidying '$${folder}'"; go mod tidy; cd - >/dev/null;\
	done

