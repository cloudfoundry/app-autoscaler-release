package app

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type logrusErrorWriter struct{ *logrus.Logger }

func (w logrusErrorWriter) Write(p []byte) (n int, err error) {
	w.Error(string(p))
	return len(p), nil
}

func Router(logger *logrus.Logger, sleep func(duration time.Duration), useMem func(useMb uint64)) *gin.Engine {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"name": "test-app"}) })
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	MemoryTests(logger, r.Group("/memory"), sleep, useMem)
	return r
}

func New(logger *logrus.Logger, address string) *http.Server {
	return &http.Server{
		Addr:         address,
		Handler:      Router(logger, nil, nil),
		ReadTimeout:  5 * time.Second,
		IdleTimeout:  2 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorLog:     log.New(logrusErrorWriter{logger}, "", 0),
	}
}
