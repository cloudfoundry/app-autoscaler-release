SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c ${SHELLFLAGS}
MAKEFLAGS = -s

acceptance-dir := ./src/acceptance
autoscaler-dir := ./src/autoscaler
changelog-dir := ./src/changelog
changeloglockcleaner-dir := ./src/changeloglockcleaner
db-dir := ./src/db
scheduler-dir := ./src/scheduler
test-app-dir := ${acceptance-dir}/assets/app/go_app

# 🚧 To-do: Remove me!
go_modules := $(shell find . -maxdepth 6 -name "*.mod" -exec dirname {} \; | sed 's|\./src/||' | sort)


OS := $(shell . /etc/lsb-release &>/dev/null && echo $${DISTRIB_ID} || uname)
db_type := postgres
DB_HOST := localhost
DBURL := $(shell case "${db_type}" in\
			 (postgres) printf "postgres://postgres:postgres@${DB_HOST}/autoscaler?sslmode=disable"; ;; \
				 (mysql) printf "root@tcp(${DB_HOST})/autoscaler?tls=false"; ;; esac)
DEBUG := false
MYSQL_TAG := 8
POSTGRES_TAG := 16
SUITES ?= broker api app
AUTOSCALER_DIR ?= $(shell pwd)
lint_config := ${AUTOSCALER_DIR}/.golangci.yaml
CI_DIR ?= ${AUTOSCALER_DIR}/ci
CI ?= false
VERSION ?= 0.0.testing
DEST ?= build

export DEBUG ?= false
export ACCEPTANCE_TESTS_FILE ?= ${DEST}/app-autoscaler-acceptance-tests-v${VERSION}.tgz
export GOWORK = off

$(shell mkdir -p target)
$(shell mkdir -p build)

.DEFAULT_GOAL := build_all

.PHONY: check-type
check-db_type:
	@case "${db_type}" in\
	 (mysql|postgres) echo " - using db_type:${db_type}"; ;;\
	 (*) echo "ERROR: db_type needs to be one of mysql|postgres"; exit 1;;\
	 esac

.PHONY: init-db
init-db: check-db_type start-db db.java-libs target/init-db-${db_type}
target/init-db-${db_type}:
	@./scripts/initialise_db.sh '${db_type}'
	@touch $@

# 🚧 To-do: Substitute me by a definition that calls the Makefile-targets of the other Makefiles!
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
	@make --directory='${autoscaler-dir}' clean
clean-scheduler:
	@echo " - cleaning scheduler test resources"
	@rm -rf src/scheduler/src/test/resources/certs
	@rm -rf src/scheduler/target
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



.PHONY: build_all build_programs build_tests
build_all: build_programs build_tests
build_programs: autoscaler.build db.java-libs scheduler.build build-test-app
build_tests:acceptance.build_tests autoscaler.build_tests changelog.build_tests changeloglockcleaner.build_tests

.PHONY: acceptance.build_tests
acceptance.build_tests:
	@make --directory='${acceptance-dir}' build_tests

.PHONY: autoscaler.build
autoscaler.build:
	@make --directory='${autoscaler-dir}' build

.PHONY: autoscaler.build_tests
autoscaler.build_tests:
	@make --directory='${autoscaler-dir}' build_tests

.PHONY: changelog.build
changelog.build:
	@make --directory='${changelog-dir}' build

.PHONY: changelog.build_tests
changelog.build_tests:
	@make --directory='${changelog-dir}' build_tests

.PHONY: changeloglockcleaner.build
changeloglockcleaner.build:
	@make --directory='${changeloglockcleaner-dir}' build

.PHONY: changeloglockcleaner.build_tests
changeloglockcleaner.build_tests:
	@make --directory='${changeloglockcleaner-dir}' build_tests

MVN_OPTS ?= -Dmaven.test.skip=true
db.java-lib-dir := src/db/target/lib
db.java-lib-files = $(shell find '${db.java-lib-dir}' -type f -name '*.jar' 2> /dev/null)
.PHONY: db.java-libs
db.java-libs: ${db.java-lib-dir} ${db.java-lib-files}
${db.java-lib-dir} ${db.java-lib-files} &: src/db/pom.xml
	@mkdir --parents '${db.java-lib-dir}'
	@echo 'Fetching db.java-libs'
	@pushd src &> /dev/null \
		&& mvn --no-transfer-progress package --projects='db' ${MVN_OPTS} \
	&& popd

