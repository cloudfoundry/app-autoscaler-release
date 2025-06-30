package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/cf"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/configutil"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/db"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/helpers"
)

var (
	ErrReadYaml                    = helpers.ErrReadYaml
	ErrScalingEngineConfigNotFound = errors.New("scalingengine config service not found")
)

const (
	DefaultHttpClientTimeout = 5 * time.Second
)

type SynchronizerConfig struct {
	ActiveScheduleSyncInterval time.Duration `yaml:"active_schedule_sync_interval"`
}

type Config struct {
	CF                  cf.Config                    `yaml:"cf"`
	Logging             helpers.LoggingConfig        `yaml:"logging"`
	Server              helpers.ServerConfig         `yaml:"server"`
	CFServer            helpers.ServerConfig         `yaml:"cf_server"`
	Health              helpers.HealthConfig         `yaml:"health"`
	Db                  map[string]db.DatabaseConfig `yaml:"db" json:"db"`
	DefaultCoolDownSecs int                          `yaml:"defaultCoolDownSecs"`
	LockSize            int                          `yaml:"lockSize"`
	HttpClientTimeout   time.Duration                `yaml:"http_client_timeout"`
}

// SetLoggingLevel implements configutil.Configurable
func (c *Config) SetLoggingLevel() {
	c.Logging.Level = strings.ToLower(c.Logging.Level)
}


func defaultConfig() Config {
	return Config{
		CF: cf.Config{
			ClientConfig: cf.ClientConfig{SkipSSLValidation: false},
		},
		Logging: helpers.LoggingConfig{
			Level: "info",
		},
		CFServer: helpers.ServerConfig{
			Port: 8082,
		},
		Server: helpers.ServerConfig{
			Port: 8080,
		},
		Health: helpers.HealthConfig{
			ServerConfig: helpers.ServerConfig{
				Port: 8081,
			},
		},
		Db:                  make(map[string]db.DatabaseConfig),
		DefaultCoolDownSecs: 300,
		LockSize:            100,
		HttpClientTimeout:   DefaultHttpClientTimeout,
	}
}

func LoadConfig(filepath string, vcapReader configutil.VCAPConfigurationReader) (*Config, error) {
	return configutil.GenericLoadConfig(filepath, vcapReader, defaultConfig, configutil.VCAPConfigurableFunc[Config](LoadVcapConfig))
}

func LoadVcapConfig(conf *Config, vcapReader configutil.VCAPConfigurationReader) error {
	if !vcapReader.IsRunningOnCF() {
		return nil
	}

	// enable plain text logging. See src/autoscaler/helpers/logger.go
	conf.Logging.PlainTextSink = true

	// Avoid port conflict: assign actual port to CF server, set BOSH server port to 0 (unused)
	conf.CFServer.Port = vcapReader.GetPort()
	conf.Server.Port = 0

	if err := configutil.LoadConfig(conf, vcapReader, "scalingengine-config"); err != nil {
		return err
	}

	if err := vcapReader.ConfigureDatabases(&conf.Db, nil, ""); err != nil {
		return err
	}

	conf.CFServer.XFCC.ValidSpaceGuid = vcapReader.GetSpaceGuid()
	conf.CFServer.XFCC.ValidOrgGuid = vcapReader.GetOrgGuid()

	return nil
}

func (c *Config) Validate() error {
	err := c.CF.Validate()
	if err != nil {
		return err
	}

	if c.Db[db.PolicyDb].URL == "" {
		return fmt.Errorf("Configuration error: db.policy_db.url is empty")
	}

	if c.Db[db.ScalingEngineDb].URL == "" {
		return fmt.Errorf("Configuration error: db.scalingengine_db.url is empty")
	}

	if c.Db[db.SchedulerDb].URL == "" {
		return fmt.Errorf("Configuration error: db.scheduler_db.url is empty")
	}

	if c.DefaultCoolDownSecs < 60 || c.DefaultCoolDownSecs > 3600 {
		return fmt.Errorf("Configuration error: DefaultCoolDownSecs should be between 60 and 3600")
	}

	if c.LockSize <= 0 {
		return fmt.Errorf("Configuration error: LockSize is less than or equal to 0")
	}

	if c.HttpClientTimeout <= time.Duration(0) {
		return fmt.Errorf("Configuration error: http_client_timeout is less-equal than 0")
	}

	if err := c.Health.Validate(); err != nil {
		return err
	}

	return nil
}
