package org.cloudfoundry.autoscaler.scheduler.rest;

import jakarta.servlet.Filter;
import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.ServletRequest;
import jakarta.servlet.ServletResponse;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import java.io.IOException;
import lombok.extern.slf4j.Slf4j;
import org.cloudfoundry.autoscaler.scheduler.conf.HealthServerConfiguration;
import org.springframework.core.annotation.Order;
import org.springframework.stereotype.Component;

@Slf4j
@Component
@Order(1)
public class MultiPortFilter implements Filter {
  HealthServerConfiguration healthServerConfiguration;

  public MultiPortFilter(HealthServerConfiguration healthServerConfiguration) {
    this.healthServerConfiguration = healthServerConfiguration;
  }

  @Override
  public void doFilter(
      ServletRequest servletRequest, ServletResponse servletResponse, FilterChain filterChain)
      throws IOException, ServletException {
    HttpServletRequest httpRequest = (HttpServletRequest) servletRequest;
    HttpServletResponse httpResponse = (HttpServletResponse) servletResponse;

    // FIXME : refactor this duplicate block
    if (isMainRequest(httpRequest)) {
      log.debug("Main server request received on port " + healthServerConfiguration.getPort());
      filterChain.doFilter(servletRequest, servletResponse);
    } else if (isHealthRequest(httpRequest)) {
      log.debug("Main server request received on port " + healthServerConfiguration.getPort());
      filterChain.doFilter(servletRequest, servletResponse);
    } else {
      httpResponse.sendError(HttpServletResponse.SC_NOT_FOUND, "Health endpoints do not exist");
    }
  }

  private boolean isHealthRequest(HttpServletRequest httpRequest) {
    return httpRequest.getLocalPort() == healthServerConfiguration.getPort()
        && httpRequest.getRequestURI().contains("health");
  }

  private boolean isMainRequest(HttpServletRequest httpRequest) {
    return httpRequest.getLocalPort() != healthServerConfiguration.getPort()
        && !httpRequest.getRequestURI().contains("health");
  }
}
