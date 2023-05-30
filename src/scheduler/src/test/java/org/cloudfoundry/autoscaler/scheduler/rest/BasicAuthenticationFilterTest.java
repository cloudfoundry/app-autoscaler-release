package org.cloudfoundry.autoscaler.scheduler.rest;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNull;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import java.io.IOException;
import java.util.Set;
import org.apache.commons.codec.binary.Base64;
import org.cloudfoundry.autoscaler.scheduler.conf.HealthServerConfiguration;
import org.junit.Before;
import org.junit.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.ValueSource;
import org.mockito.Mockito;
import org.springframework.mock.web.MockHttpServletRequest;
import org.springframework.mock.web.MockHttpServletResponse;

public class BasicAuthenticationFilterTest {
  private static final String username = "user";
  private static final String password = "pw";

  private MockHttpServletRequest req;
  private MockHttpServletResponse res;
  private FilterChain filterChainMock;

  @Before
  public void setUp() {
    req = new MockHttpServletRequest();
    res = new MockHttpServletResponse();
    filterChainMock = Mockito.mock(FilterChain.class);
    req.setLocalPort(8081);
  }

  @Test
  public void allowRequestIfPort8081andURIContainHealthWithoutUnprotectedEndpoints()
      throws IOException, ServletException {
    HealthServerConfiguration healthServerConfig =
        new HealthServerConfiguration(username, password, 8081, Set.of());

    req.setRequestURI("some/health/uri");
    String auth = username + ":" + password;
    req.addHeader("Authorization", "Basic " + Base64.encodeBase64String(auth.getBytes()));

    BasicAuthenticationFilter filter = new BasicAuthenticationFilter(healthServerConfig);

    filter.doFilter(req, res, filterChainMock);
    assertEquals(200, res.getStatus());
    Mockito.verify(filterChainMock, Mockito.times(1)).doFilter(req, res);
  }

  @Test
  public void denyHealthRequesWithAllSecuredEndpointsAndInvalidCredentials()
      throws ServletException, IOException {

    req.setRequestURI("some/health/uri");

    String auth = username + ":" + password;
    req.addHeader("Authorization", "Basic " + Base64.encodeBase64String(auth.getBytes()));
    HealthServerConfiguration healthServerConfig =
        new HealthServerConfiguration("", "", 8081, Set.of());
    BasicAuthenticationFilter userPwNullFilter = new BasicAuthenticationFilter(healthServerConfig);
    userPwNullFilter.doFilter(req, res, filterChainMock);

    Mockito.verify(filterChainMock, Mockito.times(0)).doFilter(req, res);
    assertEquals(401, res.getStatus());
    assertEquals(res.getHeader("WWW-Authenticate"), "Basic");

    res = new MockHttpServletResponse();
    healthServerConfig = new HealthServerConfiguration(username, password, 8081, Set.of());
    BasicAuthenticationFilter wrongCredsFilter = new BasicAuthenticationFilter(healthServerConfig);
    req.removeHeader("Authorization");
    String wrongCreds = "wrong-user:pw";
    req.addHeader("Authorization", "Basic " + Base64.encodeBase64String(wrongCreds.getBytes()));
    wrongCredsFilter.doFilter(req, res, filterChainMock);

    Mockito.verify(filterChainMock, Mockito.times(0)).doFilter(req, res);
    assertEquals(401, res.getStatus());
    assertEquals(res.getHeader("WWW-Authenticate"), "Basic");

    res = new MockHttpServletResponse();
    healthServerConfig = new HealthServerConfiguration(username, password, 8081, Set.of());
    BasicAuthenticationFilter noAuthHeaderFilter =
        new BasicAuthenticationFilter(healthServerConfig);
    req.removeHeader("Authorization");
    noAuthHeaderFilter.doFilter(req, res, filterChainMock);

    Mockito.verify(filterChainMock, Mockito.times(0)).doFilter(req, res);
    assertEquals(401, res.getStatus());
    assertEquals(res.getHeader("WWW-Authenticate"), "Basic");
  }

