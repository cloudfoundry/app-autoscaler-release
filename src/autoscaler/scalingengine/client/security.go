package client

import (
	"context"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/scalingengine/apis/scalinghistory"
)

type SecuritySource struct {
	Username string
	Password string
}

func (h SecuritySource) BasicAuth(_ context.Context, _ string) (scalinghistory.BasicAuth, error) {
	return scalinghistory.BasicAuth{
		Username: h.Username,
		Password: h.Password,
	}, nil
}
