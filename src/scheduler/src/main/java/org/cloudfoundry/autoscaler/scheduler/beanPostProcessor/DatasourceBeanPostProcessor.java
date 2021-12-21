package org.cloudfoundry.autoscaler.scheduler.beanPostProcessor;

import java.sql.Connection;
import java.sql.SQLException;
import javax.sql.DataSource;
import org.cloudfoundry.autoscaler.scheduler.util.error.DataSourceConfigurationException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.BeansException;
import org.springframework.beans.factory.config.BeanPostProcessor;

public class DatasourceBeanPostProcessor implements BeanPostProcessor {
  private Logger logger = LoggerFactory.getLogger(this.getClass());

  @Override
  public Object postProcessBeforeInitialization(Object bean, String beanName)
      throws BeansException {
    // TODO Auto-generated method stub
    return bean;
  }

  @Override
  public Object postProcessAfterInitialization(Object bean, String beanName) throws BeansException {
    // TODO Auto-generated method stub
    if (bean instanceof DataSource) {
      DataSource ds = (DataSource) bean;
      Connection con = null;
      try {
        con = ds.getConnection();
      } catch (SQLException e) {
        // TODO Auto-generated catch block
        logger.error("Failed to connect to datasource: " + beanName);
        throw new DataSourceConfigurationException(
            beanName, "Failed to connect to datasource:" + beanName, e);
      } finally {
        try {
          if (con != null && !con.isClosed()) {
            con.close();
          }
        } catch (SQLException e) {
          // TODO Auto-generated catch block
          logger.error("Failed to close connection from " + beanName, e);
        }
      }
    }
    return bean;
  }
}
