package org.cloudfoundry.autoscaler.scheduler.util;

import org.cloudfoundry.autoscaler.scheduler.conf.HealthServerConfiguration;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

@Component
public class HealthUtils {

  @Autowired static HealthServerConfiguration healthServerConfig;

  private HealthUtils() {}

  public static String livenessUrl() {
    return "http://localhost:" + healthServerConfig.getPort() + "/health/liveness";
  }

  public static String prometheusMetricsUrl() {
    return "http://localhost:" + healthServerConfig.getPort() + "/health/prometheus";
  }
}
