package org.cloudfoundry.autoscaler.scheduler.util;

import jakarta.servlet.ServletConfig;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServlet;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import java.io.File;
import java.io.IOException;
import java.util.Hashtable;

import java.util.Base64;
import org.apache.catalina.Context;
import org.apache.catalina.LifecycleException;
import org.apache.catalina.Service;
import org.apache.catalina.connector.Connector;
import org.apache.catalina.startup.Tomcat;
import org.apache.tomcat.util.http.fileupload.FileUtils;

public class EmbeddedTomcatUtil {
  private File baseDir = null;
  private File applicationDir;
  private Context appContext;
  private Tomcat tomcat = new Tomcat();

  public EmbeddedTomcatUtil() {
    baseDir = new File("tomcat");
    tomcat.setBaseDir(baseDir.getAbsolutePath());

    Connector httpsConnector = new Connector();
    httpsConnector.setPort(8091);

    Service service = tomcat.getService();
    service.addConnector(httpsConnector);
    applicationDir = new File(baseDir + "/webapps", "/ROOT");

    if (!applicationDir.exists()) {
      applicationDir.mkdirs();
    }
    tomcat.setSilent(false);
  }

  public void start() {
    try {
      tomcat.start();
      appContext = tomcat.addWebapp("/", applicationDir.getAbsolutePath());
    } catch (LifecycleException e) {
      throw new RuntimeException(e);
    }
  }

  public void stop() {
    try {
      tomcat.stop();
      tomcat.destroy();
      // Tomcat creates a work folder where the temporary files are stored
      FileUtils.deleteDirectory(baseDir);
    } catch (LifecycleException | IOException e) {
      throw new RuntimeException(e);
    }
  }

  public void addScalingEngineMockForAppAndScheduleId(
      String appId, Long scheduleId, int statusCode, String message) throws ServletException {
    String url = "/v1/apps/" + appId + "/active_schedules/" + scheduleId;
    tomcat.addServlet(appContext.getPath(), appId, new ScalingEngineMock(statusCode, message));
    appContext.addServletMappingDecoded(url, appId);
  }

  public void addScalingEngineMockForAppId(String appId, int statusCode, String message) {
    String url = "/v1/apps/" + appId + "/active_schedules/*";
    tomcat.addServlet(appContext.getPath(), appId, new ScalingEngineMock(statusCode, message));
    appContext.addServletMappingDecoded(url, appId);
  }

  static class ScalingEngineMock extends HttpServlet {
    // Declare and initialize with a "scalingengine:scalingengine-password", "authorized"
    private int returnStatus;
    private String returnMessage;
    private Hashtable<String, String> validUsers = new Hashtable<>();

    public void init(ServletConfig config) throws ServletException {
        super.init(config);
        this.validUsers.put("scalingengine:scalingengine-password","authorized");
    }

    ScalingEngineMock(int status, String returnMessage) {
      this.returnStatus = status;
      this.returnMessage = returnMessage;
    }

    @Override
    protected void doPut(HttpServletRequest request, HttpServletResponse response)
        throws IOException {
        response.setContentType("application/json");

        if (allowRequest(request)) {
          response.setStatus(this.returnStatus);
          if (returnMessage != null && !returnMessage.isEmpty()) {
            response.getWriter().write(returnMessage);
          }
        } else {
            response.setHeader("WWW-Authenticate", "BASIC realm=\"jswan test\"");
            response.sendError(HttpServletResponse.SC_UNAUTHORIZED);
        }
    }

    @Override
    protected void doDelete(HttpServletRequest request, HttpServletResponse response)
        throws IOException {
        response.setContentType("application/json");

        if (allowRequest(request)) {
          response.setStatus(this.returnStatus);
          if (returnMessage != null && !returnMessage.isEmpty()) {
            response.getWriter().write(returnMessage);
          }
        } else {
            response.setHeader("WWW-Authenticate", "BASIC realm=\"jswan test\"");
            response.sendError(HttpServletResponse.SC_UNAUTHORIZED);
        }
    }

    protected boolean allowRequest(HttpServletRequest request) throws IOException {
      String auth = request.getHeader("Authorization");
      if (auth == null || !auth.toUpperCase().startsWith("BASIC ")) { 
          return false;  
      }
      String userpassEncoded = auth.substring(6);
      byte[] decodedBytes = Base64.getDecoder().decode(userpassEncoded);
      String userpassDecoded = new String(decodedBytes);
  
      if ("authorized".equals(this.validUsers.get(userpassDecoded))) {
          return true;
      } else {
          return false;
      }
    }
  }
}
