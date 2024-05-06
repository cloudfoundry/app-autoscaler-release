package forwarder

import (
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/helpers"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/metricsforwarder/config"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/models"
	"code.cloudfoundry.org/go-loggregator/v9"
	"code.cloudfoundry.org/lager/v3"
)

type Emitter interface {
	EmitMetric(*models.CustomMetric)
}

type SyslogEmitter struct {
}

type MetronEmitter struct {
	client *loggregator.IngressClient
	logger lager.Logger
}

const METRICS_FORWARDER_ORIGIN = "autoscaler_metrics_forwarder"

func NewMetronEmitter(logger lager.Logger, conf *config.Config) (Emitter, error) {
	tlsConfig, err := loggregator.NewIngressTLSConfig(
		conf.LoggregatorConfig.TLS.CACertFile,
		conf.LoggregatorConfig.TLS.CertFile,
		conf.LoggregatorConfig.TLS.KeyFile,
	)
	if err != nil {
		logger.Error("could-not-create-TLS-config", err, lager.Data{"config": conf})
		return &MetronEmitter{}, err
	}

	client, err := loggregator.NewIngressClient(
		tlsConfig,
		loggregator.WithAddr(conf.LoggregatorConfig.MetronAddress),
		loggregator.WithTag("origin", METRICS_FORWARDER_ORIGIN),
		loggregator.WithLogger(helpers.NewLoggregatorGRPCLogger(logger.Session("metric_forwarder"))),
	)
	if err != nil {
		logger.Error("could-not-create-loggregator-client", err, lager.Data{"config": conf})
		return &MetronEmitter{}, err
	}

	return &MetronEmitter{
		client: client,
		logger: logger,
	}, nil

	return &MetronEmitter{}, nil
}

func hasLoggregatorConfig(conf *config.Config) bool {
	return conf.LoggregatorConfig.MetronAddress != ""
}

func NewSyslogEmitter(logger lager.Logger, conf *config.Config) (Emitter, error) {
	return &SyslogEmitter{}, nil
}

func NewMetricForwarder(logger lager.Logger, conf *config.Config) (Emitter, error) {
	if hasLoggregatorConfig(conf) {
		return NewMetronEmitter(logger, conf)
	} else {
		return NewSyslogEmitter(logger, conf)
	}
}

func (mf *SyslogEmitter) EmitMetric(metric *models.CustomMetric) {
}

func (mf *MetronEmitter) EmitMetric(metric *models.CustomMetric) {
	mf.logger.Debug("custom-metric-emit-request-received", lager.Data{"metric": metric})

	options := []loggregator.EmitGaugeOption{
		loggregator.WithGaugeAppInfo(metric.AppGUID, int(metric.InstanceIndex)),
		loggregator.WithGaugeValue(metric.Name, metric.Value, metric.Unit),
	}
	mf.client.EmitGauge(options...)
}
