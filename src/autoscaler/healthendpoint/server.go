package healthendpoint

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"time"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/routes"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/models"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/metricsforwarder/server/common"

	"code.cloudfoundry.org/lager/v3"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/http_server"
	"golang.org/x/crypto/bcrypt"
)

// basic authentication credentials struct
type basicAuthenticationMiddleware struct {
	usernameHash []byte
	passwordHash []byte
}

// middleware basic authentication middleware functionality for healthcheck
func (bam *basicAuthenticationMiddleware) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, authOK := r.BasicAuth()

		if !authOK || bcrypt.CompareHashAndPassword(bam.usernameHash, []byte(username)) != nil ||
			bcrypt.CompareHashAndPassword(bam.passwordHash, []byte(password)) != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// NewServerWithBasicAuth open the healthcheck port with basic authentication.
// Make sure that username and password is not empty.
// Parameter `healthCheckers` determines the information provided by the readiness-endpoint.
func NewServerWithBasicAuth(conf models.HealthConfig, healthCheckers []Checker, logger lager.Logger,
	gatherer prometheus.Gatherer, time func() time.Time) (ifrit.Runner, error) {
	healthRouter, err := NewHealthRouterWithBasicAuth(conf, healthCheckers, logger, gatherer, time)
	if err != nil {
		return nil, err
	}
	var addr string
	if os.Getenv("APP_AUTOSCALER_TEST_RUN") == "true" {
		addr = fmt.Sprintf("localhost:%d", conf.Port)
	} else {
		addr = fmt.Sprintf("0.0.0.0:%d", conf.Port)
	}

	logger.Info("new-health-server-basic-auth", lager.Data{"addr": addr})
	return http_server.New(addr, healthRouter), nil
}

// NewHealthRouterWithBasicAuth Make sure that username and password is not empty.
// Parameter `healthCheckers` determines the information provided by the readiness-endpoint.
func NewHealthRouterWithBasicAuth(conf models.HealthConfig, healthCheckers []Checker, logger lager.Logger,
	gatherer prometheus.Gatherer, time func() time.Time) (*mux.Router, error) {
	router := mux.NewRouter()
	authMiddleware, err := createBasicAuthMiddleware(logger, conf.HealthCheckUsernameHash,
		conf.HealthCheckUsername, conf.HealthCheckPasswordHash, conf.HealthCheckPassword)
	if err != nil {
		return nil, err
	}

	err = addLivelinessHandlers(conf, router, time, authMiddleware)
	if err != nil {
		return nil, err
	}

	if conf.ReadinessCheckEnabled {
		err = addReadinessHandler(conf, router, authMiddleware, healthCheckers, time)
		if err != nil {
			return nil, err
		}
	}

	err = addPprofHandlers(conf, router, authMiddleware)
	if err != nil {
		return nil, err
	}

	err = addPrometheusHandler(router, conf, authMiddleware, gatherer)
	if err != nil {
		return nil, err
	}

	return router, nil
}

// Adds liveliness handlers on the paths
// "/" and LIVELINESS_PATH and adds a authentication
// middleware for BasicAuth, for all paths that are not
// in "unprotectedPaths".
//
// Returns an error in case BasicAuth is required but the configuration is not set up properly.
func addLivelinessHandlers(conf models.HealthConfig, mainRouter *mux.Router, time func() time.Time,
	authMiddleware *basicAuthenticationMiddleware) error {
	livenessHandler := common.VarsFunc(readiness([]Checker{}, time))
	livenessRouter := mainRouter.PathPrefix(routes.LivenessPath).Subrouter()

	if endpointsNeedsProtection(routes.LivenessPath, conf) {
		if !conf.BasicAuthPossible() {
			msg := "Basic authentication required for endpoint %s, but credentials not set up properly."
			return fmt.Errorf(msg, routes.LivenessPath)
		}
		livenessRouter.Use(authMiddleware.middleware)
	}
	livenessRouter.Handle("", livenessHandler)

	return nil
}

// Adds a readiness handler on the path READINESS_PATH and adds authentication middleware
// for BasicAuth, if and only if READINESS_PATH is not included in the models.HealthConfig.
//
// Returns an error in case BasicAuth is required but the configuration is not set up properly.
func addReadinessHandler(conf models.HealthConfig, mainRouter *mux.Router,
	authMiddleware *basicAuthenticationMiddleware, healthCheckers []Checker, time func() time.Time,
) error {
	readinessRouter := mainRouter.PathPrefix("/health").Subrouter()
	if endpointsNeedsProtection(routes.ReadinessPath, conf) {
		readinessRouter.Use(authMiddleware.middleware)
		if !conf.BasicAuthPossible() {
			msg := "Basic authentication required for endpoint %s, but credentials not set up properly."
			return fmt.Errorf(msg, routes.ReadinessPath)
		}
	}
	// unauthenticated route
	readinessRouter.Handle("/readiness", common.VarsFunc(readiness(healthCheckers, time)))
	return nil
}

