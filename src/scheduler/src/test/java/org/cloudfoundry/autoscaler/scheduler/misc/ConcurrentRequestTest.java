package org.cloudfoundry.autoscaler.scheduler.misc;

import static org.hamcrest.MatcherAssert.assertThat;
import static org.hamcrest.core.Is.is;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.concurrent.Callable;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;

import org.apache.hc.client5.http.utils.Base64;
import org.cloudfoundry.autoscaler.scheduler.util.EmbeddedTomcatUtil;
import org.cloudfoundry.autoscaler.scheduler.util.ScalingEngineUtil;
import org.junit.AfterClass;
import org.junit.BeforeClass;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.http.HttpEntity;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpMethod;
import org.springframework.test.context.junit4.SpringRunner;
import org.springframework.web.client.RestOperations;

@RunWith(SpringRunner.class)
@SpringBootTest
public class ConcurrentRequestTest {

  @Value("${autoscaler.scalingengine.url}")
  private String scalingEngineUrl;

  @Value("${autoscaler.scalingengine.basic_auth.username}")
  private String scalingEngineUsername;

  @Value("${autoscaler.scalingengine.basic_auth.password}")
  private String scalingEnginePassword;

  @Autowired
  private RestOperations restOperations;

  private static EmbeddedTomcatUtil embeddedTomcatUtil;

  @BeforeClass
  public static void beforeClass() throws IOException {
    embeddedTomcatUtil = new EmbeddedTomcatUtil();
    embeddedTomcatUtil.start();
  }

  @AfterClass
  public static void afterClass() throws IOException, InterruptedException {
    embeddedTomcatUtil.stop();
  }

  @Test
  public void testMultipleRequests() throws Exception {

    String appId = "appId";
    long scheduleId = 0L;

    embeddedTomcatUtil.addScalingEngineMockForAppAndScheduleId(appId, scheduleId, 200, null);

    String scalingEnginePathActiveSchedule =
        ScalingEngineUtil.getScalingEngineActiveSchedulePath(scalingEngineUrl, appId, scheduleId);

    concurrentRequests(10, scalingEnginePathActiveSchedule);
  }

  private void concurrentRequests(int threadCount, String scalingEnginePathActiveSchedule)
      throws Exception {


    HttpEntity<String> request = getRequest();

    Callable<Throwable> task = () -> {
      try {
        restOperations.exchange(scalingEnginePathActiveSchedule, HttpMethod.DELETE, request,
            Void.class);
        return null;

      } catch (Throwable th) {
        return th;
      }
    };

    List<Callable<Throwable>> tasks = Collections.nCopies(threadCount, task);
    ExecutorService executorService = Executors.newFixedThreadPool(threadCount);
    List<Future<Throwable>> futures = executorService.invokeAll(tasks);
    List<Throwable> resultList = new ArrayList<>(futures.size());

    for (Future<Throwable> future : futures) {
      resultList.add(future.get());
    }

    assertThat(futures.size(), is(threadCount));
    List<Throwable> expectedList = Collections.nCopies(threadCount, null);
    assertThat(resultList, is(expectedList));
  }

  private HttpEntity<String> getRequest() {
    String plainCreds = scalingEngineUsername + ":" + scalingEnginePassword;
    byte[] plainCredsBytes = plainCreds.getBytes();
    byte[] base64CredsBytes = Base64.encodeBase64(plainCredsBytes);
    String base64Creds = new String(base64CredsBytes);

    HttpHeaders headers = new HttpHeaders();
    headers.add("Authorization", "Basic " + base64Creds);
    HttpEntity<String> request = new HttpEntity<String>(headers);
    return request;
  }
}
