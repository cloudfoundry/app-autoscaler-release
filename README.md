# Bosh Release for app-autoscaler service

(This release is under active development)

## Purpose

The purpose of this bosh release is to deploy and setup the [app-autoscaler](https://github.com/cloudfoundry-incubator/app-autoscaler) service.

## Usage

### Bosh Lite Deployment

Install [Bosh-cli-v2](https://bosh.io/docs/cli-v2.html#install)

Install and start [BOSH-Deployment](https://github.com/cloudfoundry/bosh-deployment), following its [README](https://github.com/cloudfoundry/bosh-deployment/blob/master/README.md).

Install [CF-deployment](https://github.com/cloudfoundry/cf-deployment/blob/master/cf-deployment.yml)

Install the UAA CLI, `uaac`.

```sh
gem install cf-uaac
```

Obtain `uaa_admin_client_secret`

```sh
bosh interpolate --path /uaa_admin_client_secret /path/to/cf-deployment/deployment-vars.yml
```

Use the `uaac target uaa.YOUR-DOMAIN` command to target your UAA server and obtain an access token for the admin client.

```sh
 uaac target uaa.bosh-lite.com --skip-ssl-validation
 uaac token client get admin -s <uaa_admin_client_secret>
 ```

Create a new autoscaler client

```sh
uaac client add "autoscaler_client_id" \
    --authorized_grant_types "client_credentials" \
    --authorities "cloud_controller.read,cloud_controller.admin,uaa.resource" \
    --secret "autoscaler_client_secret"
```

Create and upload release

```sh
git clone https://github.com/cloudfoundry-incubator/app-autoscaler-release
cd app-autoscaler-release
./scripts/update
bosh create-release
bosh -e YOUR_ENV upload-release
```

Deploy app-autoscaler with the newly created autoscaler client

```sh
bosh -e YOUR_ENV -d app-autoscaler \
     deploy templates/app-autoscaler-deployment.yml \
     --vars-store=bosh-lite/deployments/vars/autoscaler-deployment-vars.yml \
     -v system_domain=bosh-lite.com \
     -v cf_client_id=autoscaler_client_id \
     -v cf_client_secret=autoscaler_client_secret \
     -v skip_ssl_validation=true
```
To deploy app-autoscaler with density, use `app-autoscaler-deployment-fewer.yml`
```sh
bosh -e YOUR_ENV -d app-autoscaler \
     deploy templates/app-autoscaler-deployment-fewer.yml \
     --vars-store=bosh-lite/deployments/vars/autoscaler-deployment-vars.yml \
     -v system_domain=bosh-lite.com \
     -v cf_client_id=autoscaler_client_id \
     -v cf_client_secret=autoscaler_client_secret \
     -v skip_ssl_validation=true

```
Alternatively you can use cf-deployment vars file to provide the cf_client_id and cf_client_secret

```sh
bosh -e YOUR_ENV -d app-autoscaler \
     deploy templates/app-autoscaler-deployment.yml \
     --vars-store=bosh-lite/deployments/vars/autoscaler-deployment-vars.yml \
     -v system_domain=bosh-lite.com \
     -v skip_ssl_validation=true \
     --vars-file=<path to cf deployment vars file>
```

#### Deploy autoscaler with external postgres database

```sh
bosh -e YOUR_ENV -d app-autoscaler \
     deploy templates/app-autoscaler-deployment.yml \
     --vars-store=bosh-lite/deployments/vars/autoscaler-deployment-vars.yml \
     -v system_domain=bosh-lite.com \
     -v cf_client_id=autoscaler_client_id \
     -v cf_client_secret=autoscaler_client_secret \
     -v skip_ssl_validation=true \
     -v database_host=<database_host> \
     -v database_port=<database_port> \
     -v database_username=<database_username> \
     -v database_password=<database_password> \
     -v database_name=<database_name> \
     -o example/operation/external-db.yml
```

#### Deploy autoscaler with bosh-dns instead of consul for service registration

```sh
bosh -e YOUR_ENV -d app-autoscaler \
     deploy templates/app-autoscaler-deployment.yml \
     --vars-store=bosh-lite/deployments/vars/autoscaler-deployment-vars.yml \
     -o example/operation/bosh-dns.yml \
     -v system_domain=bosh-lite.com \
     -v cf_client_id=autoscaler_client_id \
     -v cf_client_secret=autoscaler_client_secret \
     -v skip_ssl_validation=true
```
For density deployment
```sh
bosh -e YOUR_ENV -d app-autoscaler \
     deploy templates/app-autoscaler-deployment-fewer.yml \
     --vars-store=bosh-lite/deployments/vars/autoscaler-deployment-vars.yml \
     -o example/operation/bosh-dns-fewer.yml \
     -v system_domain=bosh-lite.com \
     -v cf_client_id=autoscaler_client_id \
     -v cf_client_secret=autoscaler_client_secret \
     -v skip_ssl_validation=true
```
#### Deploy autoscaler with postgres database enabled TLS

```sh
bosh -e YOUR_ENV -d app-autoscaler \
     deploy templates/app-autoscaler-deployment.yml \
     --vars-store=bosh-lite/deployments/vars/autoscaler-deployment-vars.yml \
     -o example/operation/postgres-ssl.yml \
     -v system_domain=bosh-lite.com \
     -v cf_client_id=autoscaler_client_id \
     -v cf_client_secret=autoscaler_client_secret \
     -v skip_ssl_validation=true
```
For density deployment
```sh
bosh -e YOUR_ENV -d app-autoscaler \
     deploy templates/app-autoscaler-deployment-fewer.yml \
     --vars-store=bosh-lite/deployments/vars/autoscaler-deployment-vars.yml \
     -o example/operation/postgres-ssl-fewer.yml \
     -v system_domain=bosh-lite.com \
     -v cf_client_id=autoscaler_client_id \
     -v cf_client_secret=autoscaler_client_secret \
     -v skip_ssl_validation=true
```
>** It's advised not to make skip_ssl_validation=true for non-development environment

## Register service

Log in to Cloud Foundry with admin user, and use the following commands to register `app-autoscaler` service

```sh
cf create-service-broker autoscaler <brokerUserName> <brokerPassword> <brokerURL>
```

* `brokerUserName`: the user name to authenticate with service broker. It's default value is `autoscaler_service_broker_user`.
* `brokerPassword`: the password to authenticate with service broker. It will be stored in the file passed to the --vars-store flag (bosh-lite/deployments/vars/autoscaler-deployment-vars.yml in the example). You can find them by searching for `autoscaler_service_broker_password`.
* `brokerURL`: the URL of the service broker

All these parameters are configured in the bosh deployment. If you are using default values of deployment manifest, register the service with the commands below.

```sh
cf create-service-broker autoscaler autoscaler_service_broker_user `bosh int ./bosh-lite/deployments/vars/autoscaler-deployment-vars.yml --path /autoscaler_service_broker_password` https://autoscalerservicebroker.bosh-lite.com
```

## Acceptance test

Refer to [AutoScaler UAT guide](src/acceptance/README.md) to run acceptance test. 

## Use service

To use the service to auto-scale your applications, log in to Cloud Foundry with admin user, and use the following command to enable service access to all or specific orgs.

```sh
cf enable-service-access autoscaler [-o ORG]
```

The following commands don't require admin rights, but user needs to be Space Developer. Create the service instance, and then bind your application to the service instance with the policy as parameter.

```sh
cf create-service autoscaler  autoscaler-free-plan  <service_instance_name>
cf bind-service <app_name> <service_instance_name> -c <policy>
```

## Remove the service

Log in to Cloud Foundry with admin user, and use the following commands to remove all the service instances and the service broker of `app-autoscaler` from Cloud Foundry.

```sh
cf purge-service-offering autoscaler
cf delete-service-broker autoscaler
```
