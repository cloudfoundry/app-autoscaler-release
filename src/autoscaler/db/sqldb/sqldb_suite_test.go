package sqldb_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/db"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/models"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var dbHelper *sqlx.DB

func TestSqldb(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sqldb Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	var e error

	dbUrl := os.Getenv("DBURL")
	if dbUrl == "" {
		Fail("environment variable $DBURL is not set")
	}
	database, err := db.GetConnection(dbUrl)
	if err != nil {
		Fail("failed to parse database connection: " + err.Error())
	}

	dbHelper, e = sqlx.Open(database.DriverName, database.DSN)
	if e != nil {
		Fail("can not connect database: " + e.Error())
	}

	e = createLockTable()
	if e != nil {
		Fail("can not create test lock table: " + e.Error())
	}

	_, e = dbHelper.Exec("DELETE from binding")
	if e != nil {
		Fail("can not clean table binding: " + e.Error())
	}

	_, e = dbHelper.Exec("DELETE from service_instance")
	if e != nil {
		Fail("can not clean table service_instance: " + e.Error())
	}

	if strings.Contains(os.Getenv("DBURL"), "postgres") && getPostgresMajorVersion() >= 12 {
		deleteAllFunctions()
		addPSQLFunctions()
	}

	_ = dbHelper.Close()
	dbHelper = nil

	return []byte{}
}, func([]byte) {
	var e error

	dbUrl := os.Getenv("DBURL")
	if dbUrl == "" {
		Fail("environment variable $DBURL is not set")
	}
	database, err := db.GetConnection(dbUrl)
	if err != nil {
		Fail("failed to parse database connection: " + err.Error())
	}

	dbHelper, e = sqlx.Open(database.DriverName, database.DSN)
	if e != nil {
		Fail("can not connect database: " + e.Error())
	}
})

var _ = SynchronizedAfterSuite(func() {
	if dbHelper != nil && GinkgoParallelProcess() != 1 {
		_ = dbHelper.Close()
	}
}, func() {
	e := dropLockTable()
	if e != nil {
		Fail("can not drop test lock table: " + e.Error())
	}
	if dbHelper != nil && GinkgoParallelProcess() == 1 {
		_ = dbHelper.Close()
	}
})

func cleanInstanceMetricsTable() {
	_, e := dbHelper.Exec("DELETE FROM appinstancemetrics")
	if e != nil {
		Fail("can not clean table appinstancemetrics:" + e.Error())
	}
}

