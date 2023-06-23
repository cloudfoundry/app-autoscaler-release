package models

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/routes"
	"golang.org/x/crypto/bcrypt"
)

type HealthConfig struct {
	Port                    int      `yaml:"port"`
	HealthCheckUsername     string   `yaml:"username"`
	HealthCheckUsernameHash string   `yaml:"username_hash"`
	HealthCheckPassword     string   `yaml:"password"`
	HealthCheckPasswordHash string   `yaml:"password_hash"`
	ReadinessCheckEnabled   bool     `yaml:"readiness_enabled"`
	UnprotectedEndpoints    []string `yaml:"unprotected_endpoints"`
}

var ErrConfiguration = fmt.Errorf("health configuration error")

func (c *HealthConfig) BasicAuthPossible() bool {
	usernameVerifiable := c.HealthCheckUsername != "" || c.HealthCheckUsernameHash != ""
	passwordVerifiable := c.HealthCheckPassword != "" || c.HealthCheckPasswordHash != ""
	return usernameVerifiable && passwordVerifiable
}

func (c *HealthConfig) Validate() error {
	if c.HealthCheckUsername != "" && c.HealthCheckUsernameHash != "" {
		return fmt.Errorf("%w: both healthcheck username and healthcheck username_hash are set, please provide only one of them", ErrConfiguration)
	}

	if c.HealthCheckPassword != "" && c.HealthCheckPasswordHash != "" {
		return fmt.Errorf("%w: both healthcheck password and healthcheck password_hash are provided, please provide only one of them", ErrConfiguration)
	}

	if c.HealthCheckUsernameHash != "" {
		if _, err := bcrypt.Cost([]byte(c.HealthCheckUsernameHash)); err != nil {
			return fmt.Errorf("%w: healthcheck username_hash is not a valid bcrypt hash", ErrConfiguration)
		}
	}

	if c.HealthCheckPasswordHash != "" {
		if _, err := bcrypt.Cost([]byte(c.HealthCheckPasswordHash)); err != nil {
			return fmt.Errorf("%w: healthcheck password_hash is not a valid bcrypt hash", ErrConfiguration)
		}
	}

	if c.basicAuthIntended() && !c.BasicAuthPossible() {
		protectedHealthEndpoints := c.protectedHealthEndpoints()
		msg :=
			"some endpoints configured to use basic auth but, credentials not properly set up\n" +
				"\tprotected endpoints according to health-configuration: " +
				strings.Join(protectedHealthEndpoints, ", ")
		return fmt.Errorf("%w: %s", ErrConfiguration, msg)
	}

	return nil
}

func (c *HealthConfig) basicAuthIntended() bool {
	return len(c.protectedHealthEndpoints()) > 0
}

func (c *HealthConfig) protectedHealthEndpoints() []string {
	var protectedEndpoints []string

	allEndpointsList := []string{"/", routes.LivenessPath, routes.PrometheusPath, routes.PprofPath}
	if c.ReadinessCheckEnabled {
		allEndpointsList = append(allEndpointsList, routes.ReadinessPath)
	}

	unprotectedEndpointsSet := make(map[string]bool, len(c.UnprotectedEndpoints))
	for _, endpoint := range c.UnprotectedEndpoints {
		unprotectedEndpointsSet[endpoint] = true
	}

	for _, endpoint := range allEndpointsList {
		if _, endpointIsUnprotected := unprotectedEndpointsSet[endpoint]; !endpointIsUnprotected {
			protectedEndpoints = append(protectedEndpoints, endpoint)
		}
	}

	return protectedEndpoints
}
