package db

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/url"
	"os"
	"strings"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/configutil"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/models"
	"github.com/go-sql-driver/mysql"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

type Database struct {
	DriverName     string
	DataSourceName string
	OTELAttribute  attribute.KeyValue
}

type MySQLConfig struct {
	config *mysql.Config
	cert   string
}

/*
*

	This function is used to generate db connection info, for example,
	For mysql:
	input dbUrl: 'username:password@tcp(localhost:3306)/autoscaler?tls=custom&sslrootcert=db_ca.crt'
	  return:
	  &Database{DriverName: "mysql", DSN:"username:password@tcp(localhost:3306)/autoscaler?parseTime=true&tls=custom"}

For postgres:

	  input dbUrl: postgres://postgres:password@localhost:5432/autoscaler?sslmode=disable
	  return:
	  &Database{DriverName: "postgres", DSN:"postgres://postgres:password@localhost:5432/autoscaler?sslmode=disable"
	*
*/
func GetConnection(dbUrl string) (*Database, error) {
	database := &Database{}

	database.DriverName = detectDirver(dbUrl)

	switch database.DriverName {
	case MysqlDriverName:
		cfg, err := parseMySQLURL(dbUrl)
		if err != nil {
			return nil, err
		}

		err = registerConfig(cfg)
		if err != nil {
			return nil, err
		}
		database.DataSourceName = cfg.config.FormatDSN()
		database.OTELAttribute = semconv.DBSystemMySQL
	case PostgresDriverName:
		database.DataSourceName = dbUrl
		database.OTELAttribute = semconv.DBSystemPostgreSQL
	}
	return database, nil
}

func registerConfig(cfg *MySQLConfig) error {
	tlsValue := cfg.config.TLSConfig
	if _, isBool := readBool(tlsValue); isBool || tlsValue == "" || strings.ToLower(tlsValue) == "skip-verify" || strings.ToLower(tlsValue) == "preferred" {
		// Do nothing here
		return nil
	} else if cfg.cert != "" {
		certBytes, err := os.ReadFile(cfg.cert)
		if err != nil {
			return err
		}
		caCertPool := x509.NewCertPool()
		if ok := caCertPool.AppendCertsFromPEM(certBytes); !ok {
			return err
		}

		tlsConfig := tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		tlsConfig.RootCAs = caCertPool
		if tlsValue == "verify_identity" {
			tlsConfig.ServerName = strings.Split(cfg.config.Addr, ":")[0]
		}

		err = mysql.RegisterTLSConfig(tlsValue, &tlsConfig)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("sql ca file is not provided")
	}
	return nil
}

func readBool(input string) (value bool, valid bool) {
	switch input {
	case "1", "true", "TRUE", "True":
		return true, true
	case "0", "false", "FALSE", "False":
		return false, true
	}
	return
}

func detectDirver(dbUrl string) (driver string) {
	if strings.Contains(dbUrl, "postgres") {
		return PostgresDriverName
	} else {
		return MysqlDriverName
	}
}

// parseMySQLURL can parse the query parameters and remove invalid 'sslrootcert', it return mysql.Config and the cert file.
func parseMySQLURL(dbUrl string) (*MySQLConfig, error) {
	var caCert string
	var tlsValue string
	if strings.Contains(dbUrl, "?") {
		u, err := url.ParseQuery(strings.Split(dbUrl, "?")[1])
		if err != nil {
			return nil, err
		}
		urlParam := url.Values{}
		for k, v := range u {
			if k == "sslrootcert" {
				caCert = v[0]
				continue
			}
			if k == "tls" {
				tlsValue = v[0]
				continue
			}
			urlParam.Add(k, v[0])
		}
		dbUrl = fmt.Sprintf("%s?%s", strings.Split(dbUrl, "?")[0], urlParam.Encode())
	}

	config, err := mysql.ParseDSN(dbUrl)
	if err != nil {
		return nil, err
	}
	config.ParseTime = true

	if tlsValue != "" {
		config.TLSConfig = tlsValue
	}

	return &MySQLConfig{
		config: config,
		cert:   caCert,
	}, nil
}

func ConfigureStoredProcedureDb(confDb *map[string]DatabaseConfig, storedProcedureConfig *models.StoredProcedureConfig, vcapReader configutil.VCAPConfigurationReader) error {
	if err := ConfigureDb(StoredProcedureDb, confDb, vcapReader); err != nil {
		return err
	}

	currentStoredProcedureDb := (*confDb)[StoredProcedureDb]
	parsedUrl, err := url.Parse(currentStoredProcedureDb.URL)
	if err != nil {
		return err
	}

	if storedProcedureConfig != nil {
		if storedProcedureConfig.Username != "" {
			currentStoredProcedureDb.URL = strings.Replace(currentStoredProcedureDb.URL, parsedUrl.User.Username(), storedProcedureConfig.Username, 1)
		}
		if storedProcedureConfig.Password != "" {
			bindingPassword, _ := parsedUrl.User.Password()
			currentStoredProcedureDb.URL = strings.Replace(currentStoredProcedureDb.URL, bindingPassword, storedProcedureConfig.Password, 1)
		}
	}
	(*confDb)[StoredProcedureDb] = currentStoredProcedureDb

	return nil
}

func ConfigureDb(dbName string, confDb *map[string]DatabaseConfig, vcapReader configutil.VCAPConfigurationReader) error {
	currentDb, ok := (*confDb)[dbName]
	if !ok {
		(*confDb)[dbName] = DatabaseConfig{}
	}

	dbURL, err := vcapReader.MaterializeDBFromService(dbName)
	currentDb.URL = dbURL
	if err != nil {
		return err
	}
	(*confDb)[dbName] = currentDb

	return nil
}
