package org.cloudfoundry.autoscaler.scheduler.health;

import io.prometheus.client.Collector;
import io.prometheus.client.GaugeMetricFamily;
import java.util.ArrayList;
import java.util.List;
import javax.sql.DataSource;
import org.apache.commons.dbcp2.BasicDataSource;

public class DbStatusCollector extends Collector {

  private String namespace = "autoscaler";
  private String subSystem = "scheduler";

  private DataSource dataSource;

  public void setDataSource(DataSource dataSource) {
    this.dataSource = dataSource;
  }

  public void setPolicyDbDataSource(DataSource policyDbDataSource) {
    this.policyDbDataSource = policyDbDataSource;
  }

  private DataSource policyDbDataSource;

  private List<MetricFamilySamples> collectForDataSource(BasicDataSource dataSource, String name) {
    List<MetricFamilySamples> mfs = new ArrayList<MetricFamilySamples>();
    mfs.add(
        new GaugeMetricFamily(
            namespace + "_" + subSystem + name + "_initial_size",
            "The initial number of connections that are created when the pool is started",
            dataSource.getInitialSize()));
    mfs.add(
        new GaugeMetricFamily(
            namespace + "_" + subSystem + name + "_max_active",
            "The maximum number of active connections that can be allocated from this pool at the"
                + " same time, or negative for no limit",
            dataSource.getMaxTotal()));
    mfs.add(
        new GaugeMetricFamily(
            namespace + "_" + subSystem + name + "_max_idle",
            "The maximum number of connections that can remain idle in the pool, without extra ones"
                + " being released, or negative for no limit.",
            dataSource.getMaxIdle()));
    mfs.add(
        new GaugeMetricFamily(
            namespace + "_" + subSystem + name + "_min_idle",
            "The minimum number of active connections that can remain idle in the pool, without"
                + " extra ones being created, or 0 to create none.",
            dataSource.getMinIdle()));
    mfs.add(
        new GaugeMetricFamily(
            namespace + "_" + subSystem + name + "_active_connections_number",
            "The current number of active connections that have been allocated from this data"
                + " source",
            dataSource.getNumActive()));
    mfs.add(
        new GaugeMetricFamily(
            namespace + "_" + subSystem + name + "_idle_connections_number",
            "The current number of idle connections that are waiting to be allocated from this data"
                + " source",
            dataSource.getNumIdle()));
    return mfs;
  }

  @Override
  public List<MetricFamilySamples> collect() {
    List<MetricFamilySamples> mfs = new ArrayList<MetricFamilySamples>();
    BasicDataSource basicDataSource = (BasicDataSource) this.dataSource;
    mfs.addAll(collectForDataSource(basicDataSource, "_data_source"));

    BasicDataSource policyBasicDataSource = (BasicDataSource) this.policyDbDataSource;
    mfs.addAll(collectForDataSource(policyBasicDataSource, "_policy_db_data_source"));
    return mfs;
  }
}