.PHONY:
scheduler.build:
	@make --directory='${scheduler-dir}' build

# 🚧 To-do: Substitute me by definitions that call the Makefile-targets of the other Makefiles!
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
test: test-autoscaler scheduler.test test-changelog test-changeloglockcleaner test-acceptance-unit ## Run all unit tests
test-autoscaler: check-db_type init-db test-certs
	@echo ' - using DBURL=${DBURL} TEST=${TEST}'
	@make --directory='${autoscaler-dir}' test DBURL='${DBURL}' TEST='${TEST}'
test-autoscaler-suite: check-db_type init-db test-certs
	@make --directory='${autoscaler-dir}' testsuite TEST='${TEST}' DBURL='${DBURL}' GINKGO_OPTS='${GINKGO_OPTS}'

test-changelog:
	@make --directory='${changelog-dir}' test
test-changeloglockcleaner: init-db test-certs
	@make --directory='${changeloglockcleaner-dir}' test DBURL='${DBURL}'
test-acceptance-unit:
	@make --directory='${acceptance-dir}' test-unit
	@make --directory='${test-app-dir}' test

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
			postgres:${POSTGRES_TAG} \
			-c 'max_connections=1000' >/dev/null;\
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
	@which mysql > /dev/null &&\
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
integration: init-db test-certs build_all build-gorouterproxy
	@echo " - using DBURL=${DBURL}"
	@make --directory='${autoscaler-dir}' integration DBURL="${DBURL}"


.PHONY: lint lint-go acceptance.lint autoscaler.lint test-app.lint changelog.lint changeloglockcleaner.lint
lint: lint-go lint-ruby lint-actions lint-markdown lint-gorouterproxy
lint-go: acceptance.lint autoscaler.lint test-app.lint changelog.lint changeloglockcleaner.lint
acceptance.lint:
	@echo 'Linting acceptance-tests …'
	make --directory='${acceptance-dir}' lint
autoscaler.lint:
	@echo 'Linting autoscaler …'
	make --directory='${autoscaler-dir}' lint
# ⚠️ Not existing: scheduler.lint:
test-app.lint:
	@echo 'Linting test-app …'
	make --directory='${test-app-dir}' lint
changelog.lint:
	@echo 'Linting changelog …'
	make --directory='${changelog-dir}' lint
changeloglockcleaner.lint:
	@echo 'Linting changeloglockcleaner …'
	make --directory='${changeloglockcleaner-dir}' lint

lint-ruby:
	@echo " - ruby scripts"
	@bundle install
	@bundle exec standardrb ${RUBOCOP_OPTS} ./spec ./packages

.PHONY: lint-markdown
lint-markdown:
	@echo " - linting markdown files"
	@markdownlint-cli2 .

.PHONY: lint-actions
lint-actions:
	@echo " - linting GitHub actions"
	actionlint

lint-gorouterproxy:
	@echo " - linting: gorouterproxy"
	@pushd src/autoscaler/integration/gorouterproxy >/dev/null && golangci-lint run --config='${lint_config}' $(OPTS)

.PHONY: spec-test
spec-test:
	bundle install
	bundle exec rspec

.PHONY: bosh-release
bosh-release: build/autoscaler-test.tgz_CI_${CI}
# 🚸 In the next line, the order of the dependencies is important. Generated code needs to be
# already there for `go-mod-tidy` to work. See additional comment for that target in
# ./src/autoscaler/Makefile.
build/autoscaler-test.tgz_CI_false: build_all go-mod-tidy go-mod-vendor
	@echo " - creating bosh release into build/autoscaler-test.tgz"; \
	@bosh create-release --force --timestamp-version --tarball=build/autoscaler-test.tgz; \

build/autoscaler-test.tgz_CI_true: build_all go-mod-tidy go-mod-vendor
	@echo " - creating bosh release into ${AUTOSCALER_BOSH_TARBALL_PATH}"
	@bosh create-release ${AUTOSCALER_BOSH_BUILD_OPTS} --version ${AUTOSCALER_BOSH_VERSION} --tarball=${AUTOSCALER_BOSH_TARBALL_PATH}

