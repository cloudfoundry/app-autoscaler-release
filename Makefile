SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c ${SHELLFLAGS}
MAKEFLAGS = -s

go-acceptance-dir := ./src/acceptance
go-autoscaler-dir := ./src/autoscaler
go-changelog-dir := ./src/changelog
go-changeloglockcleander-dir := ./src/changeloglockcleaner
go-test-app-dir := ./src/acceptance/assets/app/go_app

go_modules := $(shell find . -maxdepth 6 -name "*.mod" -exec dirname {} \; | sed 's|\./src/||' | sort)
all_modules := $(go_modules) db scheduler

MVN_OPTS = "-Dmaven.test.skip=true"
OS := $(shell . /etc/lsb-release &>/dev/null && echo $${DISTRIB_ID} || uname)
db_type := postgres
DB_HOST := localhost
DBURL := $(shell case "${db_type}" in\
			 (postgres) printf "postgres://postgres:postgres@${DB_HOST}/autoscaler?sslmode=disable"; ;; \
				 (mysql) printf "root@tcp(${DB_HOST})/autoscaler?tls=false"; ;; esac)
DEBUG := false
MYSQL_TAG := 8
POSTGRES_TAG := 12
SUITES ?= broker api app
AUTOSCALER_DIR ?= $(shell pwd)
lint_config := ${AUTOSCALER_DIR}/.golangci.yaml
CI_DIR ?= ${AUTOSCALER_DIR}/ci
CI ?= false
VERSION ?= 0.0.testing
DEST ?= build

GOLANGCI_LINT_VERSION = v$(shell cat .tool-versions | grep golangci-lint  \
													| cut --delimiter=' ' --fields='2')

export DEBUG ?= false
export ACCEPTANCE_TESTS_FILE ?= ${DEST}/app-autoscaler-acceptance-tests-v${VERSION}.tgz
export GOWORK = off

$(shell mkdir -p target)
$(shell mkdir -p build)

.DEFAULT_GOAL := build-all

list-modules:
	@echo ${go_modules}

.PHONY: check-type
check-db_type:
	@case "${db_type}" in\
	 (mysql|postgres) echo " - using db_type:${db_type}"; ;;\
	 (*) echo "ERROR: db_type needs to be one of mysql|postgres"; exit 1;;\
	 esac

.PHONY: init-db
init-db: check-db_type start-db db target/init-db-${db_type}
target/init-db-${db_type}:
	@./scripts/initialise_db.sh ${db_type}
	@touch $@

.PHONY: clean-autoscaler clean-java clean-vendor clean-acceptance
clean: clean-vendor clean-autoscaler clean-java clean-targets clean-scheduler clean-certs clean-bosh-release clean-build clean-acceptance ## Clean all build and test artifacts
	@make stop-db db_type=mysql
	@make stop-db db_type=postgres
clean-build:
	@rm -rf build | true
clean-java:
	@echo " - cleaning java resources"
	@cd src && mvn clean > /dev/null && cd ..
