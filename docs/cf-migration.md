## Migration cleanup


### integration test
- use the gorouter proxy in all integration tests, use example in `src/autoscaler/integration/integration_golangapi_scheduler_test.go`

### Current acceptance tests issue
- when deploy fails running all modules of a dev mtar version, it needs an undeploy before running the acceptance tests again.



### Pending cleanup/improvements after deprecation of bosh deployment

- remove mtls endpoints, remove standalone health server, leave cf server only.
- remove storeprocedure related code.
- split apiserver into 2 services, servicebroker and apiserver
- Re write apply-changelog.sh in java

