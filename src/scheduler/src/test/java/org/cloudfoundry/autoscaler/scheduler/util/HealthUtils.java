package org.cloudfoundry.autoscaler.scheduler.util;

import java.net.MalformedURLException;
import java.net.URL;
import org.cloudfoundry.autoscaler.scheduler.conf.HealthServerConfiguration;
import org.springframework.stereotype.Component;

@Component
public class HealthUtils {

  static HealthServerConfiguration healthServerConfig;

  private HealthUtils(HealthServerConfiguration healthServerConfig) {
    this.healthServerConfig = healthServerConfig;
  }

  public static URL livenessUrl() throws MalformedURLException {
    return new URL("http://localhost:" + healthServerConfig.getPort() + "/health/liveness");
  }

  public static URL prometheusMetricsUrl() throws MalformedURLException {
    return new URL("http://localhost:" + healthServerConfig.getPort() + "/health/prometheus");
  }

  public static URL baseLivenessUrl() throws MalformedURLException {
    return new URL("http://localhost:" + healthServerConfig.getPort() + "/health");
  }
}
