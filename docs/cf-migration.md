## Source code organization 

- src/autoscaler/acceptance to src autoscaler/app-autoscaler/acceptance

## Github actions separation

### To migrate to new repo: 

- WIP - mysql.yaml - Builds and tests with MySQL 8 (test & integration suites)

- resume-ci-vms.yml - Resumes suspended CI VMs weekday mornings (05:00 UTC) and recreates router VM
- suspend-ci-vms.yml - Suspends CI VMs weekday evenings (17:00 UTC) to save costs
- acceptance_tests_mta.yaml - Runs acceptance tests for MTA deployment with additional operations on PRs
- postgres.yaml - Builds and tests with PostgreSQL 15 & 16 (test & integration suites)
- java-ci-lint.yaml - Checks Java code formatting with google-java-format and runs checkstyle
- openapi-specs-check.yaml - Validates OpenAPI specifications on PRs
- acceptance_tests_mta_close.yaml - Cleans up MTA deployment when PR closes
- tidy-go-mod.yaml - Ensures go.mod is tidy on PRs


### To leave in app-autoscaler-release repo:

- codeql-analysis.yml - Runs CodeQL security scanning for Go, Java, and Ruby on pushes/PRs to main and daily
schedule
- acceptance_tests_broker.yaml - Rename to acceptance_tests_bosh. Use submodule instead or real code.
- bosh-templates.yaml - Tests BOSH templates on PRs
- manifest.yaml - Tests manifest generation when templates/jobs/packages change
- bosh-release-checks.yaml - Ensures gosub specs are up-to-date, creates and compiles dev BOSH release
- dependency-updates-post-processing.yaml - Runs go mod tidy and make package-specs after dependency updates

### To remove or refactor.

Merge reausable yaml into broker/bosh workflow and mta.
- acceptance_tests_reusable.yaml - Reusable workflow that deploys autoscaler, runs acceptance test suites. 
- linters.yaml - Runs multiple linters via reviewdog (Go, shellcheck, actionlint, Ruby, alex, markdownlint) - split into 2, bosh related and code related linters.

Image include bosh and all the nix tooling. Ideally we would split the image into 2, one for the release.
And migrate app-autoscaler src specific tools into an specific nix project inside the app-autoscaler project.
- image.yaml - Builds and publishes Docker images to GHCR when dockerfiles or workflow change
- acceptance_tests_broker_close.yaml - Cleans up broker deployment when PR closes


### TODO
--- 
- update-all-golang-dependencies.yaml - Weekly automated update of all Go dependencies (Monday 06:00 UTC)
- renovate_config_validation.yaml - Validates Renovate configuration on PRs

## q
- how does the oss infrastructure would eventually look like? bosh + cf.

will we suppor the boshrelease release generation still in the oss ci ?
do we want to run


## Migration cleanup

### Development workflow
- Deploy single modules. MODULES=apiserver make deploy-apps.

### integration test
- use the gorouter proxy in all integration tests, use example in `src/autoscaler/integration/integration_golangapi_scheduler_test.go`

### Mtar size issue

mtar size is currently at 1.1GB, which is on the larger side for a single mtar.

Split de builder into 3 different ones:

- go apps: `make clean go-mod-vendor-mta` and make sure the scheduler war is not uploaded for modules using this builder:
- java apps: `make clean-scheduler build-scheduler`
- dbtasks:  `make vendor-changelogs clean-dbtasks package-dbtasks`

  - build the apps mtar with all modules except the acceptance tests.
  - build the acceptance tests mtar with only the acceptance tests and the required modules.

### Current acceptance tests issue
- when deploy fails running all modules of a dev mtar version, it needs an undeploy before running the acceptance tests again.


### splitting app-autoscaler-release into two repos
  - move src/autoscaler and src/acceptance test to a new repo.
  - migrate apps acceptance workflow to new repo. Leave bosh acceptance workflow in old repo.

### Code cleanup

- remove mtls endpoints, remove standalone health server, leave cf server only.
- remove storeprocedure related code.
- split apiserver into 2 services, servicebroker and apiserver

### Misc
- Re write apply-changelog.sh in java
