package org.cloudfoundry.autoscaler.scheduler.conf;

import lombok.Getter;
import lombok.Setter;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Configuration;

@Getter
@Setter
@Configuration
@ConfigurationProperties(prefix = "cfserver.healthserver")
public class HealthServerProperties {
  private String username;
  private String password;
}
