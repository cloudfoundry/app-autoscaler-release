package liveliness

import (
	"net"
	"net/http"

	"code.cloudfoundry.org/lager"
	"github.com/tedsuo/ifrit"
)

type LivenessConf struct {
	localAddr net.TCPAddr // IP-Address of the interface and port,
												// where the server listens on, e.g.: "0.0.0.0:8080"
	path string // Path where the server receives its requests, e.g. "/liveliness"
	logger lager.Logger
}

func NewServer(conf LivenessConf) (ifrit.Runner, error) {
	panic("Unimplemented!")
}

func NewLivelinessRouter() (http.ServeMux, error) {
	panic("Unimplemented!")
}