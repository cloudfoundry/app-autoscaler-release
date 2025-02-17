package org.cloudfoundry.autoscaler.scheduler.filter;

import jakarta.servlet.FilterChain;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import lombok.Setter;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.core.annotation.Order;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.ByteArrayInputStream;
import java.io.IOException;
import java.security.cert.CertificateFactory;
import java.security.cert.X509Certificate;
import java.util.Base64;

@Component
@Order(0)
@ConfigurationProperties(prefix = "cfserver")
@Setter
public class HttpAuthFilter extends OncePerRequestFilter {
  private Logger logger = LoggerFactory.getLogger(this.getClass());

  private String validSpaceGuid;
  private String validOrgGuid;

  @Override
  protected void doFilterInternal(
    HttpServletRequest request, HttpServletResponse response, FilterChain filterChain)
    throws jakarta.servlet.ServletException, IOException {

    logger.info("Received request with request " + request.getRequestURI() + " method" + request.getMethod());

    // Skip filter if the request is HTTPS
    if (request.isSecure()) {
      filterChain.doFilter(request, response);
      return;
    }
    String xfccHeader = request.getHeader("X-Forwarded-Client-Cert");
    if (xfccHeader == null || xfccHeader.isEmpty()) {
      logger.warn("Missing X-Forwarded-Client-Cert header");
      response.sendError(
        HttpServletResponse.SC_BAD_REQUEST, "Missing X-Forwarded-Client-Cert header in the request");
      return;
    }
    logger.info(
      "X-Forwarded-Client-Cert header received ... checking authorized org and space in OU");

    // xfccHeader in semicolom separated Cert=some-cert;Hash=some-hash with regular expression
    String[] parts = xfccHeader.split(";");
    String certValue = "";

    // Loop through the parts to find the one that starts with "Cert="
    for (String part : parts) {
      if (part.startsWith("Cert=")) {
        certValue = part.substring("Cert=".length());
        break;
      }
    }
    try {
       String organizationalUnit = extractOrganizationalUnit(certValue);

      // Validate both key-value pairs in OrganizationalUnit
      if (!isValidOrganizationalUnit(organizationalUnit)) {
        logger.warn("Unauthorized OrganizationalUnit: " + organizationalUnit);
        response.sendError(HttpServletResponse.SC_FORBIDDEN, "Unauthorized OrganizationalUnit");
        return;
      }
    } catch (Exception e) {
      logger.warn("Invalid certificate: " + e.getMessage());
      response.sendError(
        HttpServletResponse.SC_BAD_REQUEST, "Invalid certificate: " + e.getMessage());
      return;
    }
    // Proceed with the request
    filterChain.doFilter(request, response);
  }

  private String extractOrganizationalUnit(String certValue) throws Exception {
    X509Certificate certificate = parseCertificate(certValue);
    return certificate.getSubjectX500Principal().getName();
  }

  private X509Certificate parseCertificate(String certValue) throws Exception {
    // Extract the base64-encoded certificate from the XFCC header
    String base64Cert =
      certValue
        .replace("-----BEGIN CERTIFICATE-----", "")
        .replace("-----END CERTIFICATE-----", "")
        .replaceAll("\\s+", "");

    byte[] decodedCert = Base64.getDecoder().decode(base64Cert);

    CertificateFactory factory = CertificateFactory.getInstance("X.509");
    return (X509Certificate)
      factory.generateCertificate(new ByteArrayInputStream(decodedCert));
  }

  private boolean isValidOrganizationalUnit(String organizationalUnit) {
    boolean isSpaceValid = organizationalUnit.contains("space:" + validSpaceGuid);
    boolean isOrgValid = organizationalUnit.contains("org:" + validOrgGuid);
    return isSpaceValid && isOrgValid;
  }
}
