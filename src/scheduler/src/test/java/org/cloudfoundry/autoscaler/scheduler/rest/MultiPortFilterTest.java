package org.cloudfoundry.autoscaler.scheduler.rest;

import java.io.IOException;
import jakarta.servlet.*;
import java.util.Arrays;

import org.cloudfoundry.autoscaler.scheduler.conf.HealthServerConfiguration;

import org.junit.Test;
import org.junit.runner.RunWith;
import org.mockito.Mockito;
import org.springframework.test.context.junit4.SpringRunner;
import org.springframework.mock.web.MockHttpServletRequest;
import org.springframework.mock.web.MockHttpServletResponse;
import static org.junit.jupiter.api.Assertions.assertEquals;

import org.apache.commons.codec.binary.Base64;


@RunWith(SpringRunner.class)
public class MultiPortFilterTest {
  String username = "user";
  String password = "pw";

  @Test
  public void denyRequestIfPortNot8081orURIDoesNotContainHealth() throws IOException, ServletException {
    MockHttpServletRequest req = new MockHttpServletRequest();
    MockHttpServletResponse res = new MockHttpServletResponse();

    MultiPortFilter filter = new MultiPortFilter(null);
    filter.doFilter(req, res, null);
    assertEquals(80, req.getLocalPort());
    assertEquals(404, res.getStatus());    
  }

  @Test
  public void allowRequestIfPort8081andURIContainHealthWithoutUnprotectedEndpoints() throws IOException, ServletException {
    HealthServerConfiguration healthServerConfig = new HealthServerConfiguration(username, password, 8081, Arrays.asList());

    MockHttpServletRequest req = new MockHttpServletRequest();
    FilterChain filterChainMock = Mockito.mock(FilterChain.class);
    req.setLocalPort(8081);
    req.setRequestURI("some/health/uri");
    String auth = username + ":" + password;
    req.addHeader("Authorization", "Basic " + Base64.encodeBase64String(auth.getBytes()));//Base64.encodeBase64(auth.getBytes()));
    MockHttpServletResponse res = new MockHttpServletResponse();

    MultiPortFilter filter = new MultiPortFilter(healthServerConfig);
        
    filter.doFilter(req, res, filterChainMock);
    assertEquals(200, res.getStatus());
    Mockito.verify(filterChainMock, Mockito.times(1)).doFilter(req, res);
  }

  @Test
  public void denyRequestIfPort8081andURIContainHealthWithoutUnprotectedEndpointsUserNotAuthenticated() throws IOException, ServletException {
    MockHttpServletRequest req = new MockHttpServletRequest();
    FilterChain filterChainMock = Mockito.mock(FilterChain.class);
    req.setLocalPort(8081);
    req.setRequestURI("some/health/uri");
    String auth = username + ":" + password;
    req.addHeader("Authorization", "Basic " + Base64.encodeBase64String(auth.getBytes()));//Base64.encodeBase64(auth.getBytes()));
    MockHttpServletResponse res = new MockHttpServletResponse();

    HealthServerConfiguration healthServerConfig = new HealthServerConfiguration(null, null, 8081, Arrays.asList());
    MultiPortFilter userPwNullFilter = new MultiPortFilter(healthServerConfig);
    userPwNullFilter.doFilter(req, res, filterChainMock);
    assertEquals(401, res.getStatus());
    Mockito.verify(filterChainMock, Mockito.times(0)).doFilter(req, res);
    assertEquals(res.getHeader("WWW-Authenticate"), "Basic");

    res = new MockHttpServletResponse();
    healthServerConfig = new HealthServerConfiguration(username, password, 8081, Arrays.asList());
    MultiPortFilter malformedHeaderFilter = new MultiPortFilter(healthServerConfig);
    req.removeHeader("Authorization");
    String malformedCreds = "some-malformed-creds";
    req.addHeader("Authorization", "Basic " + Base64.encodeBase64String(malformedCreds.getBytes()));//Base64.encodeBase64(auth.getBytes()));
    malformedHeaderFilter.doFilter(req, res, filterChainMock);
    assertEquals(400, res.getStatus());
    Mockito.verify(filterChainMock, Mockito.times(0)).doFilter(req, res);
    assertEquals(res.getHeader("WWW-Authenticate"), null);

    res = new MockHttpServletResponse();
    healthServerConfig = new HealthServerConfiguration(username, password, 8081, Arrays.asList());
    MultiPortFilter wrongCredsFilter = new MultiPortFilter(healthServerConfig);
    req.removeHeader("Authorization");
    String wrongCreds = "wrong-user:pw";
    req.addHeader("Authorization", "Basic " + Base64.encodeBase64String(wrongCreds.getBytes()));//Base64.encodeBase64(auth.getBytes()));
    wrongCredsFilter.doFilter(req, res, filterChainMock);
    assertEquals(401, res.getStatus());
    Mockito.verify(filterChainMock, Mockito.times(0)).doFilter(req, res);
    assertEquals(res.getHeader("WWW-Authenticate"), "Basic");

    res = new MockHttpServletResponse();
    healthServerConfig = new HealthServerConfiguration(username, password, 8081, Arrays.asList());
    MultiPortFilter noAuthHeaderFilter = new MultiPortFilter(healthServerConfig);
    req.removeHeader("Authorization");
    noAuthHeaderFilter.doFilter(req, res, filterChainMock);
    assertEquals(401, res.getStatus());
    Mockito.verify(filterChainMock, Mockito.times(0)).doFilter(req, res);
    assertEquals(res.getHeader("WWW-Authenticate"), "Basic");
  }

