package org.cloudfoundry.autoscaler.scheduler.conf;

import java.io.File;
import java.io.FileInputStream;
import java.security.KeyStore;
import javax.net.ssl.SSLContext;
import org.apache.hc.client5.http.classic.HttpClient;
import org.apache.hc.client5.http.config.RequestConfig;
import org.apache.hc.client5.http.impl.classic.HttpClientBuilder;
import org.apache.hc.client5.http.impl.io.PoolingHttpClientConnectionManagerBuilder;
import org.apache.hc.client5.http.io.HttpClientConnectionManager;
import org.apache.hc.client5.http.ssl.HttpsSupport;
import org.apache.hc.client5.http.ssl.SSLConnectionSocketFactory;
import org.apache.hc.core5.ssl.SSLContextBuilder;
import org.apache.hc.core5.ssl.SSLContexts;
import org.apache.hc.core5.util.Timeout;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.http.client.ClientHttpRequestFactory;
import org.springframework.http.client.HttpComponentsClientHttpRequestFactory;
import org.springframework.web.client.RestOperations;
import org.springframework.web.client.RestTemplate;

@Configuration
public class RestClientConfig {
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
      @Value("${client.ssl.key-store}") String keyStoreFile,
      @Value("${client.ssl.key-store-password}") String keyStorePassword,
      @Value("${client.ssl.key-store-type}") String keyStoreType,
      @Value("${client.ssl.trust-store}") String trustStoreFile,
      @Value("${client.ssl.trust-store-password}") String trustStorePassword,
      @Value("${client.ssl.protocol}") String protocol,
      @Value("${client.httpClientTimeout}") Integer httpClientTimeout)
      throws Exception {
    KeyStore trustStore = KeyStore.getInstance(KeyStore.getDefaultType());
    KeyStore keyStore =
        KeyStore.getInstance(keyStoreType == null ? KeyStore.getDefaultType() : keyStoreType);

    try (FileInputStream trustStoreInstream = new FileInputStream(new File(trustStoreFile));
        FileInputStream keyStoreInstream = new FileInputStream(new File(keyStoreFile))) {
      trustStore.load(trustStoreInstream, trustStorePassword.toCharArray());
      keyStore.load(keyStoreInstream, keyStorePassword.toCharArray());
    }

    SSLContextBuilder sslCtxBuilder = SSLContexts.custom().loadTrustMaterial(trustStore, null);
    sslCtxBuilder = sslCtxBuilder.loadKeyMaterial(keyStore, keyStorePassword.toCharArray());

    SSLContext sslcontext = sslCtxBuilder.build();

    HttpClientBuilder builder = HttpClientBuilder.create();
    SSLConnectionSocketFactory sslsf =
        new SSLConnectionSocketFactory(
            sslcontext, new String[] {protocol}, null, HttpsSupport.getDefaultHostnameVerifier());

    HttpClientConnectionManager ccm =
        PoolingHttpClientConnectionManagerBuilder.create().setSSLSocketFactory(sslsf).build();
    builder.setConnectionManager(ccm);
    RequestConfig requestConfig =
        RequestConfig.custom()
            .setConnectTimeout(Timeout.ofSeconds(httpClientTimeout))
            .setConnectionRequestTimeout(Timeout.ofSeconds(httpClientTimeout))
            .setResponseTimeout(Timeout.ofSeconds(httpClientTimeout))
            .build();
    builder.setDefaultRequestConfig(requestConfig);
    return builder.build();
  }
}
