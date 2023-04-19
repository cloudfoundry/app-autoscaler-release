package org.cloudfoundry.autoscaler.scheduler.conf;

import java.util.ArrayList;
import java.util.List;
import org.apache.catalina.connector.Connector;
import org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class EmbeddedTomcatConfiguration {

  final HealthServerConfiguration healthServerConfig;

  public EmbeddedTomcatConfiguration(HealthServerConfiguration healthServerConfig) {
    this.healthServerConfig = healthServerConfig;
  }

  @Bean
  public TomcatServletWebServerFactory servletContainer() {
    TomcatServletWebServerFactory tomcat = new TomcatServletWebServerFactory();
    Connector[] additionalConnectors = this.additionalConnector();
    if (additionalConnectors != null && additionalConnectors.length > 0) {
      tomcat.addAdditionalTomcatConnectors(additionalConnectors);
    }
    return tomcat;
  }

  private Connector[] additionalConnector() {
    if (healthServerConfig.getPort() == 0) {
      return new Connector[0];
    }
    List<Connector> result = new ArrayList<>();
    Connector connector = new Connector("org.apache.coyote.http11.Http11NioProtocol");
    connector.setScheme("http");
    connector.setPort(healthServerConfig.getPort());
    result.add(connector);
    return result.toArray(new Connector[] {});
  }
}
