package org.cloudfoundry.autoscaler.scheduler.rest;

import jakarta.servlet.Filter;
import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.ServletRequest;
import jakarta.servlet.ServletResponse;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;
import java.util.stream.Stream;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.codec.binary.Base64;
import org.cloudfoundry.autoscaler.scheduler.conf.HealthServerConfiguration;
import org.springframework.core.annotation.Order;
import org.springframework.stereotype.Component;
import org.springframework.util.ObjectUtils;

@Slf4j
@Component
@Order(2)
public class BasicAuthenticationFilter implements Filter {
  private static final Map<String, Boolean> protectedEndpointsMap;

  static {
    protectedEndpointsMap =
        Map.of(
            "/health/prometheus", true,
            "/health/liveness", true);
  }

  final HealthServerConfiguration healthServerConfiguration;

  public BasicAuthenticationFilter(HealthServerConfiguration healthServerConfiguration) {
    this.healthServerConfiguration = healthServerConfiguration;
  }

  @Override
  public void doFilter(
      ServletRequest servletRequest, ServletResponse servletResponse, FilterChain chain)
      throws IOException, ServletException {
    HttpServletRequest httpRequest = (HttpServletRequest) servletRequest;
    HttpServletResponse httpResponse = (HttpServletResponse) servletResponse;
    List<String> unprotectedEndpointsConfig = healthServerConfiguration.getUnprotectedEndpoints();

    if (!httpRequest.getRequestURI().contains("/health")) {
      log.debug("Not a health request: " + httpRequest.getRequestURI());
      chain.doFilter(servletRequest, servletResponse);
      return;
    }

    if (ObjectUtils.isEmpty(unprotectedEndpointsConfig)) { // health endpoints are authorized
      isUserAuthenticatedOrSendError(chain, httpRequest, httpResponse);
    } else if (!ObjectUtils.isEmpty(unprotectedEndpointsConfig)) {
      Map<String, Boolean> validateMap = checkValidEndpoints(unprotectedEndpointsConfig);
      if (!ObjectUtils.isEmpty(validateMap)) {
        log.warn(
            "Health configuration: invalid unprotectedEndpoints provided: "
                + validateMap.get("invalidEndpoints"));
        httpResponse.setHeader("WWW-Authenticate", "Basic");
        httpResponse.sendError(HttpServletResponse.SC_UNAUTHORIZED);
        return;
      }
      Map<String, Boolean> unprotectedConfig = getMapFromList(unprotectedEndpointsConfig);
      List<String> allowedEndpointsWithoutBasicAuth =
          areEndpointsAuthorized(unprotectedConfig, httpRequest.getRequestURI());

      if (!allowedEndpointsWithoutBasicAuth.contains(httpRequest.getRequestURI())) {
        log.error(
            "Health configuration: Basic auth is required to access protectedEndpoints: "
                + httpRequest.getRequestURI()
                + " \nValid unprotected endpoints are: "
                + allowedEndpointsWithoutBasicAuth);
        httpResponse.setHeader("WWW-Authenticate", "Basic");
        httpResponse.sendError(HttpServletResponse.SC_UNAUTHORIZED);
        return;
      }
      chain.doFilter(servletRequest, servletResponse);
    }
  }

