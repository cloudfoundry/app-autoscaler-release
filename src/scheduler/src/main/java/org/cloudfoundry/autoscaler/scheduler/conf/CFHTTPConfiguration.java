package org.cloudfoundry.autoscaler.scheduler.conf;

import org.apache.catalina.connector.Connector;
import org.apache.catalina.valves.RemoteIpValve;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory;
import org.springframework.boot.web.server.WebServerFactoryCustomizer;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
@ConfigurationProperties(prefix = "server.http")
public class CFHTTPConfiguration {
  private Logger logger = LoggerFactory.getLogger(this.getClass());

  private int port;

  public void setPort(int port) {
    this.port = port;
  }

  @Bean
  public WebServerFactoryCustomizer<TomcatServletWebServerFactory> httpConnectorCustomizer() {
    return factory -> {
      Connector connector = new Connector(TomcatServletWebServerFactory.DEFAULT_PROTOCOL);
      connector.setPort(port);
      connector.setSecure(false); // Set to false for HTTP
      //
   //   factory.addEngineValves(remoteIpValve());
      factory.addAdditionalTomcatConnectors(connector);
    };
  }

  //private RemoteIpValve remoteIpValve() {
  //  RemoteIpValve valve = new RemoteIpValve();
  //  // Specify the headers used by your proxy:
  //  // Make sure the client certificate header is set (default is "X-Forwarded-Client-Cert")
  //  // Optionally, if you need to specify which proxies to trust:
  //  // valve.setInternalProxies("127\\.0\\.0\\.1|...");
  //  valve.setRemoteIpHeader("X-Forwarded-For");
  //  valve.setProtocolHeader("X-Forwarded-Proto");
  //  valve.setRequestAttributesEnabled(true);

  //  return valve;
  //}
}
//    @Bean
//    public FilterRegistrationBean<ForwardedHeaderFilter> forwardedHeaderFilterRegistration() {
//        FilterRegistrationBean<ForwardedHeaderFilter> registrationBean = new
// FilterRegistrationBean<>();
//        registrationBean.setFilter(new ForwardedHeaderFilter());
//        registrationBean.setOrder(0); // Lower order = higher precedence
//        return registrationBean;
//    }
//
//    @Bean
//    public FilterRegistrationBean<XfccFilter> xfccFilterRegistration(XfccFilter xfccFilter) {
//        FilterRegistrationBean<XfccFilter> registrationBean = new FilterRegistrationBean<>();
//        registrationBean.setFilter(xfccFilter);
//        registrationBean.addUrlPatterns("/*"); // Apply filter to all incoming requests
//        registrationBean.setOrder(1); // Set filter precedence
//
//        logger.info("Registering XFCC Filter for CF Server");
//        return registrationBean;
//    }
