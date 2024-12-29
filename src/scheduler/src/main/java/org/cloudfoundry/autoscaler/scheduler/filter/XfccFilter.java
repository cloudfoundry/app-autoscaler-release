package org.cloudfoundry.autoscaler.scheduler.filter;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;


import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.security.cert.CertificateFactory;
import java.security.cert.X509Certificate;
import java.util.Base64;

@Component
public class XfccFilter extends OncePerRequestFilter {
    @Value("${cfserver.validSpaceGuid}")
    private String validSpaceGuid;

    @Value("${cfserver.validOrgGuid}")
    private String validOrgGuid;

    @Override
    protected void doFilterInternal(HttpServletRequest request,
        HttpServletResponse response, FilterChain filterChain)
        throws jakarta.servlet.ServletException, IOException {


        // Skip filter if the request is HTTPS
        if (request.isSecure()) {
          filterChain.doFilter(request, response);
          return;
        }

        // Get the XFCC header
        String xfccHeader = request.getHeader("X-Forwarded-Client-Cert");
        if (xfccHeader == null || xfccHeader.isEmpty()) {
            response.sendError(HttpServletResponse.SC_BAD_REQUEST, "Missing X-Forwarded-Client-Cert header");
            return;
        }

        try {
            // Decode and parse the certificate
            X509Certificate certificate = parseCertificate(xfccHeader);

            // Extract the OrganizationalUnit
            String organizationalUnit = certificate.getSubjectX500Principal().getName();

            // Validate both key-value pairs in OrganizationalUnit
            if (!isValidOrganizationalUnit(organizationalUnit)) {
                response.sendError(HttpServletResponse.SC_FORBIDDEN, "Unauthorized OrganizationalUnit");
                return;
            }

        } catch (Exception e) {
            response.sendError(HttpServletResponse.SC_BAD_REQUEST, "Invalid certificate: " + e.getMessage());
            return;
        }

        // Proceed with the request
        filterChain.doFilter(request, response);
    }

    private X509Certificate parseCertificate(String xfccHeader) throws Exception {
        // Extract the base64-encoded certificate from the XFCC header
        String base64Cert = xfccHeader.replace("-----BEGIN CERTIFICATE-----", "")
                                      .replace("-----END CERTIFICATE-----", "")
                                      .replaceAll("\\s+", "");

        byte[] decodedCert = Base64.getDecoder().decode(base64Cert);

        CertificateFactory factory = CertificateFactory.getInstance("X.509");
        return (X509Certificate) factory.generateCertificate(new java.io.ByteArrayInputStream(decodedCert));
    }

    private boolean isValidOrganizationalUnit(String organizationalUnit) {
        // Check if both key-value pairs are present and valid
        boolean isSpaceValid = organizationalUnit.contains("space=" + validSpaceGuid);
        boolean isOrgValid = organizationalUnit.contains("org=" + validOrgGuid);
        return isSpaceValid && isOrgValid;
    }
}

