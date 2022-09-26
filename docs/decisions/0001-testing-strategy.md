---
status: proposed
date: 2022-09-26 
deciders: {list everyone involved in the decision}
---
# Testing Strategy

## Context and Problem Statement

We have three layers in our testing pyramid:
* unit tests
* integration tests (tests integration of app-autoscaler microservices)
* acceptance tests (test the app-autoscaler functionality as specified for the end-user)

Apart from that, the acceptance also are used to varying degrees to 
* test the integration with CF
* periodically monitor the functionality of the app-autoscaler (`monitor`)
* as smoke-test validating that a deployment was executed correctly (`validate`)
  * environmental validation: Would also be possible with other means (directly test environment (e.g. user credentials work))
  * operational validation: End-to-end test

## Decision Drivers

* While migrating the acceptance tests from the CF v2 to the CF v3 API, it became clear that the acceptance test scope
  is not well-defined
* When should we run which test
  * During development only
  * After each deployment
  * Continuously
* Our acceptance test suite is too big (slow, maybe unstable due to CF) and by pushing them down to the current layer, we could improve runtime

### Definitions

#### Unit tests

* We test integration with DBs (PostgreSQL, MySQL)
  * Probably wrong, move to integration or test via in-memory DB
* Should be fast, run while you are coding

* Component = class or package
  * Mocks are fakes

#### Integration tests

* Can start up a real DB? (no from a conceptual level, but too highly coupled?)
* Whenever you have mocks
* Mock CF ("assumed contracts")
* Component = Micro-service
  * Mocks are mock-servers (only external services, we start real microservices)

#### System-level integration test

* Tests that App Autoscaler and CF work together in reality
* Component = app-autoscaler-release
  * No mocks

#### Acceptance test

* Validates that service behaves as specified in the user documentation
  * Should be append-only (as removals would imply backwards incompatible changes)
* CF/CLI-triggered (+ REST API CRUD?) operations on scaling policy?
* Minimal set

#### Stress / Performance tests

* Performance/Benchmark: Assert that the product keeps certain timing SLAs (non-functional requirements) (could be without concurrency, but still needs to give significant results)
  * Stress/load test: Behaviour is still fine under stress (high concurrency)
  * With high load/concurrency we validate that it will work on production landscapes

## Considered Options

* Keep all tests in acceptance tests (and rename to `e2e` - for End-to-End tests) and tag them, e.g. `performance`, `smoke`
* Split tests out or move to different layers
* Keep acceptance suite and allow-list some as acceptance tests


## Decision Outcome

Chosen option: "{title of option 1}", because
{justification. e.g., only option, which meets k.o. criterion decision driver | which resolves force {force} | … | comes out best (see below)}.

<!-- This is an optional element. Feel free to remove. -->
### Positive Consequences

* {e.g., improvement of one or more desired qualities, …}
* …

<!-- This is an optional element. Feel free to remove. -->
### Negative Consequences

* {e.g., compromising one or more desired qualities, …}
* …

<!-- This is an optional element. Feel free to remove. -->
## Pros and Cons of the Options

### {title of option 1}

<!-- This is an optional element. Feel free to remove. -->
{example | description | pointer to more information | …}

* Good, because {argument a}
* Good, because {argument b}
<!-- use "neutral" if the given argument weights neither for good nor bad -->
* Neutral, because {argument c}
* Bad, because {argument d}
* … <!-- numbers of pros and cons can vary -->

### {title of other option}

{example | description | pointer to more information | …}

* Good, because {argument a}
* Good, because {argument b}
* Neutral, because {argument c}
* Bad, because {argument d}
* …

<!-- This is an optional element. Feel free to remove. -->
## Validation

{describe how the implementation of/compliance with the ADR is validated. E.g., by a review or an ArchUnit test}

<!-- This is an optional element. Feel free to remove. -->
## More Information

{You might want to provide additional evidence/confidence for the decision outcome here and/or
 document the team agreement on the decision and/or
 define when this decision when and how the decision should be realized and if/when it should be re-visited and/or
 how the decision is validated.
 Links to other decisions and resources might here appear as well.}

<!-- markdownlint-disable-file MD013 -->