  private void isUserAuthenticatedOrSendError(
      FilterChain filterChain, HttpServletRequest httpRequest, HttpServletResponse httpResponse)
      throws IOException, ServletException {
    final String authorizationHeader = httpRequest.getHeader("Authorization");

    if (healthServerConfiguration.getUsername() == null
        || healthServerConfiguration.getPassword() == null) {
      log.error("Health configuration: username || password not set");
      httpResponse.setHeader("WWW-Authenticate", "Basic");
      httpResponse.sendError(HttpServletResponse.SC_UNAUTHORIZED);
      return;
    }
    if (authorizationHeader == null) {
      log.error("Basic authentication not provided with the request");
      httpResponse.setHeader("WWW-Authenticate", "Basic");
      httpResponse.sendError(HttpServletResponse.SC_UNAUTHORIZED);
      return;
    }

    String base64Credentials = authorizationHeader.substring("Basic".length()).trim();
    byte[] credDecoded = Base64.decodeBase64(base64Credentials);
    String credentials = new String(credDecoded);
    String[] tokens = credentials.split(":");
    if (tokens.length != 2) {
      log.error("Malformed authorization header");
      httpResponse.sendError(HttpServletResponse.SC_BAD_REQUEST);
      return;
    }
    String username = tokens[0];
    String password = tokens[1];

    if (!areBasicAuthCredentialsCorrect(username, password)) {
      httpResponse.setHeader("WWW-Authenticate", "Basic");
      httpResponse.sendError(HttpServletResponse.SC_UNAUTHORIZED);
      return;
    }

    if (authorizationHeader != null && isUserAuthenticated(authorizationHeader)) {
      // allow access to health endpoints
      filterChain.doFilter(httpRequest, httpResponse);
    } else {
      httpResponse.setHeader("WWW-Authenticate", "Basic");
      httpResponse.sendError(HttpServletResponse.SC_UNAUTHORIZED);
    }
  }

  private Map<String, Boolean> checkValidEndpoints(List<String> unprotectedEndpointsConfig) {

    Map<String, Boolean> invalidEndpointsMap = new HashMap<>();
    for (String unprotectedEndpoint : unprotectedEndpointsConfig) {
      if (!protectedEndpointsMap.containsKey(unprotectedEndpoint)) {
        invalidEndpointsMap.put(unprotectedEndpoint, true);
      }
    }
    return invalidEndpointsMap;
  }

  private static Map<String, Boolean> getMapFromList(List<String> unprotectedEndpointsConfig) {
    return unprotectedEndpointsConfig.stream()
        .collect(Collectors.toMap(endpoint -> endpoint, endpoint -> true, (a, b) -> b));
  }

  private List<String> areEndpointsAuthorized(Map unprotectedEndpointsConfig, String requestURI) {
    Map<String, Boolean> resultUnprotectedEndpoints = new HashMap<>();
    for (Map.Entry<String, Boolean> protectedEndpoint : protectedEndpointsMap.entrySet()) {
      if (unprotectedEndpointsConfig.containsKey(protectedEndpoint.getKey())) {
        resultUnprotectedEndpoints.put(protectedEndpoint.getKey(), false);
      }
    }
    List<String> allowedEndpointsWithoutBasicAuth =
        keys(resultUnprotectedEndpoints, false).toList();
    if (isBasicAuthRequired(requestURI, allowedEndpointsWithoutBasicAuth)) {
      log.debug("Endpoints allowed without basic auth: + " + allowedEndpointsWithoutBasicAuth);
      return allowedEndpointsWithoutBasicAuth;
    }
    return allowedEndpointsWithoutBasicAuth;
  }

  private static boolean isBasicAuthRequired(String requestURI, List<String> allowedEndpoints) {
    return !ObjectUtils.isEmpty(allowedEndpoints) && allowedEndpoints.contains(requestURI);
  }

  private <K, V> Stream<K> keys(Map<K, V> map, V val) {
    return map.entrySet().stream()
        .filter(entry -> val.equals(entry.getValue()))
        .map(Map.Entry::getKey);
  }

  private boolean isUserAuthenticated(String authorization) {
    if (healthServerConfiguration.getUsername() == null
        || healthServerConfiguration.getPassword() == null) {
      log.error("Health configuration: username || password not set");
      return false;
    }
    if (authorization == null) {
      log.error("Basic authentication not provided with the request");
      return false;
    }

    String base64Credentials = authorization.substring("Basic".length()).trim();
    byte[] credDecoded = Base64.decodeBase64(base64Credentials);
    String credentials = new String(credDecoded);
    String[] tokens = credentials.split(":");
    if (tokens.length != 2) {
      log.error("Malformed authorization header");
      return false;
    }
    String username = tokens[0];
    String password = tokens[1];
    return areBasicAuthCredentialsCorrect(username, password);
  }

  private boolean areBasicAuthCredentialsCorrect(String username, String password) {
    return healthServerConfiguration.getUsername().equals(username)
        && healthServerConfiguration.getPassword().equals(password);
  }
}
