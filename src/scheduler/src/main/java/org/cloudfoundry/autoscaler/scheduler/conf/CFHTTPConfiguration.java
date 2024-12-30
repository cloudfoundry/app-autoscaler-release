package org.cloudfoundry.autoscaler.scheduler.conf;

import org.apache.catalina.connector.Connector;
import org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory;
import org.cloudfoundry.autoscaler.scheduler.filter.XfccFilter;
import org.springframework.boot.web.servlet.FilterRegistrationBean;
import org.springframework.boot.web.server.WebServerFactoryCustomizer;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class CFHTTPConfiguration {
    private Logger logger = LoggerFactory.getLogger(this.getClass());

    @Bean
    public WebServerFactoryCustomizer<TomcatServletWebServerFactory> httpConnectorCustomizer() {
        return factory -> {
            Connector connector = new Connector(TomcatServletWebServerFactory.DEFAULT_PROTOCOL);
            connector.setPort(8090);
            connector.setSecure(false); // Set to false for HTTP
            factory.addAdditionalTomcatConnectors(connector);
        };
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