.PHONY: generate-fakes autoscaler.generate-fakes test-app.generate-fakes
generate-fakes: autoscaler.generate-fakes test-app.generate-fakes
autoscaler.generate-fakes:
	make --directory='${autoscaler-dir}' generate-fakes
test-app.generate-fakes:
	make --directory='${test-app-dir}' generate-fakes

.PHONY: generate-openapi-generated-clients-and-servers
generate-openapi-generated-clients-and-servers:
	make --directory='${autoscaler-dir}' generate-openapi-generated-clients-and-servers


 .PHONY: go-mod-tidy
go-mod-tidy: acceptance.go-mod-tidy autoscaler.go-mod-tidy changelog.go-mod-tidy \
						 changeloglockcleaner.go-mod-tidy test-app.go-mod-tidy


.PHONY: autoscaler.go-mod-tidy
autoscaler.go-mod-tidy: acceptance.go-mod-tidy autoscaler.go-mod-tidy changelog.go-mod-tidy \
						 changeloglockcleaner.go-mod-tidy


.PHONY: acceptance.go-mod-tidy autoscaler.go-mod-tidy changelog.go-mod-tidy \
				changeloglockcleaner.go-mod-tidy test-app.go-mod-tidy
acceptance.go-mod-tidy:
	make --directory='${acceptance-dir}' go-mod-tidy
autoscaler.go-mod-tidy:
	make --directory='${autoscaler-dir}' go-mod-tidy
changelog.go-mod-tidy:
	make --directory='${changelog-dir}' go-mod-tidy
changeloglockcleaner.go-mod-tidy:
	make --directory='${changeloglockcleaner-dir}' go-mod-tidy
test-app.go-mod-tidy:
	make --directory='${test-app-dir}' go-mod-tidy


.PHONY: mod-download
mod-download:
	@for folder in $$(find . -maxdepth 3 -name "go.mod" -exec dirname {} \;);\
	do\
		 cd $${folder}; echo " - go mod download '$${folder}'"; go mod download; cd - >/dev/null;\
	done

.PHONY: acceptance.go-mod-vendor autoscaler.go-mod-vendor changelog.go-mod-vendor \
				changeloglockcleaner.go-mod-vendor
go-mod-vendor: clean-vendor acceptance.go-mod-vendor autoscaler.go-mod-vendor changelog.go-mod-vendor \
							 changeloglockcleaner.go-mod-vendor
acceptance.go-mod-vendor:
	make --directory='${acceptance-dir}' go-mod-vendor
autoscaler.go-mod-vendor:
	make --directory='${autoscaler-dir}' go-mod-vendor
changelog.go-mod-vendor:
	make --directory='${changelog-dir}' go-mod-vendor
changeloglockcleaner.go-mod-vendor:
	make --directory='${changeloglockcleaner-dir}' go-mod-vendor


.PHONY: update-uaac-nix-package
update-uaac-nix-package:
	make --directory='./nix/packages/uaac' gemset.nix

.PHONY: deploy-autoscaler deploy-register-cf deploy-autoscaler-bosh deploy-cleanup
deploy-autoscaler: deploy-autoscaler-bosh
deploy-register-cf:
	echo " - registering broker with cf"
	${CI_DIR}/autoscaler/scripts/register-broker.sh

deploy-autoscaler-bosh: db.java-libs go-mod-vendor scheduler.build
	echo " - deploying autoscaler"
	DEBUG="${DEBUG}" ${CI_DIR}/autoscaler/scripts/deploy-autoscaler.sh
deploy-cleanup:
	${CI_DIR}/autoscaler/scripts/cleanup-autoscaler.sh

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
	@make --directory='${autoscaler-dir}' mta-build

.PHONY: build-test-app
build-test-app:
	@make --directory='${test-app-dir}' build

build-gorouterproxy:
	@make --directory='${autoscaler-dir}' build-gorouterproxy

.PHONY: deploy-test-app
deploy-test-app:
	@make --directory='${test-app-dir}' deploy