  @Test
  public void denyHealthRequestIfAuthHeaderIsInvalid() throws IOException, ServletException {

    req.setRequestURI("some/health/uri");

    HealthServerConfiguration healthServerConfig =
        new HealthServerConfiguration(username, password, 8081, Set.of());
    BasicAuthenticationFilter malformedHeaderFilter =
        new BasicAuthenticationFilter(healthServerConfig);
    req.removeHeader("Authorization");
    String malformedCreds = "some-malformed-creds";
    req.addHeader("Authorization", "Basic " + Base64.encodeBase64String(malformedCreds.getBytes()));
    malformedHeaderFilter.doFilter(req, res, filterChainMock);

    assertEquals(400, res.getStatus());
    assertNull(res.getHeader("WWW-Authenticate"));
  }

  @Test
  public void allowRequestIfPort8081andURIContainHealthWithUnprotectedEndpoints()
      throws IOException, ServletException {

    HealthServerConfiguration healthServerConfig =
        new HealthServerConfiguration(username, password, 8081, Set.of("/health/liveness"));
    req.setRequestURI("/health/liveness");
    String auth = username + ":" + password;
    req.addHeader("Authorization", "Basic " + Base64.encodeBase64String(auth.getBytes()));

    BasicAuthenticationFilter filter = new BasicAuthenticationFilter(healthServerConfig);
    filter.doFilter(req, res, filterChainMock);

    Mockito.verify(filterChainMock, Mockito.times(1)).doFilter(req, res);
  }

  @Test
  public void denyHealthRequestWithWrongUnprotectedEndpoints()
      throws IOException, ServletException {

    req.setRequestURI("/health/liveness");
    String auth = username + ":" + password;
    req.addHeader("Authorization", "Basic " + Base64.encodeBase64String(auth.getBytes()));

    HealthServerConfiguration healthServerConfig =
        new HealthServerConfiguration(username, password, 8081, Set.of("/health/wrong-endpoint"));
    BasicAuthenticationFilter filter = new BasicAuthenticationFilter(healthServerConfig);
    filter.doFilter(req, res, filterChainMock);

    Mockito.verify(filterChainMock, Mockito.times(0)).doFilter(req, res);
    assertEquals(401, res.getStatus());
    assertEquals(res.getHeader("WWW-Authenticate"), "Basic");
  }

  @ParameterizedTest
  @ValueSource(strings = {"/health/prometheus", "/health/liveness", "/health/wrong-endpoint"})
  public void denyHealthRequestsWithNoUnprotectedEndpointsConfigThenBasicAuthRequired(
      String requestURI) throws IOException, ServletException {
    req = new MockHttpServletRequest();
    res = new MockHttpServletResponse();
    filterChainMock = Mockito.mock(FilterChain.class);
    req.setLocalPort(8081);

    req.setRequestURI(requestURI);

    HealthServerConfiguration healthServerConfig =
        new HealthServerConfiguration(username, password, 8081, Set.of());
    BasicAuthenticationFilter basicAuthenticationFilter =
        new BasicAuthenticationFilter(healthServerConfig);

    basicAuthenticationFilter.doFilter(req, res, filterChainMock);

    Mockito.verify(filterChainMock, Mockito.times(0)).doFilter(req, res);
    assertEquals(401, res.getStatus());
    assertEquals(res.getHeader("WWW-Authenticate"), "Basic");
  }

  @Test
  public void denyHealthRequestIfBasicAuthRequired() throws IOException, ServletException {

    HealthServerConfiguration healthServerConfig =
        new HealthServerConfiguration(username, password, 8081, Set.of("/health/prometheus"));
    req.setRequestURI("/health/liveness");

    BasicAuthenticationFilter noBasicAuthFilter = new BasicAuthenticationFilter(healthServerConfig);
    noBasicAuthFilter.doFilter(req, res, filterChainMock);

    Mockito.verify(filterChainMock, Mockito.times(0)).doFilter(req, res);
    assertEquals(401, res.getStatus());
    assertEquals(res.getHeader("WWW-Authenticate"), "Basic");
  }

  @Test
  public void basicAuthFilterNotAppliedIfNoHealthRequest() throws IOException, ServletException {
    req.setLocalPort(8080);
    req.setRequestURI("/routeTo8080");

    HealthServerConfiguration healthServerConfig =
        new HealthServerConfiguration(username, password, 8081, Set.of("/health/wrong-endpoint"));
    BasicAuthenticationFilter filter = new BasicAuthenticationFilter(healthServerConfig);
    filter.doFilter(req, res, filterChainMock);

    Mockito.verify(filterChainMock, Mockito.times(1)).doFilter(req, res);
    assertEquals(200, res.getStatus());
    assertEquals(8080, req.getLocalPort());
  }
}
