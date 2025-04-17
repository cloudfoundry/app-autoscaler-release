package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/cf"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/configutil"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/db"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/helpers"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/models"
)

var (
	ErrReadYaml               = errors.New("failed to read config file")
	ErrReadJson               = errors.New("failed to read vcap_services json")
	ErrOperatorConfigNotFound = errors.New("operator config service not found")
)

const (
	DefaultLoggingLevel        string = "info"
	DefaultRefreshInterval            = 24 * time.Hour
	DefaultCutoffDuration             = 30 * 24 * time.Hour
	DefaultSyncInterval               = 24 * time.Hour
	DefaultDBLockRetryInterval        = 5 * time.Second
	DefaultDBLockTTL                  = 15 * time.Second
	DefaultHttpClientTimeout          = 5 * time.Second
)

type DbPrunerConfig struct {
	DB              db.DatabaseConfig `yaml:"db"`
	RefreshInterval time.Duration     `yaml:"refresh_interval"`
	CutoffDuration  time.Duration     `yaml:"cutoff_duration"`
}

type DBLockConfig struct {
	LockTTL           time.Duration     `yaml:"ttl"`
	LockRetryInterval time.Duration     `yaml:"retry_interval"`
	DB                db.DatabaseConfig `yaml:"db"`
}

var defaultDBLockConfig = DBLockConfig{
	LockTTL:           DefaultDBLockTTL,
	LockRetryInterval: DefaultDBLockRetryInterval,
}

type ScalingEngineConfig struct {
	URL            string          `yaml:"scaling_engine_url"`
	SyncInterval   time.Duration   `yaml:"sync_interval"`
	TLSClientCerts models.TLSCerts `yaml:"tls"`
}

type SchedulerConfig struct {
	URL            string          `yaml:"scheduler_url"`
	SyncInterval   time.Duration   `yaml:"sync_interval"`
	TLSClientCerts models.TLSCerts `yaml:"tls"`
}

type AppSyncerConfig struct {
	DB           db.DatabaseConfig `yaml:"db"`
	SyncInterval time.Duration     `yaml:"sync_interval"`
}

var defaultHealthConfig = helpers.HealthConfig{
	ServerConfig: helpers.ServerConfig{
		Port: 8081,
	},
}

// TODO: move this to a config.Db ?
// app_metrics_db: db: url: <%= app_metrics_db_url %>
// scaling_engine_db: db: url: <%= scaling_engine_db_url %>
// app_syncer: db: url: <%= policy_db_url %>
// db_lock: db: url: <%= lock_db_url %>

type Config struct {
	CF cf.Config `yaml:"cf"`
	// Db                        map[string]db.DatabaseConfig `yaml:"db"`
	Health            helpers.HealthConfig  `yaml:"health"`
	Logging           helpers.LoggingConfig `yaml:"logging"`
	AppMetricsDB      DbPrunerConfig        `yaml:"app_metrics_db"`
	ScalingEngineDB   DbPrunerConfig        `yaml:"scaling_engine_db"`
	ScalingEngine     ScalingEngineConfig   `yaml:"scaling_engine"`
	Scheduler         SchedulerConfig       `yaml:"scheduler"`
	AppSyncer         AppSyncerConfig       `yaml:"app_syncer"`
	DBLock            DBLockConfig          `yaml:"db_lock"`
	HttpClientTimeout time.Duration         `yaml:"http_client_timeout"`
}

