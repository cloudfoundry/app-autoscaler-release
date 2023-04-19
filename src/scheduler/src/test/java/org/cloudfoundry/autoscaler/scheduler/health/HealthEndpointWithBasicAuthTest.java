package org.cloudfoundry.autoscaler.scheduler.health;

import static org.assertj.core.api.Assertions.assertThat;

import org.cloudfoundry.autoscaler.scheduler.conf.HealthServerConfiguration;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.context.SpringBootTest.WebEnvironment;
import org.springframework.boot.test.web.client.TestRestTemplate;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.test.annotation.DirtiesContext;
import org.springframework.test.annotation.DirtiesContext.ClassMode;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.test.context.TestPropertySource;
import org.springframework.test.context.junit4.SpringRunner;

@RunWith(SpringRunner.class)
@SpringBootTest(webEnvironment = WebEnvironment.RANDOM_PORT)
@DirtiesContext(classMode = ClassMode.BEFORE_CLASS)
@ActiveProfiles("HealthAuth")
@TestPropertySource(properties = "scheduler.healthserver.unprotectedEndpoints=")
public class HealthEndpointWithBasicAuthTest {

  @Autowired private TestRestTemplate restTemplate;

  @Autowired private HealthServerConfiguration healthServerConfig;

  @Test
  public void givenCorrectCredentialsPrometheusShouldBeAvailable() {

    ResponseEntity<String> response =
        this.restTemplate
            .withBasicAuth("prometheus", "someHash")
            .getForEntity(prometheusMetricsUrl(), String.class);
    assertThat(response.getStatusCode().value())
        .describedAs("Http status code should be OK")
        .isEqualTo(200);
    String result = response.toString();
    assertThat(result)
        .contains("jvm_info")
        .contains("jvm_buffer_pool_used_bytes")
        .contains("jvm_buffer_pool_capacity_bytes")
        .contains("jvm_buffer_pool_used_buffers")
        .contains("jvm_gc_collection_seconds_count")
        .contains("jvm_gc_collection_seconds_sum")
        .contains("jvm_classes_loaded")
        .contains("jvm_classes_loaded_total")
        .contains("jvm_classes_unloaded_total")
        .contains("jvm_threads")
        .contains("jvm_memory_bytes")
        .contains("jvm_memory_pool_bytes")
        .contains("autoscaler_scheduler_data_source")
        .contains("autoscaler_scheduler_policy_db_data_source");
  }

  @Test
  public void givenRootEndpointThenMetricsShouldNotBeAvailable() {
    ResponseEntity<String> response = this.restTemplate.getForEntity("/", String.class);
    assertThat(response.getStatusCode().value())
        .describedAs("Http status code should be Not Found")
        .isEqualTo(404);
  }

  @Test
  public void givenIncorrectCredentialsShouldReturn401() {
    ResponseEntity<String> response =
        this.restTemplate
            .withBasicAuth("bad", "auth")
            .getForEntity(prometheusMetricsUrl(), String.class);
    assertThat(response.getStatusCode().value())
        .describedAs("Http status code should be Unauthorized")
        .isEqualTo(401);
  }

  @Test
  public void givenNoCredentialsShouldReturn401() {
    ResponseEntity<String> response =
        this.restTemplate.getForEntity(prometheusMetricsUrl(), String.class);
    assertThat(response.getStatusCode().value()).isEqualTo(401);
  }

  @Test
  public void givenCorrectPasswordAndWrongUsernameFailsWith401() {
    ResponseEntity<String> response =
        this.restTemplate
            .withBasicAuth("bad", "someHash")
            .getForEntity(prometheusMetricsUrl(), String.class);
    assertThat(response.getStatusCode().value()).isEqualTo(401);
  }

  @Test
  public void givenCorrectCredentialsLivenessShouldBeAvailable() {

    ResponseEntity<String> response =
        this.restTemplate
            .withBasicAuth("prometheus", "someHash")
            .getForEntity(livenessUrl(), String.class);
    assertThat(response.getStatusCode().value())
        .describedAs("Http status code should be OK")
        .isEqualTo(200);
    assertThat(response.getHeaders().getContentType()).isEqualTo(MediaType.APPLICATION_JSON);
    assertThat(response.getBody()).isEqualTo("{\"status\":\"Up\"}");
  }

  @Test
  public void givenNoCredentialsLivenessShouldReturn401() {

    ResponseEntity<String> response = this.restTemplate.getForEntity(livenessUrl(), String.class);
    assertThat(response.getStatusCode().value())
        .describedAs("Http status code should be Unauthorized")
        .isEqualTo(401);
  }

  @Test
  public void givenIncorrectCredentialsShouldLivenessReturn401() {
    ResponseEntity<String> response =
        this.restTemplate.withBasicAuth("bad", "auth").getForEntity(livenessUrl(), String.class);
    assertThat(response.getStatusCode().value())
        .describedAs("Http status code should be Unauthorized")
        .isEqualTo(401);
  }

  private String prometheusMetricsUrl() {
    return "http://localhost:" + healthServerConfig.getPort() + "/health/prometheus";
  }

  private String livenessUrl() {
    return "http://localhost:" + healthServerConfig.getPort() + "/health/liveness";
  }
}
