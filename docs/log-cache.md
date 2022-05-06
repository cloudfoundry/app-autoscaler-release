# Log Cache and Autoscaler

Log Cache allows you to view logs and metrics over a specified period of time. It could work as a replacement for 
MetricsGateway and MetricsServer components. See [developers docs](developer_guide.md).

Autoscaler uses 2 main group of metrics, TIMER and GAUGE. it is theoretically possible for autoscaler to consume 
metrics via log-cache instead of firehose.

__TIMER__

- httpsStartStop.start
- httpsStartStop.stop

__GAUGE__

- cpu
- diks
- disk_quota
- memory
- memory_quota
- "custom_metric"

## Autoscaler metrics

Autoscaler turns Loggregator instance metrics into Autoscaler app metrics, these calculations currently happen in 
MetricsServer's envelop_processor.

### Memory and CPU 

**memoryutil** : int(math.Ceil(memory/memory_quota*100)) %
**memoryused** : int(math.Ceil(memory/(1024*1024))) MB
**cpu** : int64(math.Ceil(cpu)) %

### Throughput and ResponseTime

"throughput" : Count(httsStartStop metrics for the last x seconds)
"responsetime" : sum(stop-start)/numReq

If no timer metrics were received during the collection period for a given app, one evenlop_processor
per app will generate 2 metric and stream them to the pipe:

- 1 throughput  metrics with instance_index: 0, value: 0
- 1 responsetime metrics with instance_index: 0, value: 0

## Retrieve data from log-cache

```
curl --header "Authorization: $(cf oauth-token)" \
    "https://log-cache.((SYSTEM_DOMAIN))/api/v1/read/((APP_ID))?envelope_types=[GAUGE|TIMER]&start_time=$(($(date --utc '+%s%N') - SECONDS * 1000000000))" 
```