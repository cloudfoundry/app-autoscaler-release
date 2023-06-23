package org.cloudfoundry.autoscaler.scheduler.rest;

import static org.junit.jupiter.api.Assertions.assertEquals;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import java.io.IOException;
import org.cloudfoundry.autoscaler.scheduler.conf.HealthServerConfiguration;
import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.mock.web.MockHttpServletRequest;
import org.springframework.mock.web.MockHttpServletResponse;
import org.springframework.test.context.junit4.SpringRunner;

@RunWith(SpringRunner.class)
public class MultiPortFilterTest {

  private MockHttpServletRequest req;
  private MockHttpServletResponse res;

  @MockBean private FilterChain filterChainMock;

  @Before
  public void setUp() {
    req = new MockHttpServletRequest();
    res = new MockHttpServletResponse();
  }

  @Test
  public void shouldRespondTo8080IfURLContainsPort8080() throws IOException, ServletException {

    req.setLocalPort(8080);

    MultiPortFilter filter = new MultiPortFilter(new HealthServerConfiguration("", "", 8081, null));
    filter.doFilter(req, res, filterChainMock);
    assertEquals(8080, req.getLocalPort());
    assertEquals(200, res.getStatus());
  }

  @Test
  public void shouldRespond8080IfSchedulersURL() throws IOException, ServletException {

    req.setLocalPort(8080);

    req.setRequestURI("/v1/syncSchedules");

    req.setMethod("PUT");
    MultiPortFilter filter = new MultiPortFilter(new HealthServerConfiguration("", "", 8081, null));
    filter.doFilter(req, res, filterChainMock);
    assertEquals(8080, req.getLocalPort());
    assertEquals(200, res.getStatus());
  }

  @Test
  public void allowRequestIfPort8081WithHealthEndpoint() throws IOException, ServletException {

    req.setLocalPort(8081);

    req.setRequestURI("/health/");

    MultiPortFilter filter = new MultiPortFilter(new HealthServerConfiguration("", "", 8081, null));
    filter.doFilter(req, res, filterChainMock);
    assertEquals(8081, req.getLocalPort());
    assertEquals(200, res.getStatus());
  }

  @Test
  public void shouldRespond404IfPort8081WithNoHealthEndpoint()
      throws IOException, ServletException {
    req.setLocalPort(8081);

    MultiPortFilter filter = new MultiPortFilter(new HealthServerConfiguration("", "", 8081, null));
    filter.doFilter(req, res, null);
    assertEquals(8081, req.getLocalPort());
    assertEquals(404, res.getStatus());
  }

  @Test
  public void shouldRespond404IfPort8081WithNoHealthEndpointButWithSchedules()
      throws IOException, ServletException {
    req.setLocalPort(8081);
    req.setRequestURI("/v1/syncSchedules");
    req.setMethod("PUT");

    MultiPortFilter filter = new MultiPortFilter(new HealthServerConfiguration("", "", 8081, null));
    filter.doFilter(req, res, filterChainMock);
    assertEquals(8081, req.getLocalPort());
    assertEquals(404, res.getStatus());
  }
}
