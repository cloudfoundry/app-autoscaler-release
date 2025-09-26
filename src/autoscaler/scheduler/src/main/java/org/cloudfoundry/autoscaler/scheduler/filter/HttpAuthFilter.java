package org.cloudfoundry.autoscaler.scheduler.filter;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import lombok.RequiredArgsConstructor;
import lombok.Setter;
import org.cloudfoundry.autoscaler.scheduler.conf.HealthServerProperties;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.core.annotation.Order;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.ByteArrayInputStream;
import java.io.IOException;
import java.security.cert.CertificateException;
import java.security.cert.CertificateFactory;
import java.security.cert.X509Certificate;
import java.util.Base64;

@Component
@Order(0)
@RequiredArgsConstructor
@Setter
public class HttpAuthFilter extends OncePerRequestFilter {

  private static final String AUTHORIZATION_HEADER = "Authorization";
  private static final String BASIC_AUTH_PREFIX = "Basic ";
  private static final String XFCC_HEADER = "X-Forwarded-Client-Cert";
  private static final String HEALTH_ENDPOINT = "/health";

  private Logger logger = LoggerFactory.getLogger(this.getClass());
  @Qualifier("healthServerProperties")
  private final HealthServerProperties healthConfig;
  private String validSpaceGuid;
  private String validOrgGuid;

  @Override
  protected void doFilterInternal(
    HttpServletRequest request, HttpServletResponse response, FilterChain filterChain)
    throws ServletException, IOException {

    String forwardedProto = request.getHeader("X-Forwarded-Proto");
    boolean isHealthEndpoint = request.getRequestURI().contains(HEALTH_ENDPOINT);

    logger.info(
      "Received {} request, scheme={},X-Forwarded-Proto={} isHealthEndpoint={}",
      request.getMethod(),
      request.getScheme(),
      forwardedProto,
      isHealthEndpoint);

    if (isHealthEndpoint) {
         handleHealthEndpoint(request, response);
      return;
    }

    // Skip filter if X-Forwarded-Client-Cert is missing and not a health request
    String xfccHeader = request.getHeader(XFCC_HEADER);
    if (xfccHeader == null || xfccHeader.isEmpty()) {
      logger.warn("Missing X-Forwarded-Client-Cert header");
      response.sendError(
        HttpServletResponse.SC_BAD_REQUEST,
        "Missing X-Forwarded-Client-Cert header in the request");
      return;
    }
    logger.info(
      "X-Forwarded-Client-Cert header received ... checking authorized org and space in OU");
    try {
      String organizationalUnit = extractOrganizationalUnit(xfccHeader);
      // Validate both key-value pairs in OrganizationalUnit
      if (!isValidOrganizationalUnit(organizationalUnit)) {
        logger.warn("Unauthorized OrganizationalUnit: " + organizationalUnit);
        response.sendError(HttpServletResponse.SC_FORBIDDEN, "Unauthorized OrganizationalUnit");
        return;
      }
    } catch (CertificateException e) {
      logger.warn("Invalid certificate: " + e.getMessage());
      response.sendError(
        HttpServletResponse.SC_BAD_REQUEST, "Invalid certificate: " + e.getMessage());
      return;
    }
    // Proceed with the request
    filterChain.doFilter(request, response);
  }

  private void handleHealthEndpoint(HttpServletRequest request, HttpServletResponse response) throws IOException {
    logger.info("Handling health check request with Basic Auth");
    String authHeader = request.getHeader(AUTHORIZATION_HEADER);
    logger.info("Authorization header: {}", authHeader != null ? "present" : "missing");

    if (authHeader == null || !authHeader.startsWith(BASIC_AUTH_PREFIX)) {
      logger.warn("Missing or invalid Authorization header for health check request");
      response.sendError(HttpServletResponse.SC_UNAUTHORIZED, "Unauthorized");
      return;
    }
    String[] credentials = decodeBasicAuth(authHeader);
    if (credentials.length != 2) {
      logger.warn("Invalid Authorization header format for health check request");
      response.sendError(HttpServletResponse.SC_BAD_REQUEST, "Bad Request");
      return;
    }
    if (!credentials[0].equals(healthConfig.getUsername())
      || !credentials[1].equals(healthConfig.getPassword())) {
      logger.warn("Invalid credentials for health check request");
      response.sendError(HttpServletResponse.SC_UNAUTHORIZED, "Unauthorized");
      return;
    }

    response.setStatus(HttpServletResponse.SC_OK);
    response.setContentType("application/json");
    response.getWriter().write("{\"status\":\"UP\"}");
    response.getWriter().flush();
  }

  private String[] decodeBasicAuth(String authHeader) {
    try {
      return new String(Base64.getDecoder().decode(authHeader.substring(BASIC_AUTH_PREFIX.length())))
        .split(":");
    } catch (IllegalArgumentException e) {
      logger.warn("Failed to decode Basic Auth header: {}", e.getMessage());
      return null;
    }
  }
  private String extractOrganizationalUnit(String certValue) throws CertificateException {
    X509Certificate certificate = parseCertificate(certValue);
    return certificate.getSubjectX500Principal().getName();
  }

  private X509Certificate parseCertificate(String certValue) throws CertificateException {
    // Extract the base64-encoded certificate from the XFCC header
    String base64Cert =
      certValue
        .replace("-----BEGIN CERTIFICATE-----", "")
        .replace("-----END CERTIFICATE-----", "")
        .replaceAll("\\s+", "");

    byte[] decodedCert = Base64.getDecoder().decode(base64Cert);

    CertificateFactory factory = CertificateFactory.getInstance("X.509");
    return (X509Certificate) factory.generateCertificate(new ByteArrayInputStream(decodedCert));
  }

  private boolean isValidOrganizationalUnit(String organizationalUnit) {
    boolean isSpaceValid = organizationalUnit.contains("space:" + validSpaceGuid);
    boolean isOrgValid = organizationalUnit.contains("organization:" + validOrgGuid);
    return isSpaceValid && isOrgValid;
  }
}