// Adds a pprof handler on the path PPROF_PATH featuring several endpoints.
// Adds authentication middleware for BasiAuth, if and only if PPROF_PATH
// is not excluded in the HealthConfig.
//
// Returns an error in case BasicAuth is required but the configuration is not set up properly.
func addPprofHandlers(conf models.HealthConfig, mainRouter *mux.Router,
	authMiddleware *basicAuthenticationMiddleware) error {
	pprofRouter := mainRouter.PathPrefix(routes.PprofPath).Subrouter()

	if endpointsNeedsProtection(routes.PprofPath, conf) {
		if !conf.BasicAuthPossible() {
			msg := "Basic authentication required for endpoint %s, but credentials not set up properly."
			return fmt.Errorf(msg, routes.PprofPath)
		}
		pprofRouter.Use(authMiddleware.middleware)
	}

	pprofRouter.HandleFunc("/cmdline", pprof.Cmdline)
	pprofRouter.HandleFunc("/profile", pprof.Profile)
	pprofRouter.HandleFunc("/symbol", pprof.Symbol)
	pprofRouter.HandleFunc("/trace", pprof.Trace)
	pprofRouter.PathPrefix("").HandlerFunc(pprof.Index)

	return nil
}

// Adds a prometheus handler on the path PROMETHEUS_PATH and adds authentication middleware
// for BasicAuth, if and only if PROMETHEUS_PATH is not excluded in the HealthConfig.
//
// Returns an error in case BasicAuth is required but the configuration is not set up properly.
func addPrometheusHandler(mainRouter *mux.Router, conf models.HealthConfig,
	authMiddleware *basicAuthenticationMiddleware, gatherer prometheus.Gatherer) error {
	promHandler := promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{})

	prometheusRouter := mainRouter.PathPrefix(routes.PrometheusPath).Subrouter()
	if endpointsNeedsProtection(routes.PrometheusPath, conf) {
		prometheusRouter.Use(authMiddleware.middleware)
		if !conf.BasicAuthPossible() {
			msg := "Basic authentication required for endpoint %s, but credentials not set up properly."
			return fmt.Errorf(msg, routes.PrometheusPath)
		}
	}
	// unauthenticated routes
	// /health/prometheus
	prometheusRouter.Path("").Handler(promHandler)
	// http://<health_server:port>/
	// mainRouter.Path("/").Handler(promHandler)
	return nil
}

func endpointsNeedsProtection(path string, conf models.HealthConfig) bool {
	result := true
	for _, p := range conf.UnprotectedEndpoints {
		if p == path {
			result = false
			break
		}
	}

	return result
}

func createBasicAuthMiddleware(logger lager.Logger, usernameHash string, username string, passwordHash string, password string) (*basicAuthenticationMiddleware, error) {
	usernameHashByte, err := getUserHashBytes(logger, usernameHash, username)
	if err != nil {
		return nil, err
	}

	passwordHashByte, err := getPasswordHashBytes(logger, passwordHash, password)
	if err != nil {
		return nil, err
	}

	basicAuthentication := &basicAuthenticationMiddleware{
		usernameHash: usernameHashByte,
		passwordHash: passwordHashByte,
	}
	return basicAuthentication, nil
}

func getPasswordHashBytes(logger lager.Logger, passwordHash string, password string) ([]byte, error) {
	var passwordHashByte []byte
	var err error
	if passwordHash == "" {
		if len(password) > 72 {
			logger.Error("warning-configured-password-too-long-using-only-first-72-characters", bcrypt.ErrPasswordTooLong, lager.Data{"password-length": len(password)})
			password = password[:72]
		}
		passwordHashByte, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost) // use MinCost as the config already provided it as cleartext
		if err != nil {
			logger.Error("failed-new-server-password", err)
			return nil, err
		}
	} else {
		passwordHashByte = []byte(passwordHash)
	}
	return passwordHashByte, nil
}

func getUserHashBytes(logger lager.Logger, usernameHash string, username string) ([]byte, error) {
	var usernameHashByte []byte
	var err error
	if usernameHash == "" {
		if len(username) > 72 {
			logger.Error("warning-configured-username-too-long-using-only-first-72-characters", bcrypt.ErrPasswordTooLong, lager.Data{"username-length": len(username)})
			username = username[:72]
		}
		// when username and password are set for health check
		usernameHashByte, err = bcrypt.GenerateFromPassword([]byte(username), bcrypt.MinCost) // use MinCost as the config already provided it as cleartext
		if err != nil {
			logger.Error("failed-new-server-username", err)
			return nil, err
		}
	} else {
		usernameHashByte = []byte(usernameHash)
	}
	return usernameHashByte, err
}
