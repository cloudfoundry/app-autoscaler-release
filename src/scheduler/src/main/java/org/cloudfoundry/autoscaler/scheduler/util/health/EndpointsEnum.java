package org.cloudfoundry.autoscaler.scheduler.util.health;

import java.util.Set;

public enum EndpointsEnum {
  PROMETHEUS("/health/prometheus"),
  LIVENESS("/health/liveness");

  private final String url;

  EndpointsEnum(String url) {
    this.url = url;
  }

  public String getUrl() {
    return url;
  }

  public static EndpointsEnum valueOfEndpoint(String url) {
    for (EndpointsEnum endpoint : values()) {
      if (endpoint.url.equals(url)) {
        return endpoint;
      }
    }
    throw new IllegalArgumentException("Enum for " + url);
  }

  public static boolean isValidEndpoint(String url) {
    EndpointsEnum endpointsEnum = valueOfEndpoint(url);
    if (endpointsEnum != null) {
      return true;
    }
    return false;
  }

  public static boolean configuredEndpointsExists(Set<String> userDefindHealthEndpoints) {

    for (String configuredEndpoint : userDefindHealthEndpoints) {
      return isValidEndpoint(configuredEndpoint);
    }
    return false;
  }

  public static String displayAllEndpointValues() {
    String endpointValues = EndpointsEnum.values()[0].getUrl();
    for (int i = 1; i < EndpointsEnum.values().length; i++) {
      endpointValues += "," + EndpointsEnum.values()[i].getUrl();
    }
    return endpointValues;
  }
}
