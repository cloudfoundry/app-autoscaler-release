package healthendpoint

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"time"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/models"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/metricsforwarder/server/common"

	"code.cloudfoundry.org/lager/v3"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/http_server"
	"golang.org/x/crypto/bcrypt"
)

const (
	LivenessPath   string = "/health/liveness"
	ReadinessPath  string = "/health/readiness"
	PprofPath      string = "/debug/pprof"
	PrometheusPath string = "/health/prometheus"
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
	healthRouter, err := NewHealthRouter(conf, healthCheckers, logger, gatherer, time)
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

func NewHealthRouter(conf models.HealthConfig, healthCheckers []Checker, logger lager.Logger, gatherer prometheus.Gatherer, time func() time.Time) (*mux.Router, error) {
	var err error

	router := mux.NewRouter()

	// unauthenticated routes
	addLivenessRoute(router, healthCheckers, time)

	authMiddleware, err := createBasicAuthMiddleware(logger, conf.HealthCheckUsernameHash,
		conf.HealthCheckUsername, conf.HealthCheckPasswordHash, conf.HealthCheckPassword)
	if err != nil {
		return nil, err
	}
	// authenticated routes
	addReadinessRoute(conf, healthCheckers, router, time, authMiddleware)
	promHandler := addPrometheusRoute(router, gatherer, authMiddleware)
	addPprofRoutes(router, authMiddleware)

	// anything else should also be protected
	restRoute(router, promHandler, authMiddleware)

	return router, nil
}

func restRoute(router *mux.Router, promHandler http.Handler, authMiddleware *basicAuthenticationMiddleware) {
	restAuthRouter := router.PathPrefix("").Subrouter()
	restAuthRouter.PathPrefix("").Handler(promHandler)
	restAuthRouter.Use(authMiddleware.middleware)
}

func addLivenessRoute(router *mux.Router, healthCheckers []Checker, time func() time.Time) {
	noAuthRouter := router.PathPrefix("/health").Subrouter()
	noAuthRouter.Handle("/liveness", common.VarsFunc(readiness(healthCheckers, time)))
}

// /debug/pprof
func addPprofRoutes(router *mux.Router, authMiddleware *basicAuthenticationMiddleware) {
	pprofRouter := router.PathPrefix("/debug/pprof").Subrouter()
	pprofRouter.Use(authMiddleware.middleware)

	pprofRouter.HandleFunc("/cmdline", pprof.Cmdline)
	pprofRouter.HandleFunc("/profile", pprof.Profile)
	pprofRouter.HandleFunc("/symbol", pprof.Symbol)
	pprofRouter.HandleFunc("/trace", pprof.Trace)
	pprofRouter.PathPrefix("").HandlerFunc(pprof.Index)
}

func addPrometheusRoute(router *mux.Router, gatherer prometheus.Gatherer,
	authMiddleware *basicAuthenticationMiddleware) http.Handler {
	// /health/prometheus
	prometheusAuthRouter := router.PathPrefix("/health/prometheus").Subrouter()
	promHandler := promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{})
	prometheusAuthRouter.Handle("", promHandler)
	prometheusAuthRouter.Use(authMiddleware.middleware)
	return promHandler
}

// /health/readiness
func addReadinessRoute(conf models.HealthConfig, healthCheckers []Checker, router *mux.Router, time func() time.Time,
	authMiddleware *basicAuthenticationMiddleware) {
	readinessAuthRouter := router.PathPrefix("/health/readiness").Subrouter()
	if conf.ReadinessCheckEnabled {
		readinessAuthRouter.Handle("", common.VarsFunc(readiness(healthCheckers, time)))
		readinessAuthRouter.Use(authMiddleware.middleware)
	}
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
