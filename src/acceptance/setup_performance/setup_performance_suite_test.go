package peformance_setup_test

import (
	"acceptance/config"
	. "acceptance/helpers"
	"fmt"
	"github.com/KevinJCross/cf-test-helpers/v2/cf"
	"github.com/onsi/gomega/gexec"
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

		if cfg.UseExistingOrganization {
			// cf target to org
			targetOrg(setup)

			orgGuid := GetOrgGuid(cfg, cfg.ExistingOrganization)
			spaceNames := GetTestSpaces(orgGuid, cfg)
			if len(spaceNames) == 0 {
				return
			}
			//TODO - Do it with multiple go processes1
			deleteAllServices(orgGuid)
			// delete all apps in a test space - only from first space - what if two spaces are present
			deleteAllApps(spaceNames[0])
			DeleteSpaces(cfg.ExistingOrganization, GetTestSpaces(orgGuid, cfg), cfg.DefaultTimeoutDuration())
		} else {
			DeleteOrgs(GetTestOrgs(cfg), time.Duration(120)*time.Second)
		}

	})
}

func deleteAllApps(spaceName string) {
	apps := GetApps(cfg, setup.GetOrganizationName(), spaceName, "node-custom-metric-benchmark-")
	fmt.Printf("\nExisting apps found %d", len(apps))
	for _, appName := range apps {
		fmt.Printf(" - deleting app %s\n", appName)
		DeleteTestApp(appName, cfg.DefaultTimeoutDuration())
	}
}

func deleteAllServices(orgGuid string) {
	services := GetServices(cfg, orgGuid, GetSpaceGuid(cfg, orgGuid))
	fmt.Printf("\nExisting services found %d", len(services))
	for _, service := range services {
		fmt.Printf(" - deleting service instance %s\n", service)
		DeleteServiceInstance(cfg, setup, service)
	}
}

func targetOrg(setup *workflowhelpers.ReproducibleTestSuiteSetup) {
	cmd := cf.Cf("target", "-o", setup.GetOrganizationName()).Wait(cfg.DefaultTimeoutDuration())
	Expect(cmd).To(gexec.Exit(0), fmt.Sprintf("failed cf target org  %s : %s", setup.GetOrganizationName(), string(cmd.Err.Contents())))
}
