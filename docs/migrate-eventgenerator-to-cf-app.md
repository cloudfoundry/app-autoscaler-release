# Run Eventgenerator as CF App

==== WORK IN PROGRESS====
## Motivation

Compared with apps deployed on BOSH-provisioned VMs the CF Runtime restores failing app instances (due to infrastructure issues) way faster and is a perfect target for [12-factor](https://12factor.net) apps.

The eventgenerator is a 12-factor app, but also splits the work to be done between app instances, which means that missing app instances degrade the service as certain apps won't be scaled, so a runtime that can quickly restart failing app instances in case of hardware failures is key to minimize disruptions.


## Eventgenerator on a Bosh VM

Autoscaler's eventgenerator components is used to collect raw metrics from log-cache and process them in to application metrics internally

Following diagram shows the interaction between external components e.g., databases, services

```mermaid
flowchart TB
    subgraph Autoscaler Bosh Deployment
        subgraph Eventgenerator-VM
           https://acceptance-lc.eventgenerator.service.cf.internal:6105
           aggregated_history
           healthserver:6205
        end
        Eventgenerator-VM--datasource--- appMetrics_db[(app_metrics_db\nacceptance-lc.autoscalerpostgres.service.cf.internal:5432/autoscaler)]
        Eventgenerator-VM --datasource--- policyDb[(policybosh_db\nacceptance-lc.autoscalerpostgres.service.cf.internal:5432/autoscaler)]  
        Eventgenerator-VM--exposed\nvia bosh-dns---ScalingEngine-VM[[ScalingEngine\nhttps://acceptance-lc.scalingengine.service.cf.internal:6104]]
    end
    subgraph CF Deployment
        Logcache-cf-auth-proxy[[log-cache-cf-auth-proxy:8083]]----> |collects raw logs| Eventgenerator-VM
        Logcache-VM-->Logcache-cf-auth-proxy[[log-cache-cf-auth-proxy:8083]]
    end
```
## Eventgenerator as CF Application

### Tasks
 - Run Eventgenerator service as CF app
 - Communicate with scalingengine via exposed route_registrar route
 - Communicate with logcache via uaa authentication
```mermaid
flowchart LR
    subgraph CF Bosh Deployment
        Logcache-cf-auth-proxy[[log-cache-cf-auth-proxy:8083]]
        Logcache-cf-auth-proxy[[log-cache-cf-auth-proxy:8083]]---Logcache-VM
    end
    subgraph CF Runtime
        Eventgenerator-App--via UAA authentication-->Logcache-cf-auth-proxy
    end
    subgraph Autoscaler Bosh Deployment
        Eventgenerator-App--datasource--- appMetrics_db[(app_metrics_db\nacceptance-lc.autoscalerpostgres.service.cf.internal:5432/autoscaler)]
        Eventgenerator-App --datasource--- policyDb[(policybosh_db\nacceptance-lc.autoscalerpostgres.service.cf.internal:5432/autoscaler)]  
        Eventgenerator-App--exposed\nvia route registrar---ScalingEngine-VM[[ScalingEngine\nhttps://acceptance-lc.scalingengine.service.cf.internal:6104]]
    end
```