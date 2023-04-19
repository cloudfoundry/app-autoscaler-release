package org.cloudfoundry.autoscaler.scheduler.conf;

import static org.junit.jupiter.api.Assertions.assertDoesNotThrow;
import static org.junit.jupiter.api.Assertions.assertThrows;

import java.util.Arrays;
import java.util.List;
import org.junit.jupiter.api.Test;
import org.junit.runner.RunWith;
import org.springframework.test.context.junit4.SpringRunner;

@RunWith(SpringRunner.class)
class HealthServerConfigurationTest {
  @Test
  void givenUnprotectedEndpointListAndUsernameOrPasswordIsNull() {
    assertThrows(
        IllegalArgumentException.class,
        () ->
            new HealthServerConfiguration(null, null, 8081, Arrays.asList("test_endpoint")).init());
  }

  @Test
  void givenUnprotectedEndpointListAndUsernameOrPasswordIsEmpty() {
    assertThrows(
        IllegalArgumentException.class,
        () -> new HealthServerConfiguration("", "", 8081, Arrays.asList("test_endpoint")).init());
  }

  @Test
  void givenEmptyUnprotectedEndpointListAndUsernameOrPasswordIsNull() {
    assertThrows(
        IllegalArgumentException.class,
        () -> new HealthServerConfiguration(null, null, 8081, List.of()).init());
  }

  @Test
  void givenEmptyUnprotectedEndpointListAndUsernameOrPasswordIsEmpty() {
    assertThrows(
        IllegalArgumentException.class,
        () -> new HealthServerConfiguration("", "", 8081, List.of()).init());
  }

  @Test
  void givenUnprotectedEndpointListIsNullThenBasicAuthRequired() {
    assertDoesNotThrow(
        () -> new HealthServerConfiguration("test-user", "test-password", 8081, null).init());
  }

  @Test
  void givenEmptyUnprotectedEndpointListWithUsernameOrPassword() {
    assertDoesNotThrow(
        () -> new HealthServerConfiguration("some-user", "some-test", 8081, List.of()).init());
  }
}
