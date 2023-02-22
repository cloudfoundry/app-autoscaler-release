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

// basic authentication credentials struct
type basicAuthenticationMiddleware struct {
	usernameHash []byte
	passwordHash []byte
}

// middleware basic authentication middleware functionality for healthcheck
func (bam *basicAuthenticationMiddleware) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, authOK := r.BasicAuth()

		if !authOK || bcrypt.CompareHashAndPassword(bam.usernameHash, []byte(username)) != nil || bcrypt.CompareHashAndPassword(bam.passwordHash, []byte(password)) != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// NewServerWithBasicAuth open the healthcheck port with basic authentication.
// Make sure that username and password is not empty
func NewServerWithBasicAuth(conf models.HealthConfig, healthCheckers []Checker, logger lager.Logger, gatherer prometheus.Gatherer, time func() time.Time) (ifrit.Runner, error) {
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

func NewHealthRouter(conf models.HealthConfig, healthCheckers []Checker, logger lager.Logger,
	gatherer prometheus.Gatherer, time func() time.Time) (*mux.Router, error) {

	username := conf.HealthCheckUsername
	password := conf.HealthCheckPassword
	usernameHash := conf.HealthCheckUsernameHash
	passwordHash := conf.HealthCheckPasswordHash
	readinessCheckProtected :=
		(username != "" || usernameHash != "") && (password != "" || passwordHash != "")

	if conf.ReadinessCheckEnabled && !readinessCheckProtected {
		// TODO: Clarify if we should log out an error message instead due to backwards compatibility.
		msg := "Readiness checks intended in configuration but not all parameters provided for basic auth!"
		return nil, fmt.Errorf("%s", msg)
	}

	return routerWithBasicAuth(conf, healthCheckers, logger, gatherer, time)
}

func routerWithBasicAuth(conf models.HealthConfig, healthCheckers []Checker, logger lager.Logger,
	gatherer prometheus.Gatherer, time func() time.Time) (*mux.Router, error) {

	basicAuthentication, err := createBasicAuthMiddleware(logger, conf.HealthCheckUsernameHash, conf.HealthCheckUsername, conf.HealthCheckPasswordHash, conf.HealthCheckPassword)
	if err != nil {
		return nil, err
	}

	router := mux.NewRouter()

	// unauthenticated paths
	if conf.ReadinessCheckEnabled {
		router.Handle("/health/liveliness", common.VarsFunc(readiness([]Checker{}, time)))
	}

	//authenticated paths
	if conf.ReadinessCheckEnabled {
		addReadinessWithBasicAuth(router, basicAuthentication, healthCheckers, time)
	}
	addPprofWithBasicAuth(router, basicAuthentication)
	addPrometheusWithBasicAuth(router, basicAuthentication, gatherer)

	return router, nil
}

func addReadinessWithBasicAuth(mainRouter *mux.Router, authMiddleware *basicAuthenticationMiddleware,
	healthCheckers []Checker, time func() time.Time) {

	mainRouter.Handle("/health/readiness", common.VarsFunc(readiness(healthCheckers, time)))
	readiness := mainRouter.Path("/health/readiness").Subrouter()
	readiness.Use(authMiddleware.middleware)
}

func addPprofWithBasicAuth(mainRouter *mux.Router, authMiddleware *basicAuthenticationMiddleware) {
	pprofRouter := mainRouter.PathPrefix("/debug/pprof").Subrouter()
	pprofRouter.Use(authMiddleware.middleware)

	pprofRouter.HandleFunc("/cmdline", pprof.Cmdline)
	pprofRouter.HandleFunc("/profile", pprof.Profile)
	pprofRouter.HandleFunc("/symbol", pprof.Symbol)
	pprofRouter.HandleFunc("/trace", pprof.Trace)
	pprofRouter.PathPrefix("").HandlerFunc(pprof.Index)
}

func addPrometheusWithBasicAuth(mainRouter *mux.Router, authMiddleware *basicAuthenticationMiddleware,
	gatherer prometheus.Gatherer) {

	promHandler := promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{})

	everything := mainRouter.PathPrefix("").Subrouter()
	everything.Use(authMiddleware.middleware) // TODO: could be a problem: we don't want /health/liveliness to be protected
	everything.PathPrefix("").Handler(promHandler)
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
