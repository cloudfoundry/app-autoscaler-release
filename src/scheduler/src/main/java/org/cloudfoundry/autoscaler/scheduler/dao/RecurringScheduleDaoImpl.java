package org.cloudfoundry.autoscaler.scheduler.dao;

import java.util.List;
import org.cloudfoundry.autoscaler.scheduler.entity.RecurringScheduleEntity;
import org.cloudfoundry.autoscaler.scheduler.util.error.DatabaseValidationException;
import org.springframework.data.util.Pair;
import org.springframework.stereotype.Repository;
import org.springframework.transaction.annotation.Transactional;

@Repository("recurringScheduleDao")
public class RecurringScheduleDaoImpl extends GenericDaoImpl<RecurringScheduleEntity>
    implements RecurringScheduleDao {

  @Override
  public List<RecurringScheduleEntity> findAllRecurringSchedulesByAppId(String appId) {
    try {
      return entityManager
          .createNamedQuery(
              RecurringScheduleEntity.query_recurringSchedulesByAppId,
              RecurringScheduleEntity.class)
          .setParameter("appId", appId)
          .getResultList();

    } catch (Exception e) {
      throw new DatabaseValidationException("Find All recurring schedules by app id failed", e);
    }
  }

  @Override
  @Transactional(readOnly = true)
  public List<Pair<String, String>> getDistinctAppIdAndGuidList() {
    try {
      List<Object[]> res =
          entityManager
              .createNamedQuery(
                  RecurringScheduleEntity.query_findDistinctAppIdAndGuidFromRecurringSchedule,
                  Object[].class)
              .getResultList();
      return res.stream().map(r -> Pair.of((String) (r[0]), (String) (r[1]))).toList();

    } catch (Exception e) {
      throw new DatabaseValidationException("Find All recurring schedules failed", e);
    }
  }
}
