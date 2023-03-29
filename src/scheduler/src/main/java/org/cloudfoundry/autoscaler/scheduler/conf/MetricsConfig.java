package org.cloudfoundry.autoscaler.scheduler.conf;

import com.sun.net.httpserver.BasicAuthenticator;
import com.sun.net.httpserver.HttpContext;
import com.sun.net.httpserver.HttpServer;
import io.prometheus.client.CollectorRegistry;
import io.prometheus.client.exporter.HTTPServer;
import io.prometheus.client.exporter.HTTPServer.Builder;
import io.prometheus.client.exporter.common.TextFormat;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.configurationprocessor.json.JSONException;
import org.springframework.boot.configurationprocessor.json.JSONObject;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import java.io.IOException;
import java.io.StringWriter;
import java.net.InetSocketAddress;
import java.util.Collections;

/**
 * HTTP Server running on port {@link org/cloudfoundry/autoscaler/scheduler/conf/MetricsConfiguration#getPort() config.getPort()}
 * Used for prometheus and other health checks e.g /health/liveness
 */
@Configuration
public class MetricsConfig {

    private final Logger logger =LoggerFactory.getLogger(this.getClass());

  @Bean(destroyMethod = "close")
  public HTTPServer metricsServer(MetricsConfiguration config) throws IOException {
      logger.debug("prometheus server starting...");
      InetSocketAddress address = new InetSocketAddress(config.getPort());
      HttpServer httpServer = HttpServer.create(address, 0);


      HttpContext context = httpServer.createContext("/health/prometheus", httpExchange -> {
          String method = httpExchange.getRequestMethod();
          String path = httpExchange.getRequestURI().getPath();
          if (! "GET".equalsIgnoreCase(method)) {
              httpExchange.getResponseHeaders().set("Allow", "GET");
              httpExchange.sendResponseHeaders(405, -1);
              httpExchange.getResponseBody().close();
              return;
          }

          StringWriter respBodyWriter = new StringWriter();
          TextFormat.write004(respBodyWriter, CollectorRegistry.defaultRegistry.metricFamilySamples());
          byte[] respBody = respBodyWriter.toString().getBytes("UTF-8");
          httpExchange.getResponseHeaders().put("Context-Type", Collections.singletonList("text/plain; charset=UTF-8"));
          httpExchange.sendResponseHeaders(201, respBody.length);
          httpExchange.getResponseBody().write(respBody);
          httpExchange.getResponseBody().close();
      });

      httpServer.createContext("/health/liveness", httpExchange -> {
          String method = httpExchange.getRequestMethod();
          String path = httpExchange.getRequestURI().getPath();

          if (! "GET".equalsIgnoreCase(method)) {
              httpExchange.getResponseHeaders().set("Allow", "GET");
              httpExchange.sendResponseHeaders(405, -1);
              httpExchange.getResponseBody().close();
              return;
          }
          JSONObject jsonObject = new JSONObject();
          try {
              jsonObject.put("status", "Up");
          } catch (JSONException e) {
              throw new RuntimeException(e);
          }
          logger.info("json  "+jsonObject);
;
          httpExchange.getResponseHeaders().add( "Content-type", "application/json");
          httpExchange.getResponseHeaders().add("Content-length", Integer.toString(jsonObject.length()));

          httpExchange.sendResponseHeaders(200, 0);
          httpExchange.getResponseBody().write(jsonObject.toString().getBytes());
          httpExchange.getResponseBody().close();
      });

      httpServer.createContext("/", httpExchange -> {
          JSONObject jsonObject = new JSONObject();
          try {
              jsonObject.put("status", "Not Found");
          } catch (JSONException e) {
              throw new RuntimeException(e);
          }
          logger.info("json  "+jsonObject);
          ;
          httpExchange.getResponseHeaders().put( "Content-type", Collections.singletonList("application/json"));
          httpExchange.getResponseHeaders().add("Content-length", Integer.toString(jsonObject.length()));

          httpExchange.sendResponseHeaders(404, 0);
          httpExchange.getResponseBody().write(jsonObject.toString().getBytes());
          httpExchange.getResponseBody().close();
      });

      Builder builder = new Builder()
              .withHttpServer(httpServer);

    if (config.isBasicAuthEnabled()) {
      builder.withAuthenticator(
          new BasicAuthenticator("/health/prometheus") {
            @Override
            public boolean checkCredentials(String username, String password) {
              return config.getUsername().equals(username) && config.getPassword().equals(password);
            }
          });
    }
    HTTPServer httpServerBuild = builder.build();
    logger.info("prometheus server started on port "+httpServerBuild.getPort());

    return  httpServerBuild;
  }

}
