package org.cloudfoundry.autoscaler.scheduler.conf;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.StandardOpenOption;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.env.EnvironmentPostProcessor;
import org.springframework.core.env.ConfigurableEnvironment;
import org.springframework.core.env.MapPropertySource;
import org.springframework.core.env.PropertySource;

public class CloudFoundryConfigurationProcessor implements EnvironmentPostProcessor {

  private static final Logger logger = LoggerFactory.getLogger(CloudFoundryConfigurationProcessor.class);
  private static final String VCAP_SERVICES = "VCAP_SERVICES";
  private static final String SCHEDULER_CONFIG_TAG = "scheduler-config";
  private static final String DATABASE_TAG = "database";
  private static final ObjectMapper objectMapper = new ObjectMapper();

  @Override
  public void postProcessEnvironment(ConfigurableEnvironment environment, SpringApplication application) {
    String vcapServices = environment.getProperty(VCAP_SERVICES);
    if (vcapServices == null || vcapServices.trim().isEmpty()) {
      logger.debug("VCAP_SERVICES not found or empty, skipping Cloud Foundry configuration override");
      return;
    }

    try {
      Map<String, Object> allConfigs = new java.util.HashMap<>();
      
      // Process scheduler-config service
      Map<String, Object> schedulerConfig = extractSchedulerConfig(vcapServices);
      if (schedulerConfig != null && !schedulerConfig.isEmpty()) {
        logger.info("Found scheduler-config service in VCAP_SERVICES, applying configuration overrides");
        allConfigs.putAll(schedulerConfig);
      }
      
      // Process database services
      Map<String, Object> databaseConfigs = extractDatabaseConfigs(vcapServices);
      if (databaseConfigs != null && !databaseConfigs.isEmpty()) {
        logger.info("Found database services in VCAP_SERVICES, applying datasource configurations");
        allConfigs.putAll(databaseConfigs);
      }
      
      if (!allConfigs.isEmpty()) {
        Map<String, Object> flattenedConfig = flattenConfiguration("", allConfigs);
        PropertySource<?> cloudFoundryPropertySource = new MapPropertySource("cloudFoundryConfig", flattenedConfig);
        environment.getPropertySources().addFirst(cloudFoundryPropertySource);
        
        logger.info("Applied {} configuration properties from Cloud Foundry services", flattenedConfig.size());
      } else {
        logger.debug("No relevant Cloud Foundry services found in VCAP_SERVICES");
      }
    } catch (Exception e) {
      logger.error("Failed to process Cloud Foundry configuration from VCAP_SERVICES", e);
    }
  }

  private Map<String, Object> extractSchedulerConfig(String vcapServices) {
    try {
      TypeReference<Map<String, List<Map<String, Object>>>> typeRef = new TypeReference<Map<String, List<Map<String, Object>>>>() {};
      Map<String, List<Map<String, Object>>> services = objectMapper.readValue(vcapServices, typeRef);

      return services.values().stream()
          .flatMap(List::stream)
          .filter(this::hasSchedulerConfigTag)
          .findFirst()
          .map(service -> {
            Object credentials = service.get("credentials");
            if (credentials instanceof Map) {
              @SuppressWarnings("unchecked")
              Map<String, Object> credentialsMap = (Map<String, Object>) credentials;
              return credentialsMap;
            }
            return Map.<String, Object>of();
          })
          .orElse(null);
    } catch (Exception e) {
      logger.error("Failed to parse VCAP_SERVICES JSON", e);
      return null;
    }
  }

  private Map<String, Object> extractDatabaseConfigs(String vcapServices) {
    try {
      TypeReference<Map<String, List<Map<String, Object>>>> typeRef = new TypeReference<Map<String, List<Map<String, Object>>>>() {};
      Map<String, List<Map<String, Object>>> services = objectMapper.readValue(vcapServices, typeRef);

      Map<String, Object> databaseConfigs = new java.util.HashMap<>();
      
      services.values().stream()
          .flatMap(List::stream)
          .filter(this::hasDatabaseTag)
          .forEach(service -> {
            Object credentials = service.get("credentials");
            if (credentials instanceof Map) {
              @SuppressWarnings("unchecked")
              Map<String, Object> credentialsMap = (Map<String, Object>) credentials;
              Map<String, Object> datasourceConfig = mapDatabaseCredentialsToDataSource(credentialsMap, service);
              databaseConfigs.putAll(datasourceConfig);
            }
          });
      
      return databaseConfigs;
    } catch (Exception e) {
      logger.error("Failed to parse VCAP_SERVICES JSON for database services", e);
      return Map.of();
    }
  }

  private Map<String, Object> mapDatabaseCredentialsToDataSource(Map<String, Object> credentials, Map<String, Object> service) {
    Map<String, Object> config = new java.util.HashMap<>();
    
    // Get service tags to determine which datasources to configure
    Object tags = service.get("tags");
    if (!(tags instanceof List)) {
      return config;
    }
    
    @SuppressWarnings("unchecked")
    List<String> tagList = (List<String>) tags;
    
    // Build JDBC URL from credentials with SSL support
    String jdbcUrl = buildJdbcUrlWithSsl(credentials);
    String username = (String) credentials.get("username");
    String password = (String) credentials.get("password");
    
    if (jdbcUrl != null && username != null && password != null) {
      // Configure primary datasource if binding_db tag is present or no specific tags
      if (tagList.contains("binding_db") || (!tagList.contains("policy_db") && !tagList.contains("scalingengine_db"))) {
        config.put("spring.datasource.url", jdbcUrl);
        config.put("spring.datasource.username", username);
        config.put("spring.datasource.password", password);
        config.put("spring.datasource.driverClassName", "org.postgresql.Driver");
      }
      
      // Configure policy datasource if policy_db tag is present
      if (tagList.contains("policy_db")) {
        config.put("spring.policy-db-datasource.url", jdbcUrl);
        config.put("spring.policy-db-datasource.username", username);
        config.put("spring.policy-db-datasource.password", password);
        config.put("spring.policy-db-datasource.driverClassName", "org.postgresql.Driver");
      }
      
      logger.info("Configured datasources for database service with tags: {}", tagList);
    }
    
    return config;
  }
  
