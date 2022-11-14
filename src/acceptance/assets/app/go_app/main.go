package main

import (
	"acceptance/assets/app/go_app/internal/app"
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	address := os.Getenv("SERVER_ADDRESS") + ":" + getPort(logger)
	logger.Infof("Starting test-app : %s\n", address)
	server := app.New(logger, address)
	enableGracefulShutdown(logger, server)
	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Panicf("Error while exiting server: %s", err.Error())
	}
}

func enableGracefulShutdown(logger *logrus.Logger, server *http.Server) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		// When we get the signal...
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		// ... we gracefully shut down the server.
		// That ensures that no new connections
		err := server.Shutdown(ctx)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("Error closing server: %s", err.Error())
			err = server.Close()
			if err != nil {
				logger.Errorf("Error while forcefully closing: %s", err.Error())
			}
		}
	}()
}

func getPort(logger *logrus.Logger) string {
	port := os.Getenv("PORT")
	if port == "" {
		logger.Infof("No Env var PORT specified using 8080")
		port = "8080"
	}
	return port
}
