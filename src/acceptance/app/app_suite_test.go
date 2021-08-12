package app_test

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sort"
	"strings"
	"testing"
	"time"

	"acceptance/config"
	. "acceptance/helpers"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

const (
	PolicyPath          = "/v1/apps/{appId}/policy"
	CustomMetricPath    = "/v1/apps/{appId}/credential"
	CustomMetricCredEnv = "AUTO_SCALER_CUSTOM_METRIC_ENV"
)

var (
	cfg      *config.Config
	setup    *workflowhelpers.ReproducibleTestSuiteSetup
	interval int
	client   *http.Client

	instanceName         string
	initialInstanceCount int
)

type CFResourceObject struct {
	Resources []struct {
		GUID      string `json:"guid"`
		CreatedAt string `json:"created_at"`
		Name      string `json:"name"`
		Username  string `json:"username"`
	} `json:"resources"`
}

type CFUsers struct {
	Resources []struct {
		Entity struct {
			Username string `json:"username"`
		}
		Metadata struct {
			GUID      string `json:"guid"`
			CreatedAt string `json:"created_at"`
		}
	} `json:"resources"`
}

type CFOrgs struct {
	Resources []struct {
		Name      string `json:"name"`
		GUID      string `json:"guid"`
		CreatedAt string `json:"created_at"`
	} `json:"resources"`
}

type CFSpaces struct {
	Resources []struct {
		Entity struct {
			Name string `json:"name"`
		}
		Metadata struct {
			GUID      string `json:"guid"`
			CreatedAt string `json:"created_at"`
		}
	} `json:"resources"`
}

type CustomMetricCredential struct {
	AppID    string `json:"app_id"`
	UserName string `json:"user_name"`
	Password string `json:"password"`
	URL      string `json:"url"`
}

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)

	cfg = config.LoadConfig(t)
	componentName := "Application Scale Suite"

	if cfg.GetArtifactsDirectory() != "" {
		helpers.EnableCFTrace(cfg, componentName)
	}

	RunSpecs(t, componentName)
}

var _ = BeforeSuite(func() {

	fmt.Println("Clearing down existing test orgs/spaces...")
	orgs := getTestOrgs()

	//  DELETING APPS THEN SERVICES THEN USERS IS IMPORTANT
	for _, org := range orgs {
		orgName, orgGuid, spaceName, spaceGuid := getOrgSpaceNamesAndGuids(org)
		if spaceName != "" {
			target := cf.Cf("target", "-o", orgName, "-s", spaceName).Wait(cfg.DefaultTimeoutDuration())
			Expect(target).To(Exit(0), fmt.Sprintf("failed to target %s and %s", orgName, spaceName))

			apps := getApps(orgGuid, spaceGuid, "autoscaler-")
			deleteApps(apps, 0)

			services := getServices(orgGuid, spaceGuid, "autoscaler-")
			deleteServices(services)
		}

		//workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
		//users := getUsers(spaceGuid)
		//users = removeUserFromList(users, setup.RegularUserContext().Username)
		//	deleteUsers(users, cfg.NamePrefix)
		//})

		deleteOrg(org)
	}

	fmt.Println("Clearing down existing test orgs/spaces... Complete")

	setup = workflowhelpers.NewTestSuiteSetup(cfg)
	setup.Setup()

	workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
		if cfg.IsServiceOfferingEnabled() && cfg.ShouldEnableServiceAccess() {
			EnableServiceAccess(cfg, setup.GetOrganizationName())
		}
	})

	if cfg.IsServiceOfferingEnabled() {
		CheckServiceExists(cfg)
	}

	interval = cfg.AggregateInterval

	// #nosec G402
	client = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
			DisableCompression:  true,
			DisableKeepAlives:   true,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.SkipSSLValidation,
			},
		},
		Timeout: 30 * time.Second,
	}

})

var _ = AfterSuite(func() {
	workflowhelpers.AsUser(setup.AdminUserContext(), cfg.DefaultTimeoutDuration(), func() {
		if cfg.IsServiceOfferingEnabled() && cfg.ShouldEnableServiceAccess() {
			DisableServiceAccess(cfg, setup.GetOrganizationName())
		}
	})
	setup.Teardown()
})

func getStartAndEndTime(location *time.Location, offset, duration time.Duration) (time.Time, time.Time) {
	// Since the validation of time could fail if spread over two days and will result in acceptance test failure
	// Need to fix dates in that case.
	startTime := time.Now().In(location).Add(offset).Truncate(time.Minute)
	if startTime.Day() != startTime.Add(duration).Day() {
		startTime = startTime.Add(duration).Truncate(24 * time.Hour)
	}
	endTime := startTime.Add(duration)
	return startTime, endTime
}

func doAPIRequest(req *http.Request) (*http.Response, error) {
	resp, err := client.Do(req)
	return resp, err
}

