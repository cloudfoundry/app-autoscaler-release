package peformance_setup_test

import (
	"acceptance/config"
	. "acceptance/helpers"
	"fmt"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/KevinJCross/cf-test-helpers/v2/workflowhelpers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	cfg                *config.Config
	setup              *workflowhelpers.ReproducibleTestSuiteSetup
	originalOrgQuota   OrgQuota
	nodeAppDropletPath string
)

func TestSetup(t *testing.T) {
	RegisterFailHandler(Fail)
	cfg = config.LoadConfig()
	cfg.Prefix = "autoscaler-performance-TESTS"
	setup = workflowhelpers.NewTestSuiteSetup(cfg)
	RunSpecs(t, "Setup Performance Test Suite")
}

var _ = BeforeSuite(func() {
	var spaceGuid, orgGuid string

	if os.Getenv("SKIP_TEARDOWN") == "true" {
		fmt.Println("Skipping Teardown...")
	} else {
		cleanup()
	}

	setup = workflowhelpers.NewRunawayAppTestSuiteSetup(cfg)
	setup.Setup()

	EnableServiceAccess(setup, cfg, setup.GetOrganizationName())
	workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
		_, orgGuid, _, spaceGuid = GetOrgSpaceNamesAndGuids(cfg, setup.GetOrganizationName())
		updateOrgQuotaForPerformanceTest(orgGuid)
	})

	cleanUpServiceInstanceInParallel(setup, orgGuid, spaceGuid)

	if cfg.IsServiceOfferingEnabled() {
		CheckServiceExists(cfg, setup.TestSpace.SpaceName(), cfg.ServiceName)
	}

	fmt.Print("\ncreating droplet...")
	nodeAppDropletPath = CreateDroplet(*cfg)
	fmt.Println("done")
})

func updateOrgQuotaForPerformanceTest(orgGuid string) {
	if cfg.Performance.UpdateExistingOrgQuota {
		originalOrgQuota = GetOrgQuota(orgGuid, cfg.DefaultTimeoutDuration())
		fmt.Printf("\n=> originalOrgQuota %+v\n", originalOrgQuota)
		performanceOrgQuota := OrgQuota{
			Name:             originalOrgQuota.Name,
			AppInstances:     strconv.Itoa(cfg.Performance.AppCount * 2),
			TotalMemory:      strconv.Itoa(cfg.Performance.AppCount*256) + "MB",
			Routes:           strconv.Itoa(cfg.Performance.AppCount * 2),
			ServiceInstances: strconv.Itoa(cfg.Performance.AppCount * 2),
			RoutePorts:       "-1",
		}
		fmt.Printf("=> setting new org quota %s\n", originalOrgQuota.Name)
		UpdateOrgQuota(performanceOrgQuota, cfg.DefaultTimeoutDuration())
	}
}

func cleanUpServiceInstanceInParallel(setup *workflowhelpers.ReproducibleTestSuiteSetup, orgGuid string, spaceGuid string) {
	waitGroup := sync.WaitGroup{}
	servicesChan := make(chan string)

	serviceInstances := GetServices(cfg, orgGuid, spaceGuid)
	if len(serviceInstances) != 0 {
		fmt.Printf("\ndeleting existing service instances: %d\n", len(serviceInstances))
		for i := 0; i < len(serviceInstances); i++ {
			waitGroup.Add(1)
			i := i
			go deleteExistingServiceInstances(i, servicesChan, setup, orgGuid, spaceGuid, &waitGroup)
		}
		for _, serviceInstanceName := range serviceInstances {
			servicesChan <- serviceInstanceName
		}
		close(servicesChan)
		waitGroup.Wait()
	}
}

func deleteExistingServiceInstances(workerId int, servicesChan chan string, setup *workflowhelpers.ReproducibleTestSuiteSetup, orgGuid string, spaceGuid string, wg *sync.WaitGroup) {
	fmt.Printf("Worker %d  - Delete Service Instance starting...\n", workerId)
	defer wg.Done()
	defer GinkgoRecover()
	for instanceName := range servicesChan {
		fmt.Printf("worker %d  - deleting service instance - %s\n", workerId, instanceName)
		DeleteServiceInstance(cfg, setup, instanceName)
	}
	fmt.Printf("worker %d  - Delete Service Instance finished...\n", workerId)
}

func cleanup() {
	setup = workflowhelpers.NewTestSuiteSetup(cfg)
	workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
		DeleteOrgs(GetTestOrgs(cfg), time.Duration(120)*time.Second)

		if cfg.UseExistingOrganization {
			orgGuid := GetOrgGuid(cfg, cfg.ExistingOrganization)
			DeleteSpaces(cfg.ExistingOrganization, GetTestSpaces(orgGuid, cfg), 0*time.Second)
		}
	})
}
