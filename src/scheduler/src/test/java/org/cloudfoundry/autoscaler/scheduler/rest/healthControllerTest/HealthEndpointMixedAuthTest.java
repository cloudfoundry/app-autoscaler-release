package org.cloudfoundry.autoscaler.scheduler.rest.healthControllerTest;

import static org.assertj.core.api.Assertions.assertThat;

import java.net.MalformedURLException;
import java.net.URISyntaxException;
import java.nio.charset.StandardCharsets;
import org.cloudfoundry.autoscaler.scheduler.conf.HealthServerConfiguration;
import org.cloudfoundry.autoscaler.scheduler.util.HealthUtils;
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
@TestPropertySource(
    properties = "scheduler.healthserver.unprotectedEndpoints=" + "/health/liveness")
public class HealthEndpointMixedAuthTest {
  @Autowired private TestRestTemplate restTemplate;
  @Autowired private HealthServerConfiguration healthServerConfig;

  /*
  // case 2.1 ; config ["/health/liveness"]
  here is – by configuration – one protected endpoint "/health/prometheus" and one unprotected "/health/liveness".
  The user is authenticated.
   The user queries on "/health/prometheus".
   Expected behaviour: The request will be handled successfully.
   */
  @Test
  public void givenLivenessUnprotectedAndUserIsAuthenticatedShouldReturnPrometheusWith200()
      throws MalformedURLException, URISyntaxException {

    ResponseEntity<String> prometheusResponse =
        this.restTemplate
            .withBasicAuth("prometheus", "someHash")
            .getForEntity(HealthUtils.prometheusMetricsUrl().toURI(), String.class);

    assertThat(prometheusResponse.getStatusCode().value())
        .describedAs("Http status code should be OK")
        .isEqualTo(200);
    assertThat(prometheusResponse.getHeaders().getContentType())
        .isEqualTo(new MediaType(MediaType.TEXT_PLAIN, StandardCharsets.UTF_8));
    assertThat(prometheusResponse.getBody())
        .contains("autoscaler_scheduler_data_source_initial_size 0.0");
  }
}