  private String buildJdbcUrlWithSsl(Map<String, Object> credentials) {
    String uri = (String) credentials.get("uri");
    String hostname = (String) credentials.get("hostname");
    Object portObj = credentials.get("port");
    String dbname = (String) credentials.get("dbname");
    
    // Build base URL
    String baseUrl = null;
    if (uri != null && uri.startsWith("postgres://")) {
      // Convert postgres:// URI to jdbc:postgresql:// format
      baseUrl = uri.replace("postgres://", "jdbc:postgresql://");
    } else if (hostname != null && portObj != null && dbname != null) {
      // Build from individual components
      String port = portObj.toString();
      baseUrl = String.format("jdbc:postgresql://%s:%s/%s", hostname, port, dbname);
    }
    
    if (baseUrl == null) {
      return null;
    }
    
    // Handle SSL certificates - support both direct and mapped credential names
    String sslCert = (String) credentials.get("sslcert");
    if (sslCert == null) {
      sslCert = (String) credentials.get("client_cert");
    }
    
    String sslKey = (String) credentials.get("sslkey");
    if (sslKey == null) {
      sslKey = (String) credentials.get("client_key");
    }
    
    String sslRootCert = (String) credentials.get("sslrootcert");
    
    StringBuilder urlBuilder = new StringBuilder(baseUrl);
    if (!baseUrl.contains("?")) {
      urlBuilder.append("?");
    } else {
      urlBuilder.append("&");
    }
    
    if (sslCert != null && !sslCert.trim().isEmpty()) {
      try {
        // Create temp directory for SSL certificates
        Path tempDir = Files.createTempDirectory("db-ssl-certs");
        tempDir.toFile().deleteOnExit();
        
        // Write SSL certificate to temp file
        Path sslCertPath = tempDir.resolve("ssl-cert.pem");
        Files.write(sslCertPath, sslCert.getBytes(), StandardOpenOption.CREATE, StandardOpenOption.WRITE);
        sslCertPath.toFile().deleteOnExit();
        urlBuilder.append("sslcert=").append(sslCertPath.toAbsolutePath()).append("&");
        
        // Write SSL key if available
        if (sslKey != null && !sslKey.trim().isEmpty()) {
          Path sslKeyPath = tempDir.resolve("ssl-key.pem");
          Files.write(sslKeyPath, sslKey.getBytes(), StandardOpenOption.CREATE, StandardOpenOption.WRITE);
          sslKeyPath.toFile().deleteOnExit();
          urlBuilder.append("sslkey=").append(sslKeyPath.toAbsolutePath()).append("&");
        }
        
        // Write SSL root certificate if available
        if (sslRootCert != null && !sslRootCert.trim().isEmpty()) {
          Path sslRootCertPath = tempDir.resolve("ssl-root-cert.pem");
          Files.write(sslRootCertPath, sslRootCert.getBytes(), StandardOpenOption.CREATE, StandardOpenOption.WRITE);
          sslRootCertPath.toFile().deleteOnExit();
          urlBuilder.append("sslrootcert=").append(sslRootCertPath.toAbsolutePath()).append("&");
        } else {
          // Use the provided sslcert as root cert too
          urlBuilder.append("sslrootcert=").append(sslCertPath.toAbsolutePath()).append("&");
        }
        
        urlBuilder.append("sslmode=require");
        
        logger.info("SSL certificates materialized to temporary directory: {}", tempDir.toAbsolutePath());
        
      } catch (IOException e) {
        logger.warn("Failed to materialize SSL certificates, falling back to SSL mode without cert files", e);
        urlBuilder.append("sslmode=require");
      }
    } else {
      // No SSL certificates provided, but still try to use SSL
      urlBuilder.append("sslmode=prefer");
    }
    
    return urlBuilder.toString();
  }

  private boolean hasSchedulerConfigTag(Map<String, Object> service) {
    Object tags = service.get("tags");
    if (tags instanceof List) {
      @SuppressWarnings("unchecked")
      List<String> tagList = (List<String>) tags;
      return tagList.contains(SCHEDULER_CONFIG_TAG);
    }
    return false;
  }

  private boolean hasDatabaseTag(Map<String, Object> service) {
    Object tags = service.get("tags");
    if (tags instanceof List) {
      @SuppressWarnings("unchecked")
      List<String> tagList = (List<String>) tags;
      return tagList.contains(DATABASE_TAG);
    }
    return false;
  }

  private Map<String, Object> flattenConfiguration(String prefix, Map<String, Object> config) {
    Map<String, Object> flattened = new java.util.HashMap<>();
    
    config.forEach((key, value) -> {
      String fullKey = prefix.isEmpty() ? key : prefix + "." + key;
      
      if (value instanceof Map) {
        @SuppressWarnings("unchecked")
        Map<String, Object> nestedMap = (Map<String, Object>) value;
        flattened.putAll(flattenConfiguration(fullKey, nestedMap));
      } else {
        flattened.put(fullKey, value);
      }
    });
    
    return flattened;
  }
}