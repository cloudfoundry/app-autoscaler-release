package org.cloudfoundry.autoscaler.scheduler.filter;

import jakarta.annotation.PostConstruct;
import jakarta.servlet.FilterChain;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.security.cert.CertificateFactory;
import java.security.cert.X509Certificate;
import java.util.Base64;
import lombok.Data;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.core.annotation.Order;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

@Component
@Order(0)
@Data
@ConfigurationProperties(prefix = "cfserver")
public class XfccFilter extends OncePerRequestFilter {
  private Logger logger = LoggerFactory.getLogger(this.getClass());

  private String validSpaceGuid;
  private String validOrgGuid;

  @Override
  protected void doFilterInternal(
      HttpServletRequest request, HttpServletResponse response, FilterChain filterChain)
      throws jakarta.servlet.ServletException, IOException {

    String path = request.getRequestURI();
    String method = request.getMethod();

    logger.info("Received " + method + " Request to " + path);
    // Skip filter if the request is HTTPS
    if (request.isSecure()) {
      filterChain.doFilter(request, response);
      return;
    }

    String xfccHeader = request.getHeader("X-Forwarded-Client-Cert");
    if (xfccHeader == null || xfccHeader.isEmpty()) {
      logger.warn("Missing X-Forwarded-Client-Cert header");
      response.sendError(
          HttpServletResponse.SC_BAD_REQUEST, "Missing X-Forwarded-Client-Cert header");
      return;
    } else {
      logger.info(
          "X-Forwarded-Client-Cert header received ... checking authorized org and space in OU");
    }

    // xfccHeader in semicolom separted Cert=some-cert;Hash=some-has with regular expresion
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
      // Decode and parse the certificate
      X509Certificate certificate = parseCertificate(certValue);

      // Extract the OrganizationalUnit
      String organizationalUnit = certificate.getSubjectX500Principal().getName();

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

  private X509Certificate parseCertificate(String xfccHeader) throws Exception {
    // Extract the base64-encoded certificate from the XFCC header
    String base64Cert =
        xfccHeader
            .replace("-----BEGIN CERTIFICATE-----", "")
            .replace("-----END CERTIFICATE-----", "")
            .replaceAll("\\s+", "");

    byte[] decodedCert = Base64.getDecoder().decode(base64Cert);

    CertificateFactory factory = CertificateFactory.getInstance("X.509");
    return (X509Certificate)
        factory.generateCertificate(new java.io.ByteArrayInputStream(decodedCert));
  }

  private boolean isValidOrganizationalUnit(String organizationalUnit) {
    boolean isSpaceValid = organizationalUnit.contains("space:" + validSpaceGuid);
    boolean isOrgValid = organizationalUnit.contains("org:" + validOrgGuid);
    return isSpaceValid && isOrgValid;
  }
}
