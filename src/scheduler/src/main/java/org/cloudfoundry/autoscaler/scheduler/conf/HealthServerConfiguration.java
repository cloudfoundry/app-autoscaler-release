package org.cloudfoundry.autoscaler.scheduler.conf;

import jakarta.annotation.PostConstruct;
import java.util.List;
import java.util.Optional;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;
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
  private List<String> unprotectedEndpoints;

  @PostConstruct
  public void init() {

    validatePort();

    boolean basicAuthEnabled =
        (unprotectedEndpoints != null || ObjectUtils.isEmpty(unprotectedEndpoints));
    if (basicAuthEnabled
        && (this.username == null
            || this.password == null
            || this.username.isEmpty()
            || this.password.isEmpty())) {
      throw new IllegalArgumentException("Heath Server Basic Auth Username or password is not set");
    }
  }

  private void validatePort() {
    Optional<Integer> healthPortOptional = Optional.ofNullable(this.port);
    if (!healthPortOptional.isPresent() || healthPortOptional.get() == 0) {
      throw new IllegalArgumentException(
          "Health Configuration: health server port not defined or port=0");
    }
  }
}
