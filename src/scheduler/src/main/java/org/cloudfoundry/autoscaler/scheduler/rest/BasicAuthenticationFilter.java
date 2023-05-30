package org.cloudfoundry.autoscaler.scheduler.rest;

import jakarta.servlet.Filter;
import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.ServletRequest;
import jakarta.servlet.ServletResponse;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.util.Map;
import java.util.Set;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.codec.binary.Base64;
import org.cloudfoundry.autoscaler.scheduler.conf.HealthServerConfiguration;
import org.springframework.core.annotation.Order;
import org.springframework.http.HttpHeaders;
import org.springframework.stereotype.Component;
import org.springframework.util.ObjectUtils;

@Slf4j
@Component
@Order(2)
public class BasicAuthenticationFilter implements Filter {
  private static final String WWW_AUTHENTICATE_VALUE = "Basic";

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
    Set<String> unprotectedEndpointsConfig = healthServerConfiguration.getUnprotectedEndpoints();

    if (!httpRequest.getRequestURI().contains("/health")) {
      log.debug("Not a health request: " + httpRequest.getRequestURI());
      chain.doFilter(servletRequest, servletResponse);
      return;
    }
    final boolean allEndpointsRequireAuthorization =
        ObjectUtils.isEmpty(unprotectedEndpointsConfig);

    if (allEndpointsRequireAuthorization) { // case 1 ; config []
      allowAuthenticatedRequest(chain, httpRequest, httpResponse);

    } else if (!ObjectUtils.isEmpty(unprotectedEndpointsConfig)) {
      /*
      // case 2.1 ; config ["/health/liveness"]
      here is – by configuration – one protected endpoint "/health/prometheus" and one unprotected "/health/liveness".
      The user is authenticated.
       The user queries on "/health/prometheus".
       Expected behaviour: The request will be handled successfully.
       */

      // IMPORTANT: Match the configured health endpoints with the allowed list of endpoints to
      // cover
      // BasicAuthenticationFilterTest#denyHealthRequestWithWrongUnprotectedEndpoints()
      // Suggestion: THe following block should be part of HealthConfiguration.
      // Move this block to Health Configuration/Or adjust test
      if (!healthConfigsExists()) {
        log.error("Health Configuration: Invalid endpoints defined");
        httpResponse.setHeader(HttpHeaders.WWW_AUTHENTICATE, WWW_AUTHENTICATE_VALUE);
        httpResponse.sendError(HttpServletResponse.SC_UNAUTHORIZED);
        return;
      }

      boolean isEndpointRequireAuthentication =
          isEndpointRequireAuthentication(httpRequest.getRequestURI());
      if (isEndpointRequireAuthentication) {
        allowAuthenticatedRequest(chain, httpRequest, httpResponse);
      } else { // RequestURI() does not need authentication (
        // Case 2.2 ; config  ["/health/liveness", "/health/prometheus"]
        chain.doFilter(servletRequest, servletResponse);
      }
    }
  }

  private boolean healthConfigsExists() {
    boolean found = false;
    Map<String, Boolean> validProtectedEndpoints =
        healthServerConfiguration.getValidProtectedEndpoints();
    for (String configuredEndpoint : healthServerConfiguration.getUnprotectedEndpoints()) {
      found = validProtectedEndpoints.containsKey(configuredEndpoint);
    }
    if (!found) {
      return false;
    }
    return true;
  }

  private void allowAuthenticatedRequest(
      FilterChain filterChain, HttpServletRequest httpRequest, HttpServletResponse httpResponse)
      throws IOException, ServletException {
    final String authorizationHeader = httpRequest.getHeader("Authorization");

    if (authorizationHeader == null) {
      log.error("Basic authentication not provided with the request");
      httpResponse.setHeader(HttpHeaders.WWW_AUTHENTICATE, WWW_AUTHENTICATE_VALUE);
      httpResponse.sendError(HttpServletResponse.SC_UNAUTHORIZED);
      return;
    }

    String[] tokens = decodeAndGetAuthTokens(authorizationHeader);
    if (tokens.length != 2) {
      log.error("Malformed authorization header");
      httpResponse.sendError(HttpServletResponse.SC_BAD_REQUEST);
      return;
    }

    if (isUserAuthenticated(authorizationHeader)) {
      // allow access to health endpoints
      filterChain.doFilter(httpRequest, httpResponse);
    } else {
      log.error(
          "Health configuration: Basic auth is required to access protectedEndpoints: "
              + httpRequest.getRequestURI()
              + " \nValid unprotected endpoints are: "
              + healthServerConfiguration.getUnprotectedEndpoints());
      httpResponse.setHeader(HttpHeaders.WWW_AUTHENTICATE, WWW_AUTHENTICATE_VALUE);
      httpResponse.sendError(HttpServletResponse.SC_UNAUTHORIZED);
    }
  }

  private boolean isEndpointRequireAuthentication(String requestURI) {
    Map<String, Boolean> protectedEndpoints =
        healthServerConfiguration.getValidProtectedEndpoints();
    boolean isProtected = protectedEndpoints.containsKey(requestURI);
    boolean isUnprotectedByConfiguration =
        healthServerConfiguration.getUnprotectedEndpoints().contains(requestURI);

    return isProtected && !isUnprotectedByConfiguration;
  }

  private boolean isUserAuthenticated(String authorization) {

    if (authorization == null) {
      log.error("Basic authentication not provided with the request");
      return false;
    }

    String[] tokens = decodeAndGetAuthTokens(authorization);
    if (tokens.length != 2) {
      log.error("Malformed authorization header");
      return false;
    }
    String username = tokens[0];
    String password = tokens[1];
    return areBasicAuthCredentialsCorrect(username, password);
  }

  private static String[] decodeAndGetAuthTokens(String authorization) {
    String base64Credentials = authorization.substring(WWW_AUTHENTICATE_VALUE.length()).trim();
    byte[] credDecoded = Base64.decodeBase64(base64Credentials);
    String credentials = new String(credDecoded);
    String[] tokens = credentials.split(":");
    return tokens;
  }

  private boolean areBasicAuthCredentialsCorrect(String username, String password) {
    return healthServerConfiguration.getUsername().equals(username)
        && healthServerConfiguration.getPassword().equals(password);
  }
}
