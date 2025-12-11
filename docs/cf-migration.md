## Migration cleanup


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
