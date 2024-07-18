package helpers

import (
	"fmt"
	"net/http"
	"os"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/models"
	"code.cloudfoundry.org/lager/v3"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/http_server"
)

type ServerConfig struct {
	Port      int              `yaml:"port"`
	BasicAuth models.BasicAuth `yaml:"basic_auth"`
}

func NewHTTPServer(logger lager.Logger, conf ServerConfig, handler http.Handler) (ifrit.Runner, error) {
	var addr string
	if os.Getenv("APP_AUTOSCALER_TEST_RUN") == "true" {
		addr = fmt.Sprintf("localhost:%d", conf.Port)
	} else {
		addr = fmt.Sprintf("0.0.0.0:%d", conf.Port)
	}

	logger.Info("new-http-server", lager.Data{"serverConfig": conf})
	return http_server.New(addr, handler), nil
}