func defaultConfig() Config {
	return Config{
		CF: cf.Config{
			ClientConfig: cf.ClientConfig{SkipSSLValidation: false},
		},
		Health:  defaultHealthConfig,
		Logging: helpers.LoggingConfig{Level: DefaultLoggingLevel},
		AppMetricsDB: DbPrunerConfig{
			RefreshInterval: DefaultRefreshInterval,
			CutoffDuration:  DefaultCutoffDuration,
		},
		ScalingEngineDB: DbPrunerConfig{
			RefreshInterval: DefaultRefreshInterval,
			CutoffDuration:  DefaultCutoffDuration,
		},
		ScalingEngine: ScalingEngineConfig{
			SyncInterval: DefaultSyncInterval,
		},
		Scheduler: SchedulerConfig{
			SyncInterval: DefaultSyncInterval,
		},
		AppSyncer: AppSyncerConfig{
			SyncInterval: DefaultSyncInterval,
		},
		DBLock:            defaultDBLockConfig,
		HttpClientTimeout: DefaultHttpClientTimeout,
	}
}

func LoadConfig(filepath string, vcapReader configutil.VCAPConfigurationReader) (*Config, error) {
	conf := defaultConfig()
	if err := loadYamlFile(filepath, &conf); err != nil {
		return nil, err
	}

	if err := loadVcapConfig(&conf, vcapReader); err != nil {
		return nil, err
	}

	return &conf, nil
}

func loadVcapConfig(conf *Config, vcapReader configutil.VCAPConfigurationReader) error {
	return nil
}

func loadYamlFile(filepath string, conf *Config) error {
	if filepath == "" {
		return nil
	}
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Fprintf(os.Stdout, "failed to open config file '%s': %s\n", filepath, err)
		return ErrReadYaml
	}
	defer file.Close()

	dec := yaml.NewDecoder(file)
	dec.KnownFields(true)
	if err := dec.Decode(conf); err != nil {
		return fmt.Errorf("%w: %v", ErrReadYaml, err)
	}

	return nil
}
func (c *Config) Validate() error {
	if c.AppMetricsDB.DB.URL == "" {
		return fmt.Errorf("Configuration error: app_metrics_db.db.url is empty")
	}

	if c.AppMetricsDB.RefreshInterval <= 0 {
		return fmt.Errorf("Configuration error: app_metrics_db.refresh_interval is less than or equal to 0")
	}

	if c.AppMetricsDB.CutoffDuration <= 0 {
		return fmt.Errorf("Configuration error: app_metrics_db.cutoff_duration is less than or equal to 0")
	}

	if c.ScalingEngineDB.DB.URL == "" {
		return fmt.Errorf("Configuration error: scaling_engine_db.db.url is empty")
	}

	if c.ScalingEngineDB.RefreshInterval <= 0 {
		return fmt.Errorf("Configuration error: scaling_engine_db.refresh_interval is less than or equal to 0")
	}

	if c.ScalingEngineDB.CutoffDuration <= 0 {
		return fmt.Errorf("Configuration error: scaling_engine_db.cutoff_duration is less than or equal to 0")
	}
	if c.ScalingEngine.URL == "" {
		return fmt.Errorf("Configuration error: scaling_engine.scaling_engine_url is empty")
	}
	if c.ScalingEngine.SyncInterval <= 0 {
		return fmt.Errorf("Configuration error: scaling_engine.sync_interval is less than or equal to 0")
	}
	if c.Scheduler.URL == "" {
		return fmt.Errorf("Configuration error: scheduler.scheduler_url is empty")
	}
	if c.Scheduler.SyncInterval <= 0 {
		return fmt.Errorf("Configuration error: scheduler.sync_interval is less than or equal to 0")
	}

	if c.DBLock.DB.URL == "" {
		return fmt.Errorf("Configuration error: db_lock.db.url is empty")
	}

	if c.AppSyncer.DB.URL == "" {
		return fmt.Errorf("Configuration error: appSyncer.db.url is empty")
	}
	if c.AppSyncer.SyncInterval <= 0 {
		return fmt.Errorf("Configuration error: appSyncer.sync_interval is less than or equal to 0")
	}

	if c.HttpClientTimeout <= time.Duration(0) {
		return fmt.Errorf("Configuration error: http_client_timeout is less-equal than 0")
	}

	if err := c.Health.Validate(); err != nil {
		return err
	}

	return nil
}