func hasInstanceMetric(appId string, index int, name string, timestamp int64) bool {
	query := dbHelper.Rebind("SELECT * FROM appinstancemetrics WHERE appid = ? AND instanceindex = ? AND name = ? AND timestamp = ?")
	rows, e := dbHelper.Query(query, appId, index, name, timestamp)
	if e != nil {
		Fail("can not query table appinstancemetrics: " + e.Error())
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()
	return rows.Next()
}

func getNumberOfInstanceMetrics() int {
	var num int
	e := dbHelper.QueryRow("SELECT COUNT(*) FROM appinstancemetrics").Scan(&num)
	if e != nil {
		Fail("can not count the number of records in table appinstancemetrics: " + e.Error())
	}
	return num
}

func hasServiceInstance(serviceInstanceId string) bool {
	query := dbHelper.Rebind("SELECT * FROM service_instance WHERE service_instance_id = ?")
	rows, e := dbHelper.Query(query, serviceInstanceId)
	if e != nil {
		Fail("can not query table service_instance: " + e.Error())
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()
	return rows.Next()
}

func hasServiceInstanceWithNullDefaultPolicy(serviceInstanceId string) bool {
	query := dbHelper.Rebind("SELECT * FROM service_instance WHERE service_instance_id = ? AND default_policy IS NULL AND default_policy_guid IS NULL")
	rows, e := dbHelper.Query(query, serviceInstanceId)
	if e != nil {
		Fail("can not query table service_instance: " + e.Error())
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()

	return rows.Next()
}

func hasServiceBinding(bindingId string, serviceInstanceId string) bool {
	query := dbHelper.Rebind("SELECT * FROM binding WHERE binding_id = ? AND service_instance_id = ? ")
	rows, e := dbHelper.Query(query, bindingId, serviceInstanceId)
	if e != nil {
		Fail("can not query table binding: " + e.Error())
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()
	return rows.Next()
}

func cleanPolicyTable() {
	_, e := dbHelper.Exec("DELETE from policy_json")
	if e != nil {
		Fail("can not clean table policy_json: " + e.Error())
	}
}

func insertPolicy(appId string, scalingPolicy *models.ScalingPolicy, policyGuid string) {
	policyJson, e := json.Marshal(scalingPolicy)
	if e != nil {
		Fail("failed to marshall scaling policy" + e.Error())
	}

	query := dbHelper.Rebind("INSERT INTO policy_json(app_id, policy_json, guid) VALUES(?, ?, ?)")
	_, e = dbHelper.Exec(query, appId, string(policyJson), policyGuid)

	if e != nil {
		Fail(fmt.Sprintf("can not insert app:%s data to table policy_json: %s", appId, e.Error()))
	}
}

func insertPolicyWithGuid(appId string, scalingPolicy *models.ScalingPolicy, guid string) {
	By("Insert policy:" + guid)
	policyJson, e := json.Marshal(scalingPolicy)
	if e != nil {
		Fail("failed to marshall scaling policy" + e.Error())
	}

	query := dbHelper.Rebind("INSERT INTO policy_json(app_id, policy_json, guid) VALUES(?, ?, ?)")
	_, e = dbHelper.Exec(query, appId, string(policyJson), guid)

	if e != nil {
		Fail("can not insert data to table policy_json: " + e.Error())
	}
}

func getAppPolicy(appId string) string {
	query := dbHelper.Rebind("SELECT policy_json FROM policy_json WHERE app_id=? ")
	rows, err := dbHelper.Query(query, appId)
	if err != nil {
		Fail("failed to get policy" + err.Error())
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()
	var policyJsonStr string
	if rows.Next() {
		err = rows.Scan(&policyJsonStr)
		if err != nil {
			Fail("failed to scan policy" + err.Error())
		}
	}
	return policyJsonStr
}

func cleanAppMetricTable() {
	_, e := dbHelper.Exec("DELETE from app_metric")
	if e != nil {
		Fail("can not clean table app_metric : " + e.Error())
	}
}

func hasAppMetric(appId, metricType string, timestamp int64, value string) bool {
	query := dbHelper.Rebind("SELECT * FROM app_metric WHERE app_id = ? AND metric_type = ? AND timestamp = ? AND value = ?")
	rows, e := dbHelper.Query(query, appId, metricType, timestamp, value)
	if e != nil {
		Fail("can not query table app_metric: " + e.Error())
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()
	return rows.Next()
}

func getNumberOfAppMetrics() int {
	var num int
	e := dbHelper.QueryRow("SELECT COUNT(*) FROM app_metric").Scan(&num)
	if e != nil {
		Fail("can not count the number of records in table app_metric: " + e.Error())
	}
	return num
}

func removeScalingHistoryForApp(appId string) {
	query := dbHelper.Rebind("DELETE from scalinghistory where appId = ?")
	_, err := dbHelper.Exec(query, appId)
	if err != nil {
		Fail("can not clean table scalinghistory: " + err.Error())
	}
}

func removeCooldownForApp(appId string) {
	query := dbHelper.Rebind("DELETE from scalingcooldown where appId = ?")
	_, err := dbHelper.Exec(query, appId)
	if err != nil {
		Fail("can not clean table scalingcooldown: " + err.Error())
	}
}
func removeActiveScheduleForApp(appId string) {
	query := dbHelper.Rebind("DELETE from activeschedule where appId = ?")
	_, err := dbHelper.Exec(query, appId)
	if err != nil {
		Fail("can not clean table scalingcooldown: " + err.Error())
	}
}

func hasScalingHistory(appId string, timestamp int64) bool {
	query := dbHelper.Rebind("SELECT * FROM scalinghistory WHERE appid = ? AND timestamp = ?")
	rows, e := dbHelper.Query(query, appId, timestamp)
	if e != nil {
		Fail("can not query table scalinghistory: " + e.Error())
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()
	return rows.Next()
}

func getScalingHistoryForApp(appId string) int {
	var num int
	query := dbHelper.Rebind("SELECT COUNT(*) FROM scalinghistory WHERE appid = ?")
	row := dbHelper.QueryRow(query, appId)
	err := row.Scan(&num)
	if err != nil {
		Fail("can not count the number of records in table scalinghistory: " + err.Error())
	}
	return num
}

func hasScalingCooldownRecord(appId string, expireAt int64) bool {
	query := dbHelper.Rebind("SELECT * FROM scalingcooldown WHERE appid = ? AND expireat = ?")
	rows, e := dbHelper.Query(query, appId, expireAt)
	if e != nil {
		Fail("can not query table scalingcooldown: " + e.Error())
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()
	return rows.Next()
}

func insertActiveSchedule(appId, scheduleId string, instanceMin, instanceMax, instanceMinInitial int) error {
	query := dbHelper.Rebind("INSERT INTO activeschedule(appid, scheduleid, instancemincount, instancemaxcount, initialmininstancecount) " +
		" VALUES (?, ?, ?, ?, ?)")
	_, e := dbHelper.Exec(query, appId, scheduleId, instanceMin, instanceMax, instanceMinInitial)
	return e
}

func cleanSchedulerActiveScheduleTable() error {
	_, e := dbHelper.Exec("DELETE from app_scaling_active_schedule")
	return e
}

func insertSchedulerActiveSchedule(id int, appId string, startJobIdentifier int, instanceMin, instanceMax, instanceMinInitial int) error {
	var e error
	var query string
	if instanceMinInitial <= 0 {
		query = dbHelper.Rebind("INSERT INTO app_scaling_active_schedule(id, app_id, start_job_identifier, instance_min_count, instance_max_count) " +
			" VALUES (?, ?, ?, ?, ?)")
		_, e = dbHelper.Exec(query, id, appId, startJobIdentifier, instanceMin, instanceMax)
	} else {
		query = dbHelper.Rebind("INSERT INTO app_scaling_active_schedule(id, app_id, start_job_identifier, instance_min_count, instance_max_count, initial_min_instance_count) " +
			" VALUES (?, ?, ?, ?, ?, ?)")
		_, e = dbHelper.Exec(query, id, appId, startJobIdentifier, instanceMin, instanceMax, instanceMinInitial)
	}
	return e
}

func insertCredential(appid string, username string, password string) error {
	var err error
	query := dbHelper.Rebind("INSERT INTO credentials(id, username, password, updated_at) values(?, ?, ?, ?)")
	_, err = dbHelper.Exec(query, appid, username, password, "2011-05-18 15:36:38")
	return err
}

func getCredential(appId string) (string, string, error) {
	query := dbHelper.Rebind("SELECT username,password FROM credentials WHERE id=? ")
	rows, err := dbHelper.Query(query, appId)
	if err != nil {
		Fail("failed to get credential" + err.Error())
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()
	var username, password string
	if rows.Next() {
		err = rows.Scan(&username, &password)
		if err != nil {
			Fail("failed to scan credential" + err.Error())
		}
	}
	return username, password, nil
}
func hasCredential(appId string) bool {
	query := dbHelper.Rebind("SELECT * FROM credentials WHERE id=?")
	rows, e := dbHelper.Query(query, appId)
	if e != nil {
		Fail("can not query table credentials: " + e.Error())
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()
	return rows.Next()
}

func insertLockDetails(lock *models.Lock) (sql.Result, error) {
	query := dbHelper.Rebind("INSERT INTO test_lock (owner,lock_timestamp,ttl) VALUES (?,?,?)")
	result, err := dbHelper.Exec(query, lock.Owner, lock.LastModifiedTimestamp, int64(lock.Ttl/time.Second))
	return result, err
}

func cleanLockTable() error {
	_, err := dbHelper.Exec("DELETE FROM test_lock")
	if err != nil {
		return err
	}
	return nil
}

func dropLockTable() error {
	_, err := dbHelper.Exec("DROP TABLE test_lock")
	if err != nil {
		return err
	}
	return nil
}

func createLockTable() error {
	_, err := dbHelper.Exec(`
		CREATE TABLE IF NOT EXISTS test_lock (
			owner VARCHAR(255) PRIMARY KEY,
			lock_timestamp TIMESTAMP  NOT NULL,
			ttl BIGINT DEFAULT 0
		);
	`)
	if err != nil {
		return err
	}
	return nil
}

func validateLockInDB(ownerid string, expectedLock *models.Lock) error {
	var (
		timestamp time.Time
		ttl       time.Duration
		owner     string
	)
	query := dbHelper.Rebind("SELECT owner,lock_timestamp,ttl FROM test_lock WHERE owner=?")
	row := dbHelper.QueryRow(query, ownerid)
	err := row.Scan(&owner, &timestamp, &ttl)
	if err != nil {
		return err
	}
	errMsg := ""
	if expectedLock.Owner != owner {
		errMsg += fmt.Sprintf("mismatch owner (%s, %s),", expectedLock.Owner, owner)
	}
	if expectedLock.Ttl != time.Second*ttl {
		errMsg += fmt.Sprintf("mismatch ttl (%d, %d),", expectedLock.Ttl, time.Second*ttl)
	}
	if errMsg != "" {
		return errors.New(errMsg)
	}
	return nil
}

func validateLockNotInDB(owner string) error {
	var (
		timestamp time.Time
		ttl       time.Duration
	)
	query := dbHelper.Rebind("SELECT owner,lock_timestamp,ttl FROM test_lock WHERE owner=?")
	row := dbHelper.QueryRow(query, owner)
	err := row.Scan(&owner, &timestamp, &ttl)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	return fmt.Errorf("lock exists with owner (%s)", owner)
}

func formatPolicyString(policyStr string) (string, error) {
	scalingPolicy := &models.ScalingPolicy{}
	err := json.Unmarshal([]byte(policyStr), &scalingPolicy)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal policyJson string %s", policyStr)
	}
	policyJsonStr, err := json.Marshal(scalingPolicy)
	if err != nil {
		return "", fmt.Errorf("failed to marshal ScalingPolicy %v", scalingPolicy)
	}
	return string(policyJsonStr), nil
}

func expectServiceInstancesToEqual(actual *models.ServiceInstance, expected *models.ServiceInstance) {
	ExpectWithOffset(1, actual.ServiceInstanceId).To(Equal(expected.ServiceInstanceId))
	ExpectWithOffset(1, actual.OrgId).To(Equal(expected.OrgId))
	ExpectWithOffset(1, actual.SpaceId).To(Equal(expected.SpaceId))
	ExpectWithOffset(1, actual.DefaultPolicy).To(MatchJSON(expected.DefaultPolicy))
	ExpectWithOffset(1, actual.DefaultPolicyGuid).To(Equal(expected.DefaultPolicyGuid))
}
