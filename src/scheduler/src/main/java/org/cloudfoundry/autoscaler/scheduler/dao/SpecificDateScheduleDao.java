package org.cloudfoundry.autoscaler.scheduler.dao;

import java.util.List;
import org.cloudfoundry.autoscaler.scheduler.entity.SpecificDateScheduleEntity;
import org.springframework.data.util.Pair;

/** */
public interface SpecificDateScheduleDao extends GenericDao<SpecificDateScheduleEntity> {

  public List<SpecificDateScheduleEntity> findAllSpecificDateSchedulesByAppId(String appId);

  public List<Pair<String, String>> getDistinctAppIdAndGuidList();
}
