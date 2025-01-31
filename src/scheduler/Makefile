# Define the Java formatter command using Devbox
FORMATTER = devbox run google-java-format
JAVA_FILES = $(shell find . -name "*.java" -print | xargs realpath)

.PHONY: format check-format install

# Format all Java files in the src/ directory
format:
	@echo "Formatting Java files..."
	@$(FORMATTER) --skip-javadoc-formatting --replace $(JAVA_FILES)

# Check if files need formatting without modifying them
check-format:
	@echo "Checking Java file formatting..."
	@$(FORMATTER) --skip-javadoc-formatting --dry-run $(JAVA_FILES)

start:
	@echo "Starting the application..."
	@export DB_HOST=${DB_HOST}; \
	mvn spring-boot:run -Dspring.config.location=src/main/resources/application.yml

test:
	@echo "Running tests..."
	@export DB_HOST=${DB_HOST}; \
	mvn test --no-transfer-progress -Dspring.profiles.include=${db_type}
