package org.cloudfoundry.autoscaler.scheduler.conf;

import javax.net.ssl.SSLContext;
import org.apache.hc.client5.http.classic.HttpClient;
import org.apache.hc.client5.http.config.ConnectionConfig;
import org.apache.hc.client5.http.config.RequestConfig;
import org.apache.hc.client5.http.impl.classic.HttpClientBuilder;
import org.apache.hc.client5.http.impl.io.PoolingHttpClientConnectionManagerBuilder;
import org.apache.hc.client5.http.io.HttpClientConnectionManager;
import org.apache.hc.client5.http.ssl.HttpsSupport;
import org.apache.hc.client5.http.ssl.SSLConnectionSocketFactory;
import org.apache.hc.core5.util.Timeout;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.ssl.SslBundle;
import org.springframework.boot.ssl.SslBundles;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.http.client.ClientHttpRequestFactory;
import org.springframework.http.client.HttpComponentsClientHttpRequestFactory;
import org.springframework.web.client.RestOperations;
import org.springframework.web.client.RestTemplate;

@Configuration
public class RestClientConfig {
  private final SSLContext sslContext;
  private final boolean sslEnabled;

  @Autowired
  public RestClientConfig(SslBundles sslBundles, 
                         @Value("${client.ssl.enabled:true}") boolean sslEnabled) {
    this.sslEnabled = sslEnabled;
    if (sslEnabled) {
      SslBundle sslBundle = sslBundles.getBundle("scalingengine");
      this.sslContext = sslBundle.createSslContext();
    } else {
      this.sslContext = null;
    }
  }

  @Bean
  public RestOperations restOperations(ClientHttpRequestFactory clientHttpRequestFactory)
      throws Exception {
    return new RestTemplate(clientHttpRequestFactory);
  }

  @Bean
  public ClientHttpRequestFactory clientHttpRequestFactory(HttpClient httpClient) {
    return new HttpComponentsClientHttpRequestFactory(httpClient);
  }

  @Bean
  public HttpClient httpClient(
      @Value("${client.ssl.protocol}") String protocol,
      @Value("${client.httpClientTimeout}") Integer httpClientTimeout)
      throws Exception {

    HttpClientBuilder builder = HttpClientBuilder.create();
    
    var connectionConfig =
        ConnectionConfig.custom().setConnectTimeout(Timeout.ofSeconds(httpClientTimeout)).build();
    
    if (sslEnabled && this.sslContext != null) {
      SSLConnectionSocketFactory sslsf =
          new SSLConnectionSocketFactory(
              this.sslContext,
              new String[] {protocol},
              null,
              HttpsSupport.getDefaultHostnameVerifier());

      HttpClientConnectionManager ccm =
          PoolingHttpClientConnectionManagerBuilder.create()
              .setSSLSocketFactory(sslsf)
              .setDefaultConnectionConfig(connectionConfig)
              .build();
      builder.setConnectionManager(ccm);
    } else {
      HttpClientConnectionManager ccm =
          PoolingHttpClientConnectionManagerBuilder.create()
              .setDefaultConnectionConfig(connectionConfig)
              .build();
      builder.setConnectionManager(ccm);
    }
    
    RequestConfig requestConfig =
        RequestConfig.custom()
            .setConnectionRequestTimeout(Timeout.ofSeconds(httpClientTimeout))
            .setResponseTimeout(Timeout.ofSeconds(httpClientTimeout))
            .build();
    builder.setDefaultRequestConfig(requestConfig);
    return builder.build();
  }
}
