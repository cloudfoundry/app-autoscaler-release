package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-logr/zapr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// JSONResponse represents a JSON response structure
type JSONResponse map[string]interface{}

// writeJSON writes a JSON response with the given status code
func writeJSON(w http.ResponseWriter, statusCode int, data JSONResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// loggingMiddleware provides request logging functionality
func loggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap ResponseWriter to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			fields := []zapcore.Field{
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
			}

			// Log CF ID
			if requestID := r.Header.Get("X-Vcap-Request-Id"); requestID != "" {
				fields = append(fields, zap.String("vcap_request_id", requestID))
			}
			if passport := r.Header.Get("SAP-PASSPORT"); passport != "" {
				fields = append(fields, zap.String("sap_passport", passport))
			}

			// Support OpenTelemetry trace ID
			if span := trace.SpanFromContext(r.Context()); span.SpanContext().IsValid() {
				fields = append(fields, zap.String("w3c_trace-id", span.SpanContext().TraceID().String()))
			}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			fields = append(fields,
				zap.Int("status_code", wrapped.statusCode),
				zap.Duration("duration", duration),
			)

			logger.Info("HTTP request", fields...)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// recoveryMiddleware provides panic recovery functionality
func recoveryMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("Panic recovered",
						zap.Any("error", err),
						zap.String("stack", string(debug.Stack())),
						zap.String("path", r.URL.Path),
						zap.String("method", r.Method),
					)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// otelMiddleware provides OpenTelemetry tracing
func otelMiddleware(next http.Handler) http.Handler {
	tracer := otel.Tracer("acceptance-tests-go-app")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		ctx, span := tracer.Start(ctx, fmt.Sprintf("%s %s", r.Method, r.URL.Path))
		defer span.End()

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func Router(logger *zap.Logger, timewaster TimeWaster, memoryTest MemoryGobbler,
	cpuTest CPUWaster, diskOccupier DiskOccupier, customMetricTest CustomMetricClient) http.Handler {

	mux := http.NewServeMux()

	// Set up OpenTelemetry
	otel.SetTracerProvider(sdktrace.NewTracerProvider())
	otel.SetTextMapPropagator(propagation.TraceContext{})

	logr := zapr.NewLogger(logger)

	// Root routes
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, JSONResponse{"name": "test-app"})
	})

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, JSONResponse{"status": "ok"})
	})

	// Register test endpoints
	MemoryTests(logr, mux, memoryTest)
	ResponseTimeTests(mux, timewaster)
	CPUTests(logr, mux, cpuTest)
	DiskTest(mux, diskOccupier)
	CustomMetricsTests(logr, mux, customMetricTest)

	// Apply middleware in order: recovery -> logging -> otel -> router
	var handler http.Handler = mux
	handler = otelMiddleware(handler)
	handler = loggingMiddleware(logger)(handler)
	handler = recoveryMiddleware(logger)(handler)

	return handler
}

func New(logger *zap.Logger, address string) *http.Server {
	errorLog, _ := zap.NewStdLogAt(logger, zapcore.ErrorLevel)
	return &http.Server{
		Addr: address,
		Handler: Router(
			logger,
			&Sleeper{},
			&ListBasedMemoryGobbler{},
			&ConcurrentBusyLoopCPUWaster{},
			NewDefaultDiskOccupier("this-file-is-being-used-during-disk-occupation"),
			&CustomMetricAPIClient{},
		),
		ReadTimeout:  5 * time.Second,
		IdleTimeout:  2 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorLog:     errorLog,
	}
}
