package org.cloudfoundry.autoscaler.scheduler.conf;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNull;

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
}