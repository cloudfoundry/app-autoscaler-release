# Migration Guide: BOSH Releases to MTAR Deployments

## Overview

This guide provides instructions for migrating App Autoscaler from BOSH-based deployments to MTAR (Multi-Target Application Archive) deployments using the MultiApps Controller. MTAR is the new supported deployment method for App Autoscaler.

**Important:** In the near future, this repository will transition to community maintenance. The new MTAR releases are now maintained at [cloudfoundry/app-autoscaler](https://github.com/cloudfoundry/app-autoscaler).

### Key Changes

- **Deployment Method**: BOSH VMs → MultiApps Controller (MTAR)
- **Maintenance Location**: Current repo → [cloudfoundry/app-autoscaler](https://github.com/cloudfoundry/app-autoscaler)
- **Acceptance Tests**: BOSH jobs → CF tasks (distributed within MTAR releases)
- **Service Discovery**: Route Registrar → Cloud Foundry routes

---

## Prerequisites

### 1. MultiApps Controller Setup

Before migrating, you must have a MultiApps Controller running on your Cloud Foundry infrastructure.

#### Deploy MultiApps Controller

The MultiApps Controller is a CF application that manages the deployment of multi-target applications. You need to deploy it to your Cloud Foundry environment before you can deploy MTARs.

**Step 1: Install the MultiApps CF CLI Plugin**

First, install the CLI plugin on your local machine:

```bash
# Check if already installed
cf plugins | grep multiapps

# Install the plugin
cf install-plugin multiapps
```

Or download from [GitHub releases](https://github.com/cloudfoundry/multiapps-cli-plugin/releases).

**Step 2: Set Up PostgreSQL Database**

The MultiApps Controller requires a PostgreSQL database. If you don't have one, deploy PostgreSQL or use an existing instance.

Create a database for the MultiApps Controller:

```sql
-- Connect to your PostgreSQL instance
CREATE DATABASE multiapps_controller WITH ENCODING='UTF8';
CREATE USER multiapps_user WITH PASSWORD 'your-secure-password';
GRANT ALL PRIVILEGES ON DATABASE multiapps_controller TO multiapps_user;
```

For BOSH-deployed PostgreSQL, you can add the database using an ops file (see `ci/operations/add-multiapps-databases-to-postgres.yml` for reference).

**Step 3: Download MultiApps Controller Release**

Download the WAR file and manifest from Maven Central:

```bash
# Set the version (check for latest at https://repo.maven.apache.org/maven2/org/cloudfoundry/multiapps/multiapps-controller-web/)
export MULTIAPPS_VERSION="1.174.0"

# Download the WAR file
wget "https://repo.maven.apache.org/maven2/org/cloudfoundry/multiapps/multiapps-controller-web/${MULTIAPPS_VERSION}/multiapps-controller-web-${MULTIAPPS_VERSION}.war"

# Download the manifest
wget "https://repo.maven.apache.org/maven2/org/cloudfoundry/multiapps/multiapps-controller-web/${MULTIAPPS_VERSION}/multiapps-controller-web-${MULTIAPPS_VERSION}-manifest.yml"
```

**Step 4: Create Database Service**

Create a user-provided service for the database connection:

```bash
# Set your database connection details
DB_USERNAME="multiapps_user"
DB_PASSWORD="your-secure-password"
DB_HOST="your-postgres-host"
DB_PORT="5432"
DB_NAME="multiapps_controller"

# Create the user-provided service
cf cups deploy-service-database -p "{
  \"uri\": \"postgres://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable\",
  \"username\": \"${DB_USERNAME}\",
  \"password\": \"${DB_PASSWORD}\"
}" -t postgres
```

**Step 5: Configure Security Groups (if needed)**

If your PostgreSQL is on a private network, create a security group to allow the MultiApps Controller to connect:

```bash
# Create security group JSON file
cat > multiapps-postgres-sg.json <<EOF
[
  {
    "protocol": "tcp",
    "destination": "YOUR_POSTGRES_IP/32",
    "ports": "5432",
    "description": "Allow egress to PostgreSQL"
  }
]
EOF

# Create and bind the security group
cf create-security-group multiapps-postgres-sg multiapps-postgres-sg.json
cf bind-security-group multiapps-postgres-sg <org> --space <space>
```

**Step 6: Deploy the MultiApps Controller**

Deploy using the downloaded manifest:

```bash
# Deploy the application
cf push deploy-service -f multiapps-controller-web-${MULTIAPPS_VERSION}-manifest.yml -p multiapps-controller-web-${MULTIAPPS_VERSION}.war

# Scale up to handle large MTARs (>1GB)
cf scale deploy-service -m 4G -k 2G -f
```

**Step 7: Verify the Deployment**

```bash
# Check the application is running
cf apps | grep deploy-service

# Test the MultiApps Controller endpoint
cf app deploy-service
```

The MultiApps Controller should now be accessible and ready to deploy MTAR files

#### Required Permissions

### 1. Ensure you have the following CF roles:
- **Space Developer** or **Space Manager** in the target space
- Access to create service instances and bind services
- Permissions to deploy applications

### 2. Database Access

You will need:
- **Connection details** for your current production PostgreSQL database
- **Admin credentials** to run database migrations
- **Backup** of the current database (recommended before migration)

### 3. Current BOSH Deployment Information

Gather the following from your existing BOSH deployment:
- Current service URLs and endpoints
- TLS certificates and credentials
- Database connection strings
- Service instance configurations

---

## Migration Steps

### Phase 1: Preparation

#### 1. Obtain MTAR Release

Download the MTAR release from [cloudfoundry/app-autoscaler releases](https://github.com/cloudfoundry/app-autoscaler/releases):

```bash
# Check available releases at: https://github.com/cloudfoundry/app-autoscaler/releases
# Set the version you want to use
export AUTOSCALER_VERSION="v15.12.2"

# Download the MTAR file
wget "https://github.com/cloudfoundry/app-autoscaler/releases/download/${AUTOSCALER_VERSION}/app-autoscaler.mtar"
```

#### 2. Create Build Extension File

To maintain compatibility with your existing URLs and configuration, create an MTA extension descriptor file (`.mtaext`). This file allows you to customize the MTAR deployment without modifying the base MTAR file.

**Example: `autoscaler-custom.mtaext`**

This extension file allows you to customize the MTAR deployment to use your existing BOSH URLs and database. See the `build-extension-file.sh` script in the [cloudfoundry/app-autoscaler](https://github.com/cloudfoundry/app-autoscaler/blob/main/build-extension-file.sh) repository for a complete example.

```yaml
_schema-version: "3.3.0"
ID: production
extends: com.github.cloudfoundry.app-autoscaler-release
version: 1.0.0

modules:
  - name: apiserver
    parameters:
      instances: 2
      routes:
        - route: autoscaler.${default-domain}
        - route: autoscalerservicebroker.${default-domain}

  - name: eventgenerator
    parameters:
      instances: 2

  - name: scalingengine
    parameters:
      instances: 2

  - name: metricsforwarder
    parameters:
      instances: 2
      routes:
        - route: autoscalermetrics.${default-domain}

  - name: operator
    parameters:
      instances: 2

  - name: scheduler
    parameters:
      instances: 2

resources:
  - name: database
    parameters:
      # PostgreSQL connection URI
      uri: "postgres://username:password@postgres-host:5432/autoscaler?sslmode=verify-full"
      # Optional: TLS certificates for secure database connection
      client_cert: |
        -----BEGIN CERTIFICATE-----
        ...
        -----END CERTIFICATE-----
      client_key: |
        -----BEGIN RSA PRIVATE KEY-----
        ...
        -----END RSA PRIVATE KEY-----
      server_ca: |
        -----BEGIN CERTIFICATE-----
        ...
        -----END CERTIFICATE-----

  - name: broker-catalog
    parameters:
      services:
        - name: autoscaler
          id: autoscaler-guid
          plans:
            - name: autoscaler-free-plan
              id: autoscaler-free-plan-guid
            - name: autoscaler-standard
              id: autoscaler-standard-guid
```

**Key Configuration Options:**

- **Modules**: The MTAR includes 6 core modules (apiserver, eventgenerator, scalingengine, metricsforwarder, operator, scheduler) plus a dbtasks module
- **Routes**: Map to your existing BOSH route registrar routes to maintain continuity. Use `${default-domain}` or specify full routes
- **Database**: Point to your current production PostgreSQL database with URI and optional TLS certificates
- **Instances**: Adjust instance counts for each module (default is 2 for high availability)
- **Broker Catalog**: Configure service plans and GUIDs

#### 3. Backup autoscaler DB

Backup your current database:

```bash
pg_dump -h <db_host> -U <db_user> -d autoscaler > autoscaler_backup_$(date +%F).sql
```

### Phase 1: Deploy MTAR

#### 1. Deploy Using MultiApps Plugin

Deploy the MTAR with your custom extension file:

```bash
# Log in to your CF environment
cf login -a https://api.your-cf-domain.com

# Target the org and space for deployment
cf target -o <org> -s <space>

# Deploy with extension file
cf deploy app-autoscaler.mtar -e production.mtaext
```

The deployment process will:
1. Upload the MTAR archive to the MultiApps Controller
2. Extract and validate the MTA descriptor and extension files
3. Run the `dbtasks` module to apply database migrations
4. Deploy all application modules in order:
   - apiserver 
   - eventgenerator 
   - scalingengine 
   - metricsforwarder 
   - operator 
   - scheduler 
5. Create and bind service instances (database, configurations)
6. Map routes to the deployed applications

**Deployment Options:**

```bash
cf deploy app-autoscaler.mtar -e production.mtaext 

```

#### 2. Monitor Deployment

```bash
# Monitor logs during deployment, some autoscaler traffic should get to the app instances.
cf logs apiserver --recent
cf logs scalingengine --recent
```

### Phase 3: Coexistence and Migration

During the migration period, BOSH VMs and MTAR deployment can coexist:

#### Coexistence Strategy

1. **Database Sharing**: Both deployments can connect to the same production database
   - Ensure connection limits are sufficient
   - Monitor database performance
   - Both BOSH and MTAR deployments will read from and write to the same database

2. **Route Configuration**: Ensure routes are properly configured
   ```bash
   # Verify MTAR routes are mapped
   cf routes | grep autoscaler

   # BOSH and MTAR should use the same route hostnames
   # The MTAR extension file should specify the same routes used by BOSH:
   # - autoscaler.<domain> (API)
   # - autoscalerservicebroker.<domain> (Service Broker)
   # - autoscalermetrics.<domain> (Metrics)
   ```

3. **Progressive VM Shutdown**: Stop BOSH VMs in order while verifying autoscaler continues to work

   **Recommended order:**

   a. **Stop Operator VMs first:**
   ```bash
   bosh -d app-autoscaler stop operator

   # Verify autoscaler is still working as expected
   ```

   b. **Stop EventGenerator VMs:**
   ```bash
   bosh -d app-autoscaler stop eventgenerator
   # Verify metrics collection is working on mtar MTAR deployment app
   ```

   c. **Stop ScalingEngine VMs:**
   ```bash
   bosh -d app-autoscaler stop scalingengine
   # Verify scaling events are executed by MTAR deployment app
   ```

   d. **Stop MetricsForwarder VMs:**
   ```bash
   bosh -d app-autoscaler stop metricsforwarder
   # Verify custom metrics are still being collected by MTAR deployment app
   ```

   e. **Stop Scheduler VMs:**
   ```bash
   bosh -d app-autoscaler stop scheduler
   # Verify scheduled scaling is working
   ```

   f. **Stop API/ServiceBroker VMs last:**
   ```bash
   bosh -d app-autoscaler stop apiserver

   # Verify API is still accessible
   curl https://autoscaler.<domain>/health
   ```

   **Between each step:**
   - Check application logs for errors
   - Verify autoscaling events are being triggered
   - If issues occur, restart the BOSH VMs and investigate

### Phase 4: Decommission BOSH Deployment

After successfully stopping all BOSH VMs progressively (in Phase 3) and verifying the MTAR deployment is handling all traffic:

1. **Final monitoring period:**
   - Monitor the MTAR deployment for 24-48 hours with all BOSH VMs stopped

2. **Verify all BOSH VMs are stopped:**
   ```bash
   # Check all VMs are stopped
   bosh -d app-autoscaler instances

   # Should show all instances in 'stopped' state
   ```

3. **Decommission based on database deployment:**

   **Scenario A: Database is external (separate from autoscaler deployment)**

   If your PostgreSQL database is deployed separately (not part of the app-autoscaler BOSH deployment), you can delete the entire deployment:

   ```bash
   # Delete the full BOSH deployment
   bosh -d app-autoscaler delete-deployment

   # Confirm deletion
   bosh deployments | grep app-autoscaler

   # Optional: Clean up the release
   bosh delete-release app-autoscaler
   ```

   **Scenario B: Database is deployed within the autoscaler deployment**

   If your PostgreSQL database is part of the app-autoscaler BOSH deployment and the MTAR is still using it, you must keep the deployment but scale down all non-database instances to 0:

   ```bash
   # Create an ops file to scale instances to 0
   cat > scale-to-zero.yml <<EOF
   # Scale all autoscaler instances to 0, keep only postgres
   - type: replace
     path: /instance_groups/name=apiserver/instances
     value: 0

   - type: replace
     path: /instance_groups/name=eventgenerator/instances
     value: 0

   - type: replace
     path: /instance_groups/name=scalingengine/instances
     value: 0

   - type: replace
     path: /instance_groups/name=metricsforwarder/instances
     value: 0

   - type: replace
     path: /instance_groups/name=operator/instances
     value: 0

   - type: replace
     path: /instance_groups/name=scheduler/instances
     value: 0
   EOF

   # Re-deploy with instances scaled to 0
   bosh -d app-autoscaler deploy <manifest.yml> -o scale-to-zero.yml

   # Verify only postgres is running
   bosh -d app-autoscaler instances
   ```

   **Note:** Keep this deployment running as long as the MTAR needs the database. Once you migrate to a separate database solution, you can delete the deployment.

---

## Acceptance Tests

Acceptance tests have been migrated to CF tasks and are now distributed within the MTAR releases.

### Running Acceptance Tests

TBD 

Acceptance tests are packaged as CF tasks in the MTAR deployment:

```bash
# List available tasks
cf apps | grep autoscaler-test

# Run acceptance tests
cf run-task autoscaler-acceptance-tests "api" --name acceptance-test-api
cf run-task autoscaler-acceptance-tests "broker" --name acceptance-test-broker
cf run-task autoscaler-acceptance-tests "app" --name acceptance-test-app


# View test results
cf logs autoscaler-acceptance-tests --recent
```

### Test Configuration

Configure tests via environment variables or task parameters:

```bash
cf set-env autoscaler-acceptance-tests API_ENDPOINT https://autoscaler.your-domain.com
cf set-env autoscaler-acceptance-tests CF_DOMAIN your-domain.com
cf restage autoscaler-acceptance-tests
```

---

### Updating and Maintenance

Going forward, updates will come from [cloudfoundry/app-autoscaler](https://github.com/cloudfoundry/app-autoscaler):

```bash
# Download new release (check https://github.com/cloudfoundry/app-autoscaler/releases for available versions)
export AUTOSCALER_VERSION="vNEW_VERSION"
export MTAR_FILE="app-autoscaler-release-${AUTOSCALER_VERSION}.mtar"
wget "https://github.com/cloudfoundry/app-autoscaler/releases/download/${AUTOSCALER_VERSION}/${MTAR_FILE}"

# Deploy update (blue-green deployment recommended)
cf deploy --strategy blue-green $MTAR_FILE -e autoscaler-custom.mtaext
```

### Community Contribution

As this repository transitions to community maintenance:
- Report issues at [cloudfoundry/app-autoscaler/issues](https://github.com/cloudfoundry/app-autoscaler/issues)
- Contribute improvements via pull requests
- Join community discussions

---

## Additional Resources

- [MultiApps Controller Documentation](https://github.com/cloudfoundry/multiapps-controller)
- [MTA Development Guide](https://help.sap.com/docs/BTP/65de2977205c403bbc107264b8eccf4b/d04fc0e2ad894545aebfd7126384307c.html)
- [App Autoscaler Repository](https://github.com/cloudfoundry/app-autoscaler)

