# Developer Guide

## Autoscaler Architecture (WIP)

![Alt text](./images/autoscaler.svg)

### MetricsForwarder

![Alt text](./images/metrics_forwarder.svg)

- Provides an HTTP server to stream app custom metrics to loggregator.
- Authenticate requests via XFCC or BasicAuth.
- Validates received metrics against app policy to check if it is a required metric.
- Manages coolDown threshold for scaling events.

### EventGenerator

![Alt text](./images/eventgenerator.svg)

- Keeps Apps sharded by eventGenerator node.
- Fetches and caches AppPolicy's rules related metrics to evaluate scaling events.
- Evaluates app policies rules and generates scaling events based on metrics cache.
- Manages coolDown threshold for scaling events.

### MetricsServer (To be Deprecated)

![Alt text](./images/metrics_server.svg)

**Responsabilities:**

- For Timer metrics it caches and compiles httpStartStop events to collect a average response time and throughput metric for a configured interval, by default 60 Seconds.
- Keeps track of current metrics sharded by node.
- if persistence is enabled, it stores metrics in DB.
- Provides HTTPServer GET endpoint to retrieve metrics_history by appid/metrictype.
- Transforms GAUGE envelopes into autoscaler compatible metrics (memoryutil, )

### MetricsGateway (To be Deprecated)

![Alt text](./images/metrics_gateway.svg)

