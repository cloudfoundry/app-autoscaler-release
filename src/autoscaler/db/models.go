package db

import "time"

type DatabaseConfig struct {
	URL                   string        `yaml:"url" json:"url"`
	MaxOpenConnections    int32         `yaml:"max_open_connections" json:"max_open_connections,omitempty`
	MaxIdleConnections    int32         `yaml:"max_idle_connections" json:"max_idle_connections,omitempty`
	ConnectionMaxLifetime time.Duration `yaml:"connection_max_lifetime" json:"connection_max_lifetime,omitempty`
	ConnectionMaxIdleTime time.Duration `yaml:"connection_max_idletime" json:"connection_max_idletime,omitempty"`
}
