package mocks

import (
	. "github.com/cloudfoundry/app-autoscaler-release/cf"
	"net/http"
	"regexp"
	"time"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

type MockServer struct {
	*ghttp.Server
}

func NewMockServer() *MockServer {
	return NewMockWithServer(ghttp.NewServer())
}
func NewMockWithServer(server *ghttp.Server) *MockServer {
	return &MockServer{server}
}

func NewMockTlsServer() *MockServer {
	return &MockServer{ghttp.NewTLSServer()}
}

func (m *MockServer) Add() *AddMock {
	return &AddMock{m}
}

func (m *MockServer) Count() *CountMock {
	return &CountMock{m}
}

type CountMock struct{ server *MockServer }

func (m CountMock) Requests(urlRegExp string) int {
	count := 0
	for _, req := range m.server.ReceivedRequests() {
		found, err := regexp.Match(urlRegExp, []byte(req.RequestURI))
		if err != nil {
			panic(err)
		}
		if found {
			count++
		}
	}
	return count
}

type State struct {
	Current  string `json:"current"`
	Previous string `json:"previous"`
}
type InstanceCount struct {
	Current  int `json:"current"`
	Previous int `json:"previous"`
}

type AddMock struct{ server *MockServer }

func (a AddMock) GetApp(appState string, statusCode int, spaceGuid SpaceId) AddMock {
	created, err := time.Parse(time.RFC3339, "2022-07-21T13:42:30Z")
	Expect(err).NotTo(HaveOccurred())
	updated, err := time.Parse(time.RFC3339, "2022-07-21T14:30:17Z")
	Expect(err).NotTo(HaveOccurred())
	a.server.RouteToHandler("GET",
		regexp.MustCompile(`^/v3/apps/[^/]+$`),
		ghttp.RespondWithJSONEncoded(statusCode, App{
			Guid:      "testing-guid-get-app",
			Name:      "mock-get-app",
			State:     appState,
			CreatedAt: created,
			UpdatedAt: updated,
			Relationships: Relationships{
				Space: &Space{
					Data: SpaceData{
						Guid: spaceGuid,
					},
				},
			},
		}))
	return a
}

func (a AddMock) GetAppProcesses(processes int) AddMock {
	type processesResponse struct {
		Pagination Pagination `json:"pagination"`
		Resources  Processes  `json:"resources"`
	}
	a.server.RouteToHandler("GET",
		regexp.MustCompile(`^/v3/apps/[^/]+/processes$`),
		ghttp.RespondWithJSONEncoded(http.StatusOK, processesResponse{Resources: Processes{{Instances: processes}}}))
	return a
}

func (a AddMock) Info(url string) AddMock {
	a.server.RouteToHandler("GET", "/", ghttp.RespondWithJSONEncoded(http.StatusOK, EndpointsResponse{
		Links: Endpoints{
			Login: Href{Url: url},
			Uaa:   Href{Url: url},
		},
	}))
	return a
}

func (a AddMock) ScaleAppWebProcess() AddMock {
	a.server.RouteToHandler("POST", regexp.MustCompile(`^/v3/apps/[^/]+/processes/web/actions/scale$`), ghttp.RespondWith(http.StatusAccepted, "{}"))
	return a
}

func (a AddMock) Roles(statusCode int, roles ...Role) AddMock {
	a.server.RouteToHandler("GET", "/v3/roles",
		ghttp.RespondWithJSONEncoded(statusCode, Response[Role]{Resources: roles}))
	return a
}

func (a AddMock) ServiceInstance(planGuid string) AddMock {
	a.server.RouteToHandler("GET", regexp.MustCompile(`^/v3/service_instances/[^/]+$`),
		ghttp.RespondWithJSONEncoded(http.StatusOK, &ServiceInstance{
			Guid:          "service-instance-mock-guid",
			Type:          "managed",
			Relationships: ServiceInstanceRelationships{ServicePlan: ServicePlanRelation{Data: ServicePlanData{Guid: planGuid}}},
		}),
	)
	return a
}

func (a AddMock) ServicePlan(brokerPlanId string) AddMock {
	a.server.RouteToHandler("GET", regexp.MustCompile(`^/v3/service_plans/[^/]+$`),
		ghttp.RespondWithJSONEncoded(http.StatusOK, ServicePlan{BrokerCatalog: BrokerCatalog{Id: brokerPlanId}}),
	)
	return a
}

func (a AddMock) UserInfo(statusCode int, testUserId string) AddMock {
	a.server.RouteToHandler(http.MethodGet, "/userinfo",
		ghttp.RespondWithJSONEncoded(statusCode,
			struct {
				UserId string `json:"user_id"`
			}{testUserId}))
	return a
}

func (a AddMock) CheckToken(testUserScope []string) AddMock {
	a.server.RouteToHandler(http.MethodPost, "/check_token",
		ghttp.RespondWithJSONEncoded(http.StatusOK,
			struct {
				Scope []string `json:"scope"`
			}{
				testUserScope,
			}))
	return a
}

func (a AddMock) OauthToken(accessToken string) AddMock {
	a.server.RouteToHandler(http.MethodPost, "/oauth/token",
		ghttp.RespondWithJSONEncoded(http.StatusOK, Tokens{AccessToken: accessToken, ExpiresIn: 12000}))
	return a
}
