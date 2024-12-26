package org.cloudfoundry.autoscaler.scheduler.conf;

import java.io.IOException;
import java.util.logging.Logger;
import org.apache.http.impl.bootstrap.HttpServer;
import org.apache.http.impl.bootstrap.ServerBootstrap;
import org.apache.http.protocol.HttpRequestHandler;
import org.cloudfoundry.autoscaler.scheduler.filter.XfccFilter;
import org.springframework.boot.web.servlet.FilterRegistrationBean;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.apache.http.entity.StringEntity;

@Configuration
public class CFServerConfig {
  private Logger logger = Logger.getLogger(this.getClass().getName());

  HttpServer cfServer(CFServerConfiguration config) throws IOException {
    // Define a simple request handler (example only)
    HttpRequestHandler requestHandler =
        (request, response, context) -> {
          response.setStatusCode(200);
          response.setEntity(new StringEntity("Hello from CFServer!"));
        };

        // Build the HTTP server
        HttpServer server = ServerBootstrap.bootstrap()
                .setListenerPort(config.getPort())
                .registerHandler("*", requestHandler) // Register a default handler
                .create();

        logger.info("Configured HttpServer on port: " + config.getPort());
        return server;
  }

    @Bean
    public FilterRegistrationBean<XfccFilter> xfccFilterRegistration(XfccFilter xfccFilter) {
        FilterRegistrationBean<XfccFilter> registrationBean = new FilterRegistrationBean<>();
        registrationBean.setFilter(xfccFilter);
        registrationBean.addUrlPatterns("/*"); // Apply filter to all incoming requests
        registrationBean.setOrder(1); // Set filter precedence

        logger.info("Registering XFCC Filter for CF Server");
        return registrationBean;
    }
}

