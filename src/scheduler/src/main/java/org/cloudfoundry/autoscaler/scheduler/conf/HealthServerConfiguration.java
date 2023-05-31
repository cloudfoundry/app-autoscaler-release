package org.cloudfoundry.autoscaler.scheduler.conf;

import jakarta.annotation.PostConstruct;
import java.util.HashMap;
import java.util.Map;
import java.util.Optional;
import java.util.Set;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.cloudfoundry.autoscaler.scheduler.util.health.EndpointsEnum;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;
import org.springframework.util.ObjectUtils;

@ConfigurationProperties(prefix = "scheduler.healthserver")
@Data
@Component
@AllArgsConstructor
@NoArgsConstructor
public class HealthServerConfiguration {
  private String username;
  private String password;
  private Integer port;
  private Set<String> unprotectedEndpoints;

  @PostConstruct
  public void init() {

    validatePort();
    validateConfiguredEndpoints();

    // We need the username and password in health configuration if and only if
    // -  atleast one endpoints exists in the unprotectedEndpoints configuration
    // -  and the unprotectedEndpoints is empty => all endpoints are protected
    boolean basicAuthEnabled =
        (unprotectedEndpoints != null || ObjectUtils.isEmpty(unprotectedEndpoints));
    if (basicAuthEnabled
        && (this.username == null
            || this.password == null
            || this.username.isEmpty()
            || this.password.isEmpty())) {
      throw new IllegalArgumentException(
          "Health Server Basic Auth Username or password is not set");
    }
  }

  private void validatePort() {
    Optional<Integer> healthPortOptional = Optional.ofNullable(this.port);
    if (!healthPortOptional.isPresent() || healthPortOptional.get() == 0) {
      throw new IllegalArgumentException(
          "Health Configuration: health server port not defined or set to unsupported port-number"
              + " `0`");
    }
  }

  private void validateConfiguredEndpoints() {

    Map<String, Boolean> invalidEndpointsMap = new HashMap<>();
    if (unprotectedEndpoints == null) {
      return;
    }
    for (String unprotectedEndpoint : unprotectedEndpoints) {

      if (!EndpointsEnum.isValidEndpoint(unprotectedEndpoint)) {
        invalidEndpointsMap.put(unprotectedEndpoint, true);
      }
    }
    if (!ObjectUtils.isEmpty(invalidEndpointsMap)) {
      throw new IllegalArgumentException(
          "Health configuration: invalid unprotectedEndpoints provided: "
              + invalidEndpointsMap
              + "\n"
              + "validate endpoints are: "
              + EndpointsEnum.displayAllEndpointValues());
    }
  }

  public boolean isUnprotectedByConfiguration(String requestURI) {
    return this.getUnprotectedEndpoints().contains(requestURI);
  }
}
