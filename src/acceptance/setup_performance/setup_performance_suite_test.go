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

	"github.com/cloudfoundry/cf-test-helpers/v2/cf"
	"github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry/cf-test-helpers/v2/workflowhelpers"

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

	if os.Getenv("SKIP_TEARDOWN") == "true" {
		fmt.Println("Skipping Teardown...")
	} else {
		cleanup()
	}

	setup = workflowhelpers.NewRunawayAppTestSuiteSetup(cfg)
	setup.Setup()

	EnableServiceAccess(setup, cfg, setup.GetOrganizationName())
	workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
		_, orgGuid, _, spaceGuid := GetOrgSpaceNamesAndGuids(cfg, setup.GetOrganizationName())
		Expect(spaceGuid).NotTo(BeEmpty())
		updateOrgQuotaForPerformanceTest(orgGuid)
	})

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

func cleanup() {
	setup = workflowhelpers.NewTestSuiteSetup(cfg)
	workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
		fmt.Println("\nCleaning up test leftovers...")
		if cfg.UseExistingOrganization {
			targetOrg(setup)
			orgGuid := GetOrgGuid(cfg, cfg.ExistingOrganization)
			spaceNames := GetTestSpaces(orgGuid, cfg)
			if len(spaceNames) == 0 {
				return
			}
			waitGroup := sync.WaitGroup{}
			waitGroup.Add(2)

			deleteAllServices(cfg.Performance.SetupWorkers, orgGuid, GetSpaceGuid(cfg, orgGuid), &waitGroup)
			// delete all apps in a test space - only from first space - what if two spaces are present
			deleteAllApps(cfg.Performance.SetupWorkers, orgGuid, GetSpaceGuid(cfg, orgGuid), &waitGroup)
			fmt.Println("\nWaiting for services and apps to be deleted...")
			waitGroup.Wait()
			DeleteSpaces(cfg.ExistingOrganization, GetTestSpaces(orgGuid, cfg), cfg.DefaultTimeoutDuration())
		} else {
			DeleteOrgs(GetTestOrgs(cfg), time.Duration(120)*time.Second)
		}
	})
}

func deleteAllServices(workersCount int, orgGuid string, spaceGuid string, parentWaitGroup *sync.WaitGroup) {
	defer parentWaitGroup.Done()
	defer GinkgoRecover()
	waitGroup := sync.WaitGroup{}
	servicesChan := make(chan string)

	services := GetServices(cfg, orgGuid, spaceGuid)
	if len(services) == 0 {
		fmt.Printf("- deleting existing service instances: %d\n", len(services))
		return
	}
	fmt.Printf("- deleting existing service instances: %d\n", len(services))
	for i := 1; i <= workersCount; i++ {
		waitGroup.Add(1)
		go deleteExistingServiceInstances(i, servicesChan, setup, &waitGroup)
	}
	for _, serviceInstanceName := range services {
		servicesChan <- serviceInstanceName
	}
	close(servicesChan)
	waitGroup.Wait()
}

func deleteAllApps(workersCount int, orgGuid string, spaceName string, parentWaitGroup *sync.WaitGroup) {
	defer parentWaitGroup.Done()
	defer GinkgoRecover()
	waitGroup := sync.WaitGroup{}
	appsChan := make(chan string)

	apps := GetApps(cfg, orgGuid, spaceName, "node-custom-metric-benchmark-")
	fmt.Printf("\n- deleting existing app instances: %d\n", len(apps))
	if len(apps) == 0 {
		return
	}
	for i := 1; i <= workersCount; i++ {
		waitGroup.Add(1)
		go deleteExistingApps(i, appsChan, &waitGroup)
	}
	for _, app := range apps {
		appsChan <- app
	}
	close(appsChan)
	waitGroup.Wait()
}

func deleteExistingServiceInstances(workerId int, servicesChan chan string, setup *workflowhelpers.ReproducibleTestSuiteSetup, wg *sync.WaitGroup) {
	defer wg.Done()
	defer GinkgoRecover()

	for instanceName := range servicesChan {
		fmt.Printf("worker %d  - deleting service instance - %s\n", workerId, instanceName)
		DeleteServiceInstance(cfg, setup, instanceName)
	}
}

func deleteExistingApps(workerId int, appsChan chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	defer GinkgoRecover()

	for appName := range appsChan {
		fmt.Printf("worker %d  - deleting app instance - %s\n", workerId, appName)
		DeleteTestApp(appName, cfg.DefaultTimeoutDuration())
	}
}

func targetOrg(setup *workflowhelpers.ReproducibleTestSuiteSetup) {
	cmd := cf.Cf("target", "-o", setup.GetOrganizationName()).Wait(cfg.DefaultTimeoutDuration())
	Expect(cmd).To(gexec.Exit(0), fmt.Sprintf("failed cf target org  %s : %s", setup.GetOrganizationName(), string(cmd.Err.Contents())))
}
