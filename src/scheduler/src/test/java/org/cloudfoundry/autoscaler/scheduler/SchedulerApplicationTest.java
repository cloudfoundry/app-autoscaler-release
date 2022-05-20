package org.cloudfoundry.autoscaler.scheduler;

import static org.hamcrest.Matchers.equalToIgnoringCase;
import static org.junit.Assert.assertThat;

import org.apache.commons.dbcp2.BasicDataSource;
import org.junit.Rule;
import org.junit.Test;
import org.junit.rules.ExpectedException;
import org.junit.runner.RunWith;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.context.ApplicationContextException;
import org.springframework.test.context.junit4.SpringRunner;

@RunWith(SpringRunner.class)
@SpringBootTest
public class SchedulerApplicationTest {
  @Rule public ExpectedException expectedEx = ExpectedException.none();
  @Autowired private BasicDataSource dataSource;

  @Test
  public void testTomcatConnectionPoolNameCorrect() {
    assertThat(
        dataSource.getClass().getName(),
        equalToIgnoringCase("org.apache.commons.dbcp2.BasicDataSource"));
  }

  @Test
  public void testApplicationExitsWhenSchedulerDbUnreachable() {
    expectedEx.expect(ApplicationContextException.class);
    SchedulerApplication.main(
        new String[] {
          "--spring.autoconfigure.exclude="
              + "org.springframework.boot.actuate.autoconfigure.jdbc."
              + "DataSourceHealthIndicatorAutoConfiguration",
          "--spring.datasource.url=jdbc:postgresql://127.0.0.1/wrong-scheduler-db"
        });
  }

  @Test
  public void testApplicationExitsWhenPolicyDbUnreachable() {
    expectedEx.expect(ApplicationContextException.class);
    SchedulerApplication.main(
        new String[] {
          "--spring.autoconfigure.exclude="
              + "org.springframework.boot.actuate.autoconfigure.jdbc."
              + "DataSourceHealthIndicatorAutoConfiguration",
          "--spring.policy-db-datasource.url=jdbc:postgresql://127.0.0.1/wrong-policy-db"
        });
  }
}
