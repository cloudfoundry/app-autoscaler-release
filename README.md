# Bosh Release for app-autoscaler service

(This release is under active development)

## Purpose

The purpose of this bosh release is to deploy and setup the [app-autoscaler](https://github.com/cloudfoundry-incubator/app-autoscaler) service.

## Usage

### Bosh Lite Deployment

Install [Bosh-cli-v2](https://bosh.io/docs/cli-v2.html#install)

Install and start [BOSH-Deployment](https://github.com/cloudfoundry/bosh-deployment), following its [README](https://github.com/cloudfoundry/bosh-deployment/blob/master/README.md).

Install [CF-deployment](https://github.com/cloudfoundry/cf-deployment/blob/master/cf-deployment.yml)

Create and upload release

```sh
git clone https://github.com/cloudfoundry-incubator/app-autoscaler-release
cd app-autoscaler-release
./scripts/update
bosh create-release
bosh -e YOUR_ENV upload-release
```

Deploy app-autoscaler

```sh
bosh -e YOUR_ENV -d app-autoscaler \
     deploy templates/app-autoscaler-deployment.yml \
     --vars-store=bosh-lite/deployments/vars/autoscaler-deployment-vars.yml \
     -v system_domain=bosh-lite.com \
     -v cf_admin_password=<cf admin password> \
     -v skip_ssl_validation=true
```

Alternatively you can use cf-deployment vars file to provide the cf_admin_password

```sh
bosh -e YOUR_ENV -d app-autoscaler \
     deploy templates/app-autoscaler-deployment.yml \
     --vars-store=bosh-lite/deployments/vars/autoscaler-deployment-vars.yml \
     -v system_domain=bosh-lite.com \
     -v skip_ssl_validation=true \
     --vars-file=<path to cf deployment vars file>
```

#### Deploy autoscaler with `client_credentials` flow

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
    --authorities "cloud_controller.read,cloud_controller.admin" \
    --secret "autoscaler_client_secret"
```

Deploy autoscaler with the newly created autoscaler client

```sh
bosh -e YOUR_ENV -d app-autoscaler \
     deploy templates/app-autoscaler-deployment.yml \
     --vars-store=bosh-lite/deployments/vars/autoscaler-deployment-vars.yml \
     -v system_domain=bosh-lite.com \
     -v autoscaler_client_id=autoscaler_client_id \
     -v autoscaler_client_secret=autoscaler_client_secret \
     -v skip_ssl_validation=true \
     -o example/operation/client-credentials.yml
```

#### Deploy autoscaler with external postgres database

```sh
bosh -e YOUR_ENV -d app-autoscaler \
     deploy templates/app-autoscaler-deployment.yml \
     --vars-store=bosh-lite/deployments/vars/autoscaler-deployment-vars.yml \
     -v system_domain=bosh-lite.com \
     -v cf_admin_password=<cf admin password> \
     -v skip_ssl_validation=true \
     -v database_host=<database_host> \
     -v database_port=<database_port> \
     -v database_username=<database_username> \
     -v database_password=<database_password> \
     -v database_name=<database_name> \
     -o example/operation/external-db.yml
```

>** It's advised not to make skip_ssl_validation=true for non-development environment

## Register service

Log in to Cloud Foundry with admin user, and use the following commands to register `app-autoscaler` service

```sh
cf create-service-broker autoscaler <brokerUserName> <brokerPassword> <brokerURL>
```

* `brokerUserName`: the user name to authenticate with service broker
* `brokerPassword`: the password to authenticate with service broker
* `borkerURL`: the URL of the service broker

All these parameters are configured in the bosh deployment. If you are using default values of deployment manifest, register the service with the commands below.

```sh
cf create-service-broker autoscaler username password https://autoscalerservicebroker.bosh-lite.com
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