  @Test
  public void allowRequestIfPort8081andURIContainHealthWithUnprotectedEndpoints() throws IOException, ServletException {
    HealthServerConfiguration healthServerConfig = new HealthServerConfiguration(username, password, 8081, Arrays.asList("/health/liveness"));

    MockHttpServletRequest req = new MockHttpServletRequest();
    FilterChain filterChainMock = Mockito.mock(FilterChain.class);
    req.setLocalPort(8081);
    req.setRequestURI("/health/liveness");
    String auth = username + ":" + password;
    req.addHeader("Authorization", "Basic " + Base64.encodeBase64String(auth.getBytes()));//Base64.encodeBase64(auth.getBytes()));
    MockHttpServletResponse res = new MockHttpServletResponse();

    MultiPortFilter filter = new MultiPortFilter(healthServerConfig);
        
    filter.doFilter(req, res, filterChainMock);

    Mockito.verify(filterChainMock, Mockito.times(1)).doFilter(req, res);
  }

  @Test
  public void denyRequestIfPort8081andURIContainHealthWithUnprotectedEndpoints() throws IOException, ServletException {
    MockHttpServletRequest req = new MockHttpServletRequest();
    FilterChain filterChainMock = Mockito.mock(FilterChain.class);
    req.setLocalPort(8081);
    req.setRequestURI("/health/liveness");
    String auth = username + ":" + password;
    req.addHeader("Authorization", "Basic " + Base64.encodeBase64String(auth.getBytes()));//Base64.encodeBase64(auth.getBytes()));
    MockHttpServletResponse res = new MockHttpServletResponse();

    
    HealthServerConfiguration healthServerConfig = new HealthServerConfiguration(username, password, 8081, Arrays.asList("/health/wrong-endpoint"));
    MultiPortFilter filter = new MultiPortFilter(healthServerConfig);
        
    filter.doFilter(req, res, filterChainMock);

    Mockito.verify(filterChainMock, Mockito.times(0)).doFilter(req, res);
    assertEquals(401, res.getStatus());
    assertEquals(res.getHeader("WWW-Authenticate"), "Basic");

    healthServerConfig = new HealthServerConfiguration(username, password, 8081, Arrays.asList("/health/liveness"));
    filter = new MultiPortFilter(healthServerConfig);
    req.setRequestURI("/health/wrong-endpoint");
    res = new MockHttpServletResponse();    
    filter.doFilter(req, res, filterChainMock);

    Mockito.verify(filterChainMock, Mockito.times(0)).doFilter(req, res);
    assertEquals(401, res.getStatus());
    assertEquals(res.getHeader("WWW-Authenticate"), "Basic");
  }
}
