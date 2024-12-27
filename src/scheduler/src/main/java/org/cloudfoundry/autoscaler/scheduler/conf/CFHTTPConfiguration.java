package org.cloudfoundry.autoscaler.scheduler.conf;

import org.apache.catalina.connector.Connector;
import org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory;
import org.springframework.boot.web.server.WebServerFactoryCustomizer;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class CFHTTPConfiguration {

    @Bean
    public WebServerFactoryCustomizer<TomcatServletWebServerFactory> httpConnectorCustomizer() {
        return factory -> {
            Connector connector = new Connector(TomcatServletWebServerFactory.DEFAULT_PROTOCOL);
            connector.setPort(8090);
            connector.setSecure(false); // Set to false for HTTP
            factory.addAdditionalTomcatConnectors(connector);
        };
    }
}
