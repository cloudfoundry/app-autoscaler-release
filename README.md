# Application Autoscaler

The Application Autoscaler provides the capability to adjust the computation resources for Cloud Foundry applications
through

* dynamic scaling based on application performance metrics
* scheduled scaling based on time

## Local Development

### Prerequisites

* [Docker](https://www.docker.com/products/docker-desktop/) to spin up the required databases
* [devbox](https://github.com/jetify-com/devbox) to start a shell with all required tools
* clone of https://github.com/cloudfoundry/app-autoscaler-env-bbl-state next to the clone of this repo to access systems
* [direnv](https://direnv.net/) to spin up the shell properly before running the make-targets

### Make Targets

| Category          | Description                                                                            | Target                                                                                                                                                    |
|-------------------|----------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|
| mock              | generate mocks                                                                         | `make generate-fakes`                                                                                                                                     |
| unit-tests        | run against PostgreSQL                                                                 | `make test`                                                                                                                                               |
| unit-tests        | run against specific PostgreSQL version                                                | <pre><code>make clean #Only if you're changing versions to refresh the running docker image<br/>make test POSTGRES_TAG=x.y</code></pre>                   |
| unit-tests        | run against MySQL                                                                      | `make test db_type=mysql`                                                                                                                                 |
| unit-tests        | run against specific MySQL version                                                     | <pre><code>make clean #Only if you're changing versions to refresh the running docker image<br/>make test db_type=mysql MYSQL_TAG=x.y</code></pre>        |
| integration-tests | run against PostgreSQL                                                                 | `make integration`                                                                                                                                        |
| integration-tests | run against specific PostgreSQL version                                                | <pre><code>make clean #Only if you're changing versions to refresh the running docker image<br/>make integration POSTGRES_TAG=x.y</code></pre>            |
| integration-tests | run against MySQL                                                                      | `make integration db_type=mysql`                                                                                                                          |
| integration-tests | run against specific MySQL version                                                     | <pre><code>make clean #Only if you're changing versions to refresh the running docker image<br/>make integration db_type=mysql MYSQL_TAG=x.y</code></pre> |
| acceptance-tests  | run acceptance-tests, see [AutoScaler UAT guide](src/acceptance/README.md) for details | `make acceptance-tests`                                                                                                                                   |
| lint              | check code style                                                                       | `make lint`                                                                                                                                               |
| lint              | check code style and apply auto-fixes                                                  | `OPTS=--fix RUBOCOP_OPTS=-A make lint`                                                                                                                    |
| build             | compile project                                                                        | `make build`                                                                                                                                              |
| deploy            | deploy Application Autoscaler and register the service broker in CF                    | `make deploy-autoscaler`                                                                                                                                  |
| cleanup           | remove build artifacts                                                                 | `make clean`                                                                                                                                              |

## Use Application Autoscaler Service

Refer to [user guide](docs/Readme.md) for the details of how to use the Auto-Scaler service, including policy
definition, supported metrics, public API specification and command line tool.

## Monitor Microservices

The app-autoscaler provides a number of health endpoints that are available externally that can be used to check the
state of each component. Each health endpoint is protected with basic auth (apart from the api server), the usernames
are listed in the table below, but the passwords are available in credhub.

| Component        | Health URL                                                   | Username         | Password Key                                 |
|------------------|--------------------------------------------------------------|------------------|----------------------------------------------|
| eventgenerator   | https://autoscaler-eventgenerator.((system_domain))/health   | eventgenerator   | /autoscaler_eventgenerator_health_password   |
| metricsforwarder | https://autoscaler-metricsforwarder.((system_domain))/health | metricsforwarder | /autoscaler_metricsforwarder_health_password |
| scalingengine    | https://autoscaler-scalingengine.((system_domain))/health    | scalingengine    | /autoscaler_scalingengine_health_password    |
| operator         | https://autoscaler-operator.((system_domain))/health         | operator         | /autoscaler_operator_health_password         |
| scheduler        | https://autoscaler-scheduler.((system_domain))/health        | scheduler        | /autoscaler_scheduler_health_password        |

These endpoints can be disabled by using the ops
file [`example/operations/disable-basicauth-on-health-endpoints.yml`](operations/disable-basicauth-on-health-endpoints.yml)

## License

This project is released under version 2.0 of the [Apache License](LICENSE).