func CreatePolicyWithAPI(appGUID, policy string) {
	oauthToken := OauthToken(cfg)
	policyURL := fmt.Sprintf("%s%s", cfg.ASApiEndpoint, strings.Replace(PolicyPath, "{appId}", appGUID, -1))
	req, err := http.NewRequest("PUT", policyURL, bytes.NewBuffer([]byte(policy)))
	Expect(err).ShouldNot(HaveOccurred())
	req.Header.Add("Authorization", oauthToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := doAPIRequest(req)
	Expect(err).ShouldNot(HaveOccurred())
	defer resp.Body.Close()
	Expect(resp.StatusCode == 200 || resp.StatusCode == 201).Should(BeTrue())
	Expect([]int{http.StatusOK, http.StatusCreated}).To(ContainElement(resp.StatusCode))
}

func DeletePolicyWithAPI(appGUID string) {
	oauthToken := OauthToken(cfg)
	policyURL := fmt.Sprintf("%s%s", cfg.ASApiEndpoint, strings.Replace(PolicyPath, "{appId}", appGUID, -1))
	req, err := http.NewRequest("DELETE", policyURL, nil)
	Expect(err).ShouldNot(HaveOccurred())
	req.Header.Add("Authorization", oauthToken)

	resp, err := doAPIRequest(req)
	Expect(err).ShouldNot(HaveOccurred())
	defer resp.Body.Close()
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
}

func CreatePolicy(appName, appGUID, policy string) {
	if cfg.IsServiceOfferingEnabled() {
		instanceName = generator.PrefixedRandomName("autoscaler", "service")
		createService := cf.Cf("create-service", cfg.ServiceName, cfg.ServicePlan, instanceName).Wait(cfg.DefaultTimeoutDuration())
		Expect(createService).To(Exit(0), "failed creating service")

		bindService := cf.Cf("bind-service", appName, instanceName, "-c", policy).Wait(cfg.DefaultTimeoutDuration())
		Expect(bindService).To(Exit(0), "failed binding service to app with a policy ")
	} else {
		CreatePolicyWithAPI(appGUID, policy)
	}
}

func DeletePolicy(appName, appGUID string) {
	if cfg.IsServiceOfferingEnabled() {
		unbindService := cf.Cf("unbind-service", appName, instanceName).Wait(cfg.DefaultTimeoutDuration())
		Expect(unbindService).To(Exit(0), "failed unbinding service from app")
		deleteService := cf.Cf("delete-service", instanceName, "-f").Wait(cfg.DefaultTimeoutDuration())
		Expect(deleteService).To(Exit(0))
	} else {
		DeletePolicyWithAPI(appGUID)
	}
}

func getTestOrgs() []string {
	rawOrgs := cf.Cf("curl", "/v3/organizations").Wait(cfg.DefaultTimeoutDuration())
	Expect(rawOrgs).To(Exit(0), "unable to get orgs")

	var orgs CFOrgs
	err := json.Unmarshal(rawOrgs.Out.Contents(), &orgs)
	Expect(err).ShouldNot(HaveOccurred())

	var orgNames []string
	for _, org := range orgs.Resources {
		if strings.HasPrefix(org.Name, cfg.NamePrefix) {
			orgNames = append(orgNames, org.Name)
		}
	}

	return orgNames
}

func getOrgSpaceNamesAndGuids(org string) (string, string, string, string) {
	orgGuidByte := cf.Cf("org", org, "--guid").Wait(cfg.DefaultTimeoutDuration())
	orgGuid := strings.TrimSuffix(string(orgGuidByte.Out.Contents()), "\n")

	rawSpaces := cf.Cf("curl", fmt.Sprintf("/v2/organizations/%s/spaces", orgGuid)).Wait(cfg.DefaultTimeoutDuration())
	Expect(rawSpaces).To(Exit(0), "unable to get spaces")
	var spaces CFSpaces
	err := json.Unmarshal(rawSpaces.Out.Contents(), &spaces)
	Expect(err).ShouldNot(HaveOccurred())
	if len(spaces.Resources) == 0 {
		return org, orgGuid, "", ""
	}

	return org, orgGuid, spaces.Resources[0].Entity.Name, spaces.Resources[0].Metadata.GUID
}

func getServices(orgGuid, spaceGuid string, prefix string) []string {
	var services CFResourceObject
	rawServices := cf.Cf("curl", "/v3/service_instances?space_guids="+spaceGuid+"&organization_guids="+orgGuid).Wait(cfg.DefaultTimeoutDuration())
	Expect(rawServices).To(Exit(0), "unable to get services")
	err := json.Unmarshal(rawServices.Out.Contents(), &services)
	Expect(err).ShouldNot(HaveOccurred())

	var names []string
	for _, service := range services.Resources {
		if strings.HasPrefix(service.Name, prefix) {
			names = append(names, service.Name)
		}
	}

	return names
}

func getUsers(spaceGuid string) CFUsers {
	var users CFUsers
	rawUsers := cf.Cf("curl", "/v2/users?q=managed_space_guid:"+spaceGuid).Wait(cfg.DefaultTimeoutDuration())
	Expect(rawUsers).To(Exit(0), "unable to get users")
	err := json.Unmarshal(rawUsers.Out.Contents(), &users)
	Expect(err).ShouldNot(HaveOccurred())
	return users
}

func getApps(orgGuid, spaceGuid string, prefix string) []string {
	var apps CFResourceObject
	rawApps := cf.Cf("curl", "/v3/apps?space_guids="+spaceGuid+"&organization_guids="+orgGuid).Wait(cfg.DefaultTimeoutDuration())
	Expect(rawApps).To(Exit(0), "unable to get apps")
	err := json.Unmarshal(rawApps.Out.Contents(), &apps)
	Expect(err).ShouldNot(HaveOccurred())

	var names []string
	for _, app := range apps.Resources {
		if strings.HasPrefix(app.Name, prefix) {
			names = append(names, app.Name)
		}
	}

	return names
}

func deleteServices(services []string) {
	for _, service := range services {
		deleteService := cf.Cf("delete-service", service, "-f").Wait(cfg.DefaultTimeoutDuration())
		Expect(deleteService).To(Exit(0), fmt.Sprintf("unable to delete service %s", service))
	}
}

func deleteOrg(org string) {
	deleteOrg := cf.Cf("delete-org", org, "-f").Wait(cfg.DefaultTimeoutDuration())
	Expect(deleteOrg).To(Exit(0), fmt.Sprintf("unable to delete org %s", org))
}

func deleteUsers(users CFUsers, prefix string) {
	for _, res := range users.Resources {
		username := res.Entity.Username
		if strings.HasPrefix(username, prefix) {
			deleteUser := cf.Cf("delete-user", username, "-f").Wait(cfg.DefaultTimeoutDuration())
			Expect(deleteUser).To(Exit(0), "unable to delete user")
		}
	}
}

func sortAppsByCreatedAt(apps CFResourceObject) (map[string]string, []string) {
	appList := make(map[string]string)
	var keys []string

	for _, res := range apps.Resources {
		appList[res.CreatedAt] = res.Name
		keys = append(keys, res.CreatedAt)
	}
	// Sort by date
	sort.Strings(keys)

	return appList, keys
}

func deleteApps(apps []string, threshold int) {
	for _, app := range apps {
		deleteApp := cf.Cf("delete", app, "-f").Wait(cfg.DefaultTimeoutDuration())
		Expect(deleteApp).To(Exit(0), fmt.Sprintf("unable to delete app %s", app))
	}
	// appList, keys := sortAppsByCreatedAt(apps)

	// numDelete := len(keys) - threshold
	// if numDelete > 0 {
	// 	for ind, key := range keys {
	// 		if ind == numDelete {
	// 			break
	// 		}
	// 		if strings.HasPrefix(appList[key], prefix) {
	// 			deleteApp := cf.Cf("delete", appList[key], "-f").Wait(cfg.DefaultTimeoutDuration())
	// 			Expect(deleteApp).To(Exit(0), "unable to delete app")
	// 		}
	// 	}
	// }
}

func removeUserFromList(users CFUsers, name string) CFUsers {
	for i := range users.Resources {
		if users.Resources[i].Entity.Username == name {
			users.Resources = append(users.Resources[:i], users.Resources[i+1:]...)
			break
		}
	}
	return users
}

func CreateCustomMetricCred(appName, appGUID string) {
	oauthToken := OauthToken(cfg)
	customMetricURL := fmt.Sprintf("%s%s", cfg.ASApiEndpoint, strings.Replace(CustomMetricPath, "{appId}", appGUID, -1))
	req, err := http.NewRequest("PUT", customMetricURL, nil)
	Expect(err).ShouldNot(HaveOccurred())
	req.Header.Add("Authorization", oauthToken)

	resp, err := doAPIRequest(req)
	Expect(err).ShouldNot(HaveOccurred())
	defer resp.Body.Close()
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	Expect(err).NotTo(HaveOccurred())
	setEnv := cf.Cf("set-env", appName, CustomMetricCredEnv, string(bodyBytes)).Wait(cfg.DefaultTimeoutDuration())
	Expect(setEnv).To(Exit(0), "failed set custom metric credential env")
}
func DeleteCustomMetricCred(appGUID string) {
	oauthToken := OauthToken(cfg)
	customMetricURL := fmt.Sprintf("%s%s", cfg.ASApiEndpoint, strings.Replace(CustomMetricPath, "{appId}", appGUID, -1))
	req, err := http.NewRequest("DELETE", customMetricURL, nil)
	Expect(err).ShouldNot(HaveOccurred())
	req.Header.Add("Authorization", oauthToken)

	resp, err := doAPIRequest(req)
	Expect(err).ShouldNot(HaveOccurred())
	defer resp.Body.Close()
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
}
