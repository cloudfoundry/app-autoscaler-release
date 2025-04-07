war_file_path := ./target/scheduler-${scheduler_version}.war
.PHONY: build
build: ${war_file_path}

scheduler_version := $(shell xq --xpath='/project/version' ./pom.xml)
java_files = $(shell find . -type f -name "*.java" -print | xargs realpath)
MVN_BUILD_OPTS ?= '-Dmaven.test.skip=true'
${war_file_path}: $(java_files) ./pom.xml ./Makefile
	@echo "# building $@ (root in $$(pwd))"
	@mvn --no-transfer-progress package '${MVN_BUILD_OPTS}'



# Format all Java files in the src/ directory
.PHONY: format
format:
	@echo "Formatting Java files..."
	@google-java-format --skip-javadoc-formatting --replace $(java_files)

# Check if files need formatting without modifying them
.PHONY: check-format
check-format:
	@echo "Checking Java file formatting..."
	@google-java-format --skip-javadoc-formatting --dry-run $(java_files)



.PHONY: start
start:
	@echo "Starting the application..."
	@export DB_HOST=${DB_HOST}; \
	mvn spring-boot:run '-Dspring.config.location=./src/main/resources/application.yml'



.PHONY: test
test: test-certificates
	@echo 'Running tests …' \
		'\n\t🚸 Please keep in mind that this requires '
	@export DB_HOST='${DB_HOST}'; \
	mvn test --no-transfer-progress '-Dspring.profiles.include=${db_type}'

test-certs-dir := src/test/resources/certs
test-certs = $(shell find './${test-certs-dir}' -type f 2> /dev/null)
.PHONY: test-certificates
test-certificates: ${test-certs-dir}
${test-certs-dir} ${test-certs} &:
	@mkdir --parents '${test-certs-dir}'
	@./scripts/generate_unit_test_certs.sh

clean:
	rm --force --recursive \
		'./target' \
		'${test-certs-dir}'