.PHONY: build-acceptance-tests
build-acceptance-tests:
	@make --directory='${acceptance-dir}' build_tests

.PHONY: acceptance-tests
acceptance-tests: build-test-app acceptance-tests-config ## Run acceptance tests against OSS dev environment (requrires a previous deployment of the autoscaler)
	@make --directory='${acceptance-dir}' run-acceptance-tests
.PHONY: acceptance-cleanup
acceptance-cleanup:
	@make --directory='${acceptance-dir}' acceptance-tests-cleanup
.PHONY: acceptance-tests-config
acceptance-tests-config:
	make --directory='${acceptance-dir}' acceptance-tests-config

.PHONY: cleanup-concourse
cleanup-concourse:
	@${CI_DIR}/autoscaler/scripts/cleanup-concourse.sh

.PHONY: cleanup-autoscaler-deployments
cleanup-autoscaler-deployments:
	@${CI_DIR}/autoscaler/scripts/cleanup-autoscaler-deployments.sh

.PHONY: cf-login
cf-login:
	@echo '⚠️ Please note that this login only works for cf and concourse,' \
		  'in spite of performing a login as well on bosh and credhub.' \
		  'The necessary changes to the environment get lost when make exits its process.'
	@${CI_DIR}/autoscaler/scripts/os-infrastructure-login.sh

.PHONY: setup-performance
setup-performance: build-test-app
	export NODES=1;\
	export SUITES="setup_performance";\
	export DEPLOYMENT_NAME="autoscaler-performance";\
	make acceptance-tests-config;\
	make --directory='${acceptance-dir}' run-acceptance-tests

.PHONY: run-performance
run-performance:
	export NODES=1;\
	export DEPLOYMENT_NAME="autoscaler-performance";\
	export SUITES="run_performance";\
	make acceptance-tests-config;\
	make --directory='${acceptance-dir}' run-acceptance-tests


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
		redocly lint --extends=minimal --format=$(if $(GITHUB_ACTIONS),github-actions,codeframe) "$${file}" ; \
	done


# 🚧 To-do: Substitute me by a definition that calls the Makefile-targets of the other Makefiles!
.PHONY: go-get-u
go-get-u: $(addsuffix .go-get-u,$(go_modules))
# 🚧 s.o
.PHONY: %.go-get-u
%.go-get-u: % generate-fakes
	@echo " - go get -u" $<
	cd src/$< && \
	go get -u ./...


# This target is defined here rather than directly in the component “scheduler” itself, because it depends on targets outside that component. In the future, it will be moved back to that component and reference a dependency to a Makefile on the same level – the one for the component it depends on.
.PHONY: start-scheduler scheduler.start
start-scheduler: scheduler.start
scheduler.start: check-db_type init-db
	pushd '${scheduler-dir}'; \
		echo "Starting the application in $(pwd) …"; \
		export DB_HOST='${DB_HOST}'; \
		mvn spring-boot:run \
			'-Dspring.config.location=./src/main/resources/application.yml'; \
	popd

# This target is defined here rather than directly in the component “scheduler” itself, because it depends on targets outside that component. In the future, it will be moved back to that component and reference a dependency to a Makefile on the same level – the one for the component it depends on.
.PHONY: scheduler.test
scheduler.test: check-db_type scheduler.test-certificates init-db
	pushd '${scheduler-dir}'; \
		echo "Running tests in $(pwd) …"; \
		export DB_HOST='${DB_HOST}'; \
		mvn test \
			--no-transfer-progress '-Dspring.profiles.include=${db_type}'; \
	popd

.PHONY: scheduler.test-certificates
scheduler.test-certificates:
	make --directory='${scheduler-dir}' test-certificates



list-apps:
	echo " - listing apps"
	DEBUG="${DEBUG}" ${CI_DIR}/../scripts/list_apps.sh

deploy-apps:
	echo " - deploying apps"
	DEBUG="${DEBUG}" ${CI_DIR}/autoscaler/scripts/deploy-apps.sh

undeploy-apps:
	echo " - undeploying apps"
	DEBUG="${DEBUG}" ${CI_DIR}/autoscaler/scripts/undeploy-apps.sh
help: ## Show this help
	@grep --extended-regexp --no-filename '\s##\s' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
