## Migration cleanup


### integration test
- use the gorouter proxy in all integration tests, use example in `src/autoscaler/integration/integration_golangapi_scheduler_test.go`

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
