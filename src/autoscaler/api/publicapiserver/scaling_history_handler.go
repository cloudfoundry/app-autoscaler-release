package publicapiserver

import (
	"context"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/api/config"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/helpers"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/helpers/apis/scalinghistory"
	"code.cloudfoundry.org/lager/v3"
	"go.opentelemetry.io/otel/trace"
)

var (
	_ = scalinghistory.Handler(&ScalingHistoryHandler{})
	_ = scalinghistory.SecurityHandler(&ScalingHistoryHandler{})
	_ = scalinghistory.SecuritySource(&ScalingHistoryHandler{})
)

type ScalingHistoryHandler struct {
	logger              lager.Logger
	conf                *config.Config
	scalingEngineClient *http.Client
	server              *scalinghistory.Server
	client              *scalinghistory.Client
}

func NewScalingHistoryHandler(logger lager.Logger, conf *config.Config) (*ScalingHistoryHandler, error) {
	seClient, err := helpers.CreateHTTPClient(&conf.ScalingEngine.TLSClientCerts, helpers.DefaultClientConfig(), logger.Session("scaling_client"))
	if err != nil {
		return nil, fmt.Errorf("error creating scaling history HTTP client: %w", err)
	}

	newHandler := &ScalingHistoryHandler{
		logger:              logger.Session("scaling-history-handler"),
		conf:                conf,
		scalingEngineClient: seClient,
	}
	if server, err := scalinghistory.NewServer(newHandler, newHandler); err != nil {
		return nil, fmt.Errorf("error creating ogen scaling history server: %w", err)
	} else {
		newHandler.server = server
	}
	if client, err := scalinghistory.NewClient(conf.ScalingEngine.ScalingEngineUrl, newHandler, scalinghistory.WithClient(seClient)); err != nil {
		return nil, fmt.Errorf("error creating ogen scaling history client: %w", err)
	} else {
		newHandler.client = client
	}

	return newHandler, nil
}

func (h *ScalingHistoryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.server.ServeHTTP(w, r)
}
func (h *ScalingHistoryHandler) NewError(_ context.Context, _ error) *scalinghistory.ErrorResponseStatusCode {
	result := &scalinghistory.ErrorResponseStatusCode{}
	result.SetStatusCode(500)
	result.SetResponse(scalinghistory.ErrorResponse{
		Code:    scalinghistory.NewOptString(http.StatusText(500)),
		Message: scalinghistory.NewOptString("Error retrieving scaling history from scaling engine"),
	})
	return result
}

func (h *ScalingHistoryHandler) HandleBearerAuth(ctx context.Context, operationName string, t scalinghistory.BearerAuth) (context.Context, error) {
	// This handler is a no-op, as this handler shall only be available used behind our own auth middleware
	return ctx, nil
}

func (h *ScalingHistoryHandler) V1AppsGUIDScalingHistoriesGet(ctx context.Context, params scalinghistory.V1AppsGUIDScalingHistoriesGetParams) (*scalinghistory.History, error) {
	traceId := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	logger := h.logger.Session("get-scaling-histories", lager.Data{"x_b3_traceid": traceId, "app_guid": params.GUID})
	logger.Info("start")
	defer logger.Info("end")

	result, err := h.client.V1AppsGUIDScalingHistoriesGet(ctx, params)
	if err != nil {
		logger.Error("get", err)
	}
	return result, err
}

func (h *ScalingHistoryHandler) BearerAuth(_ context.Context, _ string) (scalinghistory.BearerAuth, error) {
	// we are calling the scalingengine server authenticated via mTLS, so no bearer token is necessary
	return scalinghistory.BearerAuth{Token: "none"}, nil
}
