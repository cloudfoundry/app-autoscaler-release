package healthendpoint

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"time"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/models"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/metricsforwarder/server/common"

	"code.cloudfoundry.org/lager"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/http_server"
	"golang.org/x/crypto/bcrypt"
)

const (
	LIVELINESS_PATH string = "/health/liveliness"
	READINESS_PATH  string = "/health/readiness"
	PPROF_PATH      string = "/debug/pprof"
	PROMETHEUS_PATH string = "/health/prometheus"
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
// Make sure that username and password is not empty
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

	livelinessHandler := common.VarsFunc(readiness([]Checker{}, time))
	paths := []string{"/", LIVELINESS_PATH}
	for _, path := range paths {
		mainRouter.Handle(path, livelinessHandler)
		if endpointsNeedsProtection(path, conf) {
			if !conf.BasicAuthPossible() {
				msg := "Basic authentication required for endpoint %s, but credentials not set up properly."
				return fmt.Errorf(msg, path)
			}
			sr := mainRouter.Path(path).Subrouter()
			sr.Use(authMiddleware.middleware)
		}
	}

	return nil
}

// Adds a readiness handler on the path READINESS_PATH and adds authentication middleware
// for BasiAuth, if and only if READINESS_PATH is not excluded in the HealthConfig.
//
// Returns an error in case BasicAuth is required but the configuration is not set up properly.
func addReadinessHandler(conf models.HealthConfig, mainRouter *mux.Router,
	authMiddleware *basicAuthenticationMiddleware, healthCheckers []Checker, time func() time.Time,
) error {

	readinessHandler := common.VarsFunc(readiness(healthCheckers, time))
	path := READINESS_PATH
	mainRouter.Handle(path, readinessHandler)
	if endpointsNeedsProtection(path, conf) {
		if !conf.BasicAuthPossible() {
			msg := "Basic authentication required for endpoint %s, but credentials not set up properly."
			return fmt.Errorf(msg, path)
		}
		readiness := mainRouter.Path("/health/readiness").Subrouter()
		readiness.Use(authMiddleware.middleware)
	}

	return nil
}

// Adds a pprof handler on the path PPROF_PATH featuring several endpoints.
// Adds authentication middleware for BasiAuth, if and only if PPROF_PATH
// is not excluded in the HealthConfig.
//
// Returns an error in case BasicAuth is required but the configuration is not set up properly.
func addPprofHandlers(conf models.HealthConfig, mainRouter *mux.Router,
	authMiddleware *basicAuthenticationMiddleware) error {

	mainPath := PPROF_PATH
	pprofRouter := mainRouter.PathPrefix(mainPath).Subrouter()
	if endpointsNeedsProtection(mainPath, conf) {
		if !conf.BasicAuthPossible() {
			msg := "Basic authentication required for endpoint %s, but credentials not set up properly."
			return fmt.Errorf(msg, mainPath)
		}
		pprofRouter.Use(authMiddleware.middleware)
	}

	pprofRouter.HandleFunc("", pprof.Index)
	pprofRouter.HandleFunc("/cmdline", pprof.Cmdline)
	pprofRouter.HandleFunc("/profile", pprof.Profile)
	pprofRouter.HandleFunc("/symbol", pprof.Symbol)
	pprofRouter.HandleFunc("/trace", pprof.Trace)

	return nil
}

// Adds a prometheus handler on the path PROMETHEUS_PATH and adds authentication middleware
// for BasiAuth, if and only if PROMETHEUS_PATH is not excluded in the HealthConfig.
//
// Returns an error in case BasicAuth is required but the configuration is not set up properly.
func addPrometheusHandler(mainRouter *mux.Router, conf models.HealthConfig,
	authMiddleware *basicAuthenticationMiddleware, gatherer prometheus.Gatherer) error {

	promHandler := promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{})
	path := PROMETHEUS_PATH
	prometheusRouter := mainRouter.PathPrefix(path).Subrouter()
	if endpointsNeedsProtection(path, conf) {
		if !conf.BasicAuthPossible() {
			msg := "Basic authentication required for endpoint %s, but credentials not set up properly."
			return fmt.Errorf(msg, path)
		}
		prometheusRouter.Use(authMiddleware.middleware)
	}
	prometheusRouter.PathPrefix("").Handler(promHandler)

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
