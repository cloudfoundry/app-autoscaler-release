package app

import (
	"net/http"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/go-logr/zapr"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Router(logger *zap.Logger, timewaster TimeWaster, memoryTest MemoryGobbler, cpuTest CPUWaster, customMetricTest CustomMetricClient) *gin.Engine {
	r := gin.New()

	otel.SetTextMapPropagator(b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader)))
	r.Use(otelgin.Middleware("acceptance-tests-go-app"))

	r.Use(ginzap.GinzapWithConfig(logger, &ginzap.Config{TimeFormat: time.RFC3339, UTC: true, TraceID: true}))
	r.Use(ginzap.RecoveryWithZap(logger, true))

	logr := zapr.NewLogger(logger)

	r.GET("/", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"name": "test-app"}) })
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	MemoryTests(logr, r.Group("/memory"), memoryTest)
	ResponseTimeTests(logr, r.Group("/responsetime"), timewaster)
	CPUTests(logr, r.Group("/cpu"), cpuTest)
	CustomMetricsTests(logr, r.Group("/custom-metrics"), customMetricTest)
	return r
}

func New(logger *zap.Logger, address string) *http.Server {
	errorLog, _ := zap.NewStdLogAt(logger, zapcore.ErrorLevel)
	return &http.Server{
		Addr:         address,
		Handler:      Router(logger, &Sleeper{}, &ListBasedMemoryGobbler{}, &ConcurrentBusyLoopCPUWaster{}, &CustomMetricAPIClient{}),
		ReadTimeout:  5 * time.Second,
		IdleTimeout:  2 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorLog:     errorLog,
	}
}
