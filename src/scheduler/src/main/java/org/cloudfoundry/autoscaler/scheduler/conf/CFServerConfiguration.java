package org.cloudfoundry.autoscaler.scheduler.conf;

import jakarta.annotation.PostConstruct;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@ConfigurationProperties(prefix = "cf_server")
@Data
@Component
@AllArgsConstructor
@NoArgsConstructor
public class CFServerConfiguration{
  private int port;
  private String validOrgGuid;
  private String validSpaceGuid;

  @PostConstruct
  public void init() {
    if (this.port <= 0) {
      throw new IllegalStateException("CF Server Port is not set");
    }
    if (this.validOrgGuid == null || this.validOrgGuid.isEmpty()) {
      throw new IllegalStateException("CF Server validOrgGuid is not set");
    }
  }

  public long getSocketTimeout() {
    return 10;
  }
}
