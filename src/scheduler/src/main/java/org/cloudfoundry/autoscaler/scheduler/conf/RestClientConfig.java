package org.cloudfoundry.autoscaler.scheduler.conf;

import java.time.Duration;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.web.client.RestTemplateBuilder;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.client.RestOperations;

@Configuration
public class RestClientConfig {
  @Value("${autoscaler.scalingengine.basic_auth.password}")
  private String scalingEnginePassword;

  @Value("${autoscaler.scalingengine.basic_auth.username}")
  private String scalingEngineUsername;

  @Value("${client.httpClientTimeout}")
  private Integer httpClientTimeout ;

  @Bean
  public RestOperations restOperations(RestTemplateBuilder restTemplateBuilder) {
    return restTemplateBuilder
      .setConnectTimeout( Duration.ofSeconds(httpClientTimeout))
      .setReadTimeout( Duration.ofSeconds(httpClientTimeout))
      .basicAuthentication(scalingEngineUsername, scalingEnginePassword).build();
  }
}
