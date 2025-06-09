package org.cloudfoundry.autoscaler.scheduler.conf;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
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
  private static final ObjectMapper objectMapper = new ObjectMapper();

  @Override
  public void postProcessEnvironment(ConfigurableEnvironment environment, SpringApplication application) {
    String vcapServices = environment.getProperty(VCAP_SERVICES);
    if (vcapServices == null || vcapServices.trim().isEmpty()) {
      logger.debug("VCAP_SERVICES not found or empty, skipping Cloud Foundry configuration override");
      return;
    }

    try {
      Map<String, Object> schedulerConfig = extractSchedulerConfig(vcapServices);
      if (schedulerConfig != null && !schedulerConfig.isEmpty()) {
        logger.info("Found scheduler-config service in VCAP_SERVICES, applying configuration overrides");
        
        Map<String, Object> flattenedConfig = flattenConfiguration("", schedulerConfig);
        PropertySource<?> cloudFoundryPropertySource = new MapPropertySource("cloudFoundrySchedulerConfig", flattenedConfig);
        environment.getPropertySources().addFirst(cloudFoundryPropertySource);
        
        logger.info("Applied {} configuration properties from Cloud Foundry scheduler-config service", flattenedConfig.size());
      } else {
        logger.debug("No scheduler-config service found in VCAP_SERVICES");
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

  private boolean hasSchedulerConfigTag(Map<String, Object> service) {
    Object tags = service.get("tags");
    if (tags instanceof List) {
      @SuppressWarnings("unchecked")
      List<String> tagList = (List<String>) tags;
      return tagList.contains(SCHEDULER_CONFIG_TAG);
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