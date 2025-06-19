package org.cloudfoundry.autoscaler.scheduler.conf;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertTrue;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.boot.SpringApplication;
import org.springframework.core.env.ConfigurableEnvironment;
import org.springframework.core.env.StandardEnvironment;

public class CloudFoundryConfigurationProcessorTest {

  private CloudFoundryConfigurationProcessor processor;
  private ConfigurableEnvironment environment;
  private SpringApplication application;

  @BeforeEach
  public void setUp() {
    processor = new CloudFoundryConfigurationProcessor();
    environment = new StandardEnvironment();
    application = new SpringApplication();
  }

  @Test
  public void testNoVcapServices() {
    processor.postProcessEnvironment(environment, application);
    assertNull(environment.getProperty("spring.datasource.url"));
  }

  @Test
  public void testVcapServicesWithSchedulerConfig() {
    String vcapServices = """
        {
          "user-provided": [
            {
              "name": "scheduler-config-service",
              "tags": ["scheduler-config"],
              "credentials": {
                "spring": {
                  "datasource": {
                    "url": "jdbc:postgresql://cf-db-host:5432/autoscaler",
                    "username": "cf-db-user",
                    "password": "cf-db-password"
                  }
                },
                "autoscaler": {
                  "scalingengine": {
                    "url": "https://cf-scaling-engine:8091"
                  }
                },
                "server": {
                  "port": 8080
                }
              }
            }
          ]
        }
        """;

    environment.getPropertySources().addLast(
        new org.springframework.core.env.MapPropertySource("test",
            java.util.Map.of("VCAP_SERVICES", vcapServices)));

    processor.postProcessEnvironment(environment, application);

    assertEquals("jdbc:postgresql://cf-db-host:5432/autoscaler",
        environment.getProperty("spring.datasource.url"));
    assertEquals("cf-db-user", environment.getProperty("spring.datasource.username"));
    assertEquals("cf-db-password", environment.getProperty("spring.datasource.password"));
    assertEquals("https://cf-scaling-engine:8091",
        environment.getProperty("autoscaler.scalingengine.url"));
    assertEquals("8080", environment.getProperty("server.port"));
  }

  @Test
  public void testVcapServicesWithoutSchedulerConfigTag() {
    String vcapServices = """
        {
          "user-provided": [
            {
              "name": "other-service",
              "tags": ["other-tag"],
              "credentials": {
                "spring": {
                  "datasource": {
                    "url": "jdbc:postgresql://other-host:5432/other"
                  }
                }
              }
            }
          ]
        }
        """;

    environment.getPropertySources().addLast(
        new org.springframework.core.env.MapPropertySource("test",
            java.util.Map.of("VCAP_SERVICES", vcapServices)));

    processor.postProcessEnvironment(environment, application);

    assertNull(environment.getProperty("spring.datasource.url"));
  }

  @Test
  public void testInvalidVcapServicesJson() {
    String vcapServices = "invalid json";

    environment.getPropertySources().addLast(
        new org.springframework.core.env.MapPropertySource("test",
            java.util.Map.of("VCAP_SERVICES", vcapServices)));

    processor.postProcessEnvironment(environment, application);

    assertNull(environment.getProperty("spring.datasource.url"));
  }

  @Test
  public void testEmptyVcapServices() {
    environment.getPropertySources().addLast(
        new org.springframework.core.env.MapPropertySource("test",
            java.util.Map.of("VCAP_SERVICES", "")));

    processor.postProcessEnvironment(environment, application);

    assertNull(environment.getProperty("spring.datasource.url"));
  }

  @Test
  public void testVcapServicesWithDatabaseService() {
    String vcapServices = """
        {
          "postgresql-db": [
            {
              "label": "postgresql-db",
              "name": "autoscaler-db",
              "tags": ["relational", "binding_db", "policy_db"],
              "credentials": {
                "username": "dbuser",
                "password": "dbpass",
                "hostname": "db-host.example.com",
                "dbname": "autoscaler_db",
                "port": "5432",
                "uri": "postgres://dbuser:dbpass@db-host.example.com:5432/autoscaler_db",
                "sslcert": "-----BEGIN CERTIFICATE-----\\nMIICert...\\n-----END CERTIFICATE-----",
                "sslrootcert": "-----BEGIN CERTIFICATE-----\\nMIIRoot...\\n-----END CERTIFICATE-----"
              }
            }
          ]
        }
        """;

    environment.getPropertySources().addLast(
        new org.springframework.core.env.MapPropertySource("test",
            java.util.Map.of("VCAP_SERVICES", vcapServices)));

    processor.postProcessEnvironment(environment, application);

    String datasourceUrl = environment.getProperty("spring.datasource.url");
    assertNotNull(datasourceUrl);
    assertEquals("jdbc:postgresql://db-host.example.com:5432/autoscaler_db?sslmode=require", datasourceUrl);

    assertEquals("dbuser", environment.getProperty("spring.datasource.username"));
    assertEquals("dbpass", environment.getProperty("spring.datasource.password"));
    assertEquals("org.postgresql.Driver", environment.getProperty("spring.datasource.driverClassName"));

    // Should also configure policy datasource since policy_db tag is present
    String policyDatasourceUrl = environment.getProperty("spring.policy-db-datasource.url");
    assertNotNull(policyDatasourceUrl);
    assertEquals("dbuser", environment.getProperty("spring.policy-db-datasource.username"));
    assertEquals("dbpass", environment.getProperty("spring.policy-db-datasource.password"));
  }