clean-targets:
	@echo " - cleaning build target files"
	@rm --recursive --force target/* &> /dev/null || echo " . Already clean"
clean-vendor:
	@echo " - cleaning vendored go"
	@find . -depth -name "vendor" -type d -exec rm -rf {} \;
clean-fakes:
	@echo " - cleaning fakes"
	@find . -depth -name "fakes" -type d -exec rm -rf {} \;
clean-autoscaler:
	@make --directory='./src/autoscaler' clean
clean-scheduler:
	@echo " - cleaning scheduler test resources"
	@rm -rf src/scheduler/src/test/resources/certs
clean-certs:
	@echo " - cleaning test certs"
	@rm -f test-certs/*
	@rm --force --recursive src/scheduler/src/test/resources/certs
clean-bosh-release:
	@echo " - cleaning bosh dev releases"
	@rm -rf dev_releases
	@rm -rf .dev_builds
clean-acceptance:
	@echo ' - cleaning acceptance (⚠️ This keeps the file “src/acceptance/acceptance_config.json” if present!)'
	@rm src/acceptance/ginkgo* &> /dev/null || true
	@rm -rf src/acceptance/results &> /dev/null || true

.PHONY: build build-test build-tests build-all $(all_modules)
build: $(all_modules)
build-tests: build-test
build-test: $(addprefix test_,$(go_modules))
build-all: generate-openapi-generated-clients-and-servers build build-test build-test-app ## Build all modules and tests
db: target/db
target/db:
	@echo "# building $@"
	@cd src && mvn --no-transfer-progress package -pl db ${MVN_OPTS} && cd ..
	@touch $@
scheduler:
	@echo "# building $@"
	@cd src && mvn --no-transfer-progress package -pl scheduler ${MVN_OPTS} && cd ..
autoscaler:
	@make --directory='./src/autoscaler' build
changeloglockcleaner:
	@make --directory='./src/changeloglockcleaner' build
changelog:
	@make --directory='./src/changelog' build
$(addprefix test_,$(go_modules)):
	@echo "# Compiling '$(patsubst test_%,%,$@)' tests"
	@make --directory='./src/$(patsubst test_%,%,$@)' build_tests


.PHONY: test-certs
test-certs: target/autoscaler_test_certs src/scheduler/src/test/resources/certs
target/autoscaler_test_certs:
	@./scripts/generate_test_certs.sh
	@touch $@
src/scheduler/src/test/resources/certs:
	@./src/scheduler/scripts/generate_unit_test_certs.sh


.PHONY: test test-autoscaler test-scheduler test-changelog test-changeloglockcleaner
test: test-autoscaler test-scheduler test-changelog test-changeloglockcleaner test-acceptance-unit ## Run all unit tests
test-autoscaler: check-db_type init-db test-certs
	@echo ' - using DBURL=${DBURL} TEST=${TEST}'
	@make --directory='./src/autoscaler' test DBURL='${DBURL}' TEST='${TEST}'
test-autoscaler-suite: check-db_type init-db test-certs
	@make --directory='./src/autoscaler' testsuite TEST='${TEST}' DBURL='${DBURL}'
test-scheduler: check-db_type init-db test-certs
	@export DB_HOST=${DB_HOST}; \
	cd src && mvn test --no-transfer-progress -Dspring.profiles.include=${db_type} && cd ..
test-changelog:
	@make --directory='./src/changelog' test
test-changeloglockcleaner: init-db test-certs
	@make --directory='./src/changeloglockcleaner' test DBURL="${DBURL}"
test-acceptance-unit:
	@make --directory='./src/acceptance' test-unit
	@make --directory='./src/acceptance/assets/app/go_app' test


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
	@echo -n " - Waiting for table creation"
	@which mysql >/dev/null &&\
	{\
		T=0;\
		until [[ ! -z "$(shell mysql -u "root" -h "${DB_HOST}"  --port=3306 -qfsBe "SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME='autoscaler'" 2> /dev/null)" ]]\
			|| [[ $${T} -gt 30 ]];\
		do echo -n "."; sleep 1; T=$$((T+1)); done;\
	}
	@[ ! -z "$(shell mysql -u "root" -h "${DB_HOST}" --port=3306 -qfsBe "SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME='autoscaler'"  2> /dev/null)" ]\
		|| { echo "ERROR: Mysql timed out creating database"; exit 1; }


.PHONY: stop-db
stop-db: check-db_type
	@echo " - Stopping ${db_type}"
	@rm target/start-db-${db_type} &> /dev/null || echo " - Seems the make target was deleted stopping anyway!"
	@docker rm -f ${db_type} &> /dev/null || echo " - we could not stop and remove docker named '${db_type}'"

.PHONY: integration
integration: generate-openapi-generated-clients-and-servers build init-db test-certs ## Run all integration tests
	@echo " - using DBURL=${DBURL}"
	@make --directory='./src/autoscaler' integration DBURL="${DBURL}"


.PHONY: lint
lint: lint-go lint-ruby lint-actions lint-markdown ## Run all linters

.PHONY:lint $(addprefix lint_,$(go_modules))
lint-go: build-all $(addprefix lint_,$(go_modules))

lint-ruby:
	@echo " - ruby scripts"
	@bundle install
	@bundle exec rubocop ${RUBOCOP_OPTS} ./spec ./packages

.PHONY: lint-markdown
lint-markdown:
	@echo " - linting markdown files"
	@markdownlint-cli2 .

.PHONY: lint-actions
lint-actions:
	@echo " - linting GitHub actions"
	actionlint

$(addprefix lint_,$(go_modules)): lint_%:
	@echo " - linting: $(patsubst lint_%,%,$@)"
	@pushd src/$(patsubst lint_%,%,$@) >/dev/null && golangci-lint run --config ${lint_config} ${OPTS} --timeout 5m

.PHONY: spec-test
spec-test:
	bundle install
	bundle exec rspec

.PHONY: bosh-release
bosh-release: go-mod-tidy go-mod-vendor scheduler db build/autoscaler-test.tgz
build/autoscaler-test.tgz:
	@echo " - building bosh release into build/autoscaler-test.tgz"
	@mkdir -p build
	@bosh create-release --force --timestamp-version --tarball=build/autoscaler-test.tgz

.PHONY: generate-fakes autoscaler.generate-fakes test-app.generate-fakes
generate-fakes: autoscaler.generate-fakes test-app.generate-fakes
autoscaler.generate-fakes:
	make --directory='${go-autoscaler-dir}' generate-fakes
test-app.generate-fakes:
	make --directory='${go-test-app-dir}' generate-fakes

.PHONY: generate-openapi-generated-clients-and-servers
generate-openapi-generated-clients-and-servers:
	make --directory='${go-autoscaler-dir}' generate-openapi-generated-clients-and-servers

.PHONY: go-mod-tidy
go-mod-tidy: acceptance.go-mod-tidy autoscaler.go-mod-tidy changelog.go-mod-tidy \
						 changeloglockcleander.go-mod-tidy test-app.go-mod-tidy

.PHONY: acceptance.go-mod-tidy autoscaler.go-mod-tidy changelog.go-mod-tidy \
				changeloglockcleander.go-mod-tidy test-app.go-mod-tidy
acceptance.go-mod-tidy:
	make --directory='${go-acceptance-dir}' go-mod-tidy
autoscaler.go-mod-tidy:
	make --directory='${go-autoscaler-dir}' go-mod-tidy
changelog.go-mod-tidy:
	make --directory='${go-changelog-dir}' go-mod-tidy
changeloglockcleander.go-mod-tidy:
	make --directory='${go-changeloglockcleander-dir}' go-mod-tidy
test-app.go-mod-tidy:
	make --directory='${go-test-app-dir}' go-mod-tidy



.PHONY: mod-download
mod-download:
	@for folder in $$(find . -maxdepth 3 -name "go.mod" -exec dirname {} \;);\
	do\
		 cd $${folder}; echo " - go mod download '$${folder}'"; go mod download; cd - >/dev/null;\
	done

.PHONY: acceptance.go-mod-vendor autoscaler.go-mod-vendor changelog.go-mod-vendor \
				changeloglockcleander.go-mod-vendor
go-mod-vendor: clean-vendor acceptance.go-mod-vendor autoscaler.go-mod-vendor changelog.go-mod-vendor \
							 changeloglockcleander.go-mod-vendor
acceptance.go-mod-vendor:
	make --directory='${go-acceptance-dir}' go-mod-vendor
autoscaler.go-mod-vendor:
	make --directory='${go-autoscaler-dir}' go-mod-vendor
changelog.go-mod-vendor:
	make --directory='${go-changelog-dir}' go-mod-vendor
changeloglockcleander.go-mod-vendor:
	make --directory='${go-changeloglockcleander-dir}' go-mod-vendor

.PHONY: uuac
uaac:
	which uaac || gem install cf-uaac

.PHONY: update-uaac-nix-package
update-uaac-nix-package:
	make --directory='./nix/packages/uaac' gemset.nix

.PHONY: deploy-autoscaler deploy-register-cf deploy-autoscaler-bosh deploy-cleanup
deploy-autoscaler: go-mod-vendor uaac db scheduler deploy-autoscaler-bosh deploy-register-cf ## Deploy autoscaler to OSS dev environment
deploy-register-cf:
	echo " - registering broker with cf"
	${CI_DIR}/autoscaler/scripts/register-broker.sh
deploy-autoscaler-bosh:
	echo " - deploying autoscaler"
	DEBUG="${DEBUG}" ${CI_DIR}/autoscaler/scripts/deploy-autoscaler.sh
deploy-cleanup:
	${CI_DIR}/autoscaler/scripts/cleanup-autoscaler.sh;

bosh-release-path := ./target/bosh-releases
prometheus-bosh-release-path := ${bosh-release-path}/prometheus
$(shell mkdir -p ${prometheus-bosh-release-path})

download-prometheus-release: ${prometheus-bosh-release-path}/manifests
${prometheus-bosh-release-path}/manifests &:
	pushd '${prometheus-bosh-release-path}' > /dev/null ;\
		git clone --recurse-submodules 'https://github.com/bosh-prometheus/prometheus-boshrelease' . ;\
	popd > /dev/null


deploy-prometheus: ${prometheus-bosh-release-path}/manifests
	export DEPLOYMENT_NAME='prometheus' ;\
	export PROMETHEUS_DIR='${prometheus-bosh-release-path}' ;\
	export BBL_GCP_REGION="$$(yq eval '.jobs.[] | select(.name == "deploy-prometheus") | .plan.[] | select(.task == "deploy-prometheus") | .params.BBL_GCP_REGION' './ci/infrastructure/pipeline.yml')" ;\
	export BBL_GCP_ZONE="$$(yq eval '.jobs.[] | select(.name == "deploy-prometheus") | .plan.[] | select(.task == "deploy-prometheus") | .params.BBL_GCP_ZONE' './ci/infrastructure/pipeline.yml')" ;\
	export SLACK_WEBHOOK="$$(credhub get --name='/bosh-autoscaler/prometheus/alertmanager_slack_api_url' --quiet)" ;\
	${CI_DIR}/infrastructure/scripts/deploy-prometheus.sh;


.PHONY: mta-release
mta-release: mta-build
	@echo " - building mtar release '${VERSION}' to dir: '${DEST}' "

.PHONY: acceptance-release
acceptance-release: clean-acceptance go-mod-tidy go-mod-vendor build-test-app
	@echo " - building acceptance test release '${VERSION}' to dir: '${DEST}' "
	@mkdir -p ${DEST}
	${AUTOSCALER_DIR}/scripts/compile-acceptance-tests.sh
	@tar --create --auto-compress --directory="src" --file="${ACCEPTANCE_TESTS_FILE}" 'acceptance'

.PHONY: mta-build
mta-build:
	@echo " - building mta"
	@make --directory='./src/autoscaler' mta-build

.PHONY: build-test-app
build-test-app:
	@make --directory='./src/acceptance/assets/app/go_app' build

.PHONY: deploy-test-app
deploy-test-app:
	@make --directory='./src/acceptance/assets/app/go_app' deploy

.PHONY: build-acceptance-tests
build-acceptance-tests:
	@make --directory='./src/acceptance' build_tests

.PHONY: acceptance-tests
acceptance-tests: build-test-app acceptance-tests-config ## Run acceptance tests against OSS dev environment (requrires a previous deployment of the autoscaler)
	@make --directory='./src/acceptance' run-acceptance-tests
.PHONY: acceptance-cleanup
acceptance-cleanup:
	@make --directory='./src/acceptance' acceptance-tests-cleanup
.PHONY: acceptance-tests-config
acceptance-tests-config:
	make --directory='./src/acceptance' acceptance-tests-config

.PHONY: cleanup-concourse
cleanup-concourse:
	@${CI_DIR}/autoscaler/scripts/cleanup-concourse.sh

.PHONY: cf-login
cf-login: ## Login to OSS CF dev environment
	@${CI_DIR}/autoscaler/scripts/cf-login.sh

.PHONY: setup-performance
setup-performance: build-test-app
	export NODES=1;\
	export SUITES="setup_performance";\
	export DEPLOYMENT_NAME="autoscaler-performance";\
	make acceptance-tests-config;\
	make --directory='./src/acceptance' run-acceptance-tests

.PHONY: run-performance
run-performance:
	export NODES=1;\
	export DEPLOYMENT_NAME="autoscaler-performance";\
	export SUITES="run_performance";\
	make acceptance-tests-config;\
    make --directory='./src/acceptance' run-acceptance-tests


.PHONY: run-act
run-act:
	${AUTOSCALER_DIR}/scripts/run_act.sh;\


package-specs: go-mod-tidy go-mod-vendor
	@echo " - Updating the package specs"
	@./scripts/sync-package-specs


## Prometheus Alerts
.PHONY: alerts-silence
alerts-silence:
	export SILENCE_TIME_MINS=480;\
	echo " - Silencing deployment '${DEPLOYMENT_NAME} 8 hours'";\
	${CI_DIR}/autoscaler/scripts/silence_prometheus_alert.sh BOSHJobProcessExtendedUnhealthy ;\
	${CI_DIR}/autoscaler/scripts/silence_prometheus_alert.sh BOSHJobProcessUnhealthy ;\
	${CI_DIR}/autoscaler/scripts/silence_prometheus_alert.sh BOSHJobExtendedUnhealthy ;\
	${CI_DIR}/autoscaler/scripts/silence_prometheus_alert.sh BOSHJobProcessUnhealthy ;\
	${CI_DIR}/autoscaler/scripts/silence_prometheus_alert.sh BOSHJobEphemeralDiskPredictWillFill ;\
	${CI_DIR}/autoscaler/scripts/silence_prometheus_alert.sh BOSHJobUnhealthy ;

.PHONY: docker-login docker docker-image
docker-login: target/docker-login
target/docker-login:
	docker login ghcr.io
	@touch $@
docker-image: docker-login
	docker build -t ghcr.io/cloudfoundry/app-autoscaler-release-tools:latest  ci/dockerfiles/autoscaler-tools
	docker push ghcr.io/cloudfoundry/app-autoscaler-release-tools:latest
validate-openapi-specs: $(wildcard ./api/*.openapi.yaml)
	for file in $^ ; do \
		swagger-cli validate "$${file}" ; \
	done

.PHONY: go-get-u
go-get-u: $(addsuffix .go-get-u,$(go_modules))

.PHONY: %.go-get-u
%.go-get-u: % generate-fakes
	@echo " - go get -u" $<
	cd src/$< && \
	go get -u ./...


deploy-apps:
	echo " - deploying apps"
	DEBUG="${DEBUG}" ${CI_DIR}/autoscaler/scripts/deploy-apps.sh

help: ## Show this help
	@grep --extended-regexp --no-filename '\s##\s' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
