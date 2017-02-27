package app

import (
	"fmt"
	"testing"

	"acceptance/config"

	"encoding/json"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"math"
	"strconv"
	"time"
)

type appSummary struct {
	RunningInstances int `json:"running_instances"`
}

const MB = 1024 * 1024

var (
	cfg   *config.Config
	setup *workflowhelpers.ReproducibleTestSuiteSetup
)

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)

	cfg = config.LoadConfig(t)
	componentName := "Application Scale Suite"
	rs := []Reporter{}

	if cfg.GetArtifactsDirectory() != "" {
		helpers.EnableCFTrace(cfg, componentName)
		rs = append(rs, helpers.NewJUnitReporter(cfg, componentName))
	}

	RunSpecsWithDefaultAndCustomReporters(t, componentName, rs)
}

var _ = BeforeSuite(func() {

	setup = workflowhelpers.NewTestSuiteSetup(cfg)
	setup.Setup()

	serviceExists := cf.Cf("marketplace", "-s", cfg.ServiceName).Wait(cfg.DefaultTimeoutDuration())
	Expect(serviceExists).To(Exit(0), fmt.Sprintf("Service offering, %s, does not exist", cfg.ServiceName))
})

var _ = AfterSuite(func() {
	setup.Teardown()
})

func runningInstances(appGUID string, timeout time.Duration) int {
	cmd := cf.Cf("curl", "/v2/apps/"+appGUID+"/summary")
	Expect(cmd.Wait(timeout)).To(Exit(0))

	var summary appSummary
	err := json.Unmarshal(cmd.Out.Contents(), &summary)
	Expect(err).ToNot(HaveOccurred())
	return summary.RunningInstances
}

func waitForNInstancesRunning(appGUID string, instances int, timeout time.Duration) {
	Eventually(func() int {
		return runningInstances(appGUID, timeout)
	}, timeout, 10*time.Second).Should(Equal(instances))
}

type instanceStats struct {
	MemQuota uint64 `json:"mem_quota"`
	Usage    instanceUsage
}

type instanceUsage struct {
	Mem uint64
}

type instanceInfo struct {
	State string
	Stats instanceStats
}

type appStats map[string]*instanceInfo

func memoryUsed(appGUID string, index int, timeout time.Duration) (uint64, uint64) {
	cmd := cf.Cf("curl", "/v2/apps/"+appGUID+"/stats")
	Expect(cmd.Wait(timeout)).To(Exit(0))

	var stats appStats
	err := json.Unmarshal(cmd.Out.Contents(), &stats)
	Expect(err).ToNot(HaveOccurred())

	instance := stats[strconv.Itoa(index)]
	if instance == nil {
		return 0, 0
	}
	return instance.Stats.Usage.Mem, instance.Stats.MemQuota
}

func allInstancesMemoryUsed(appGUID string, timeout time.Duration) []uint64 {
	cmd := cf.Cf("curl", "/v2/apps/"+appGUID+"/stats")
	Expect(cmd.Wait(timeout)).To(Exit(0))

	var stats appStats
	err := json.Unmarshal(cmd.Out.Contents(), &stats)
	Expect(err).ToNot(HaveOccurred())

	if len(stats) == 0 {
		return []uint64{}
	}

	mem := make([]uint64, len(stats))

	for k, instance := range stats {
		i, err := strconv.Atoi(k)
		Expect(err).NotTo(HaveOccurred())
		mem[i] = instance.Stats.Usage.Mem
	}

	return mem
}

func averageMemoryUsedByInstance(appGUID string, timeout time.Duration) uint64 {
	memoryUsedArray := allInstancesMemoryUsed(appGUID, timeout)
	instanceCount := len(memoryUsedArray)
	if instanceCount == 0 {
		return math.MaxInt64
	}

	var memSum uint64
	for _, m := range memoryUsedArray {
		memSum += m
	}

	return memSum / uint64(len(memoryUsedArray))
}