  @Test
  public void testVcapServicesWithDatabaseServiceNoSsl() {
    String vcapServices = """
        {
          "postgresql-db": [
            {
              "label": "postgresql-db",
              "name": "autoscaler-db",
              "tags": ["relational", "binding_db"],
              "credentials": {
                "username": "dbuser",
                "password": "dbpass",
                "hostname": "db-host.example.com",
                "dbname": "autoscaler_db",
                "port": "5432"
              }
            }
          ]
        }
        """;

    environment.getPropertySources().addLast(
        new org.springframework.core.env.MapPropertySource("test",
            java.util.Map.of("VCAP_SERVICES", vcapServices)));

    processor.postProcessEnvironment(environment, application);

    String datasourceUrl = environment.getProperty("spring.datasource.url");
    assertNotNull(datasourceUrl);
    assertEquals("jdbc:postgresql://db-host.example.com:5432/autoscaler_db?sslmode=require", datasourceUrl);
    assertEquals("dbuser", environment.getProperty("spring.datasource.username"));
    assertEquals("dbpass", environment.getProperty("spring.datasource.password"));

    // Should configure policy datasource since binding_db tag is present
    assertNotNull(environment.getProperty("spring.policy-db-datasource.url"));
    assertEquals("jdbc:postgresql://db-host.example.com:5432/autoscaler_db?sslmode=require", environment.getProperty("spring.policy-db-datasource.url"));
  }

  @Test
  public void testVcapServicesWithClientCertCredentialMapping() {
    String vcapServices = """
        {
          "postgresql-db": [
            {
              "label": "postgresql-db",
              "name": "autoscaler-db",
              "tags": ["relational", "binding_db"],
              "credentials": {
                "username": "dbuser",
                "password": "dbpass",
                "hostname": "db-host.example.com",
                "dbname": "autoscaler_db",
                "port": "5432",
                "client_cert": "-----BEGIN CERTIFICATE-----\\nMIICert...\\n-----END CERTIFICATE-----",
                "client_key": "-----BEGIN PRIVATE KEY-----\\nMIIKey...\\n-----END PRIVATE KEY-----",
                "sslrootcert": "-----BEGIN CERTIFICATE-----\\nMIIRoot...\\n-----END CERTIFICATE-----"
              }
            }
          ]
        }
        """;

    environment.getPropertySources().addLast(
        new org.springframework.core.env.MapPropertySource("test",
            java.util.Map.of("VCAP_SERVICES", vcapServices)));

    processor.postProcessEnvironment(environment, application);

    String datasourceUrl = environment.getProperty("spring.datasource.url");
    assertNotNull(datasourceUrl);
    assertEquals("jdbc:postgresql://db-host.example.com:5432/autoscaler_db?sslmode=require", datasourceUrl);

    assertEquals("dbuser", environment.getProperty("spring.datasource.username"));
    assertEquals("dbpass", environment.getProperty("spring.datasource.password"));
    assertEquals("org.postgresql.Driver", environment.getProperty("spring.datasource.driverClassName"));
  }

  @Test
  public void testVcapServicesWithClientCertOnlyCredentialMapping() {
    String vcapServices = """
        {
          "postgresql-db": [
            {
              "label": "postgresql-db",
              "name": "autoscaler-db",
              "tags": ["relational", "binding_db"],
              "credentials": {
                "username": "dbuser",
                "password": "dbpass",
                "hostname": "db-host.example.com",
                "dbname": "autoscaler_db",
                "port": "5432",
                "client_cert": "-----BEGIN CERTIFICATE-----\\nMIICert...\\n-----END CERTIFICATE-----"
              }
            }
          ]
        }
        """;

    environment.getPropertySources().addLast(
        new org.springframework.core.env.MapPropertySource("test",
            java.util.Map.of("VCAP_SERVICES", vcapServices)));

    processor.postProcessEnvironment(environment, application);

    String datasourceUrl = environment.getProperty("spring.datasource.url");
    assertNotNull(datasourceUrl);
    assertEquals("jdbc:postgresql://db-host.example.com:5432/autoscaler_db?sslmode=require", datasourceUrl);
  }

  @Test
  public void testVcapServicesPrefersSslcertOverClientCert() {
    String vcapServices = """
        {
          "postgresql-db": [
            {
              "label": "postgresql-db",
              "name": "autoscaler-db",
              "tags": ["relational", "binding_db"],
              "credentials": {
                "username": "dbuser",
                "password": "dbpass",
                "hostname": "db-host.example.com",
                "dbname": "autoscaler_db",
                "port": "5432",
                "sslcert": "-----BEGIN CERTIFICATE-----\\nMIISSLCert...\\n-----END CERTIFICATE-----",
                "sslkey": "-----BEGIN PRIVATE KEY-----\\nMIISSLKey...\\n-----END PRIVATE KEY-----",
                "client_cert": "-----BEGIN CERTIFICATE-----\\nMIICert...\\n-----END CERTIFICATE-----",
                "client_key": "-----BEGIN PRIVATE KEY-----\\nMIIKey...\\n-----END PRIVATE KEY-----"
              }
            }
          ]
        }
        """;

    environment.getPropertySources().addLast(
        new org.springframework.core.env.MapPropertySource("test",
            java.util.Map.of("VCAP_SERVICES", vcapServices)));

    processor.postProcessEnvironment(environment, application);

    String datasourceUrl = environment.getProperty("spring.datasource.url");
    assertNotNull(datasourceUrl);
    assertEquals("jdbc:postgresql://db-host.example.com:5432/autoscaler_db?sslmode=require", datasourceUrl);
  }
}
