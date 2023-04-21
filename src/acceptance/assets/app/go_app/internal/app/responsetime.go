package app

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
)

func ResponseTimeTests(logger logr.Logger, r *gin.RouterGroup, sleep func(duration time.Duration)) *gin.RouterGroup {

	var m *MemTest
	if sleep == nil {
		sleep = Sleep
	}

	r.GET("/slow/:delayInMS", func(c *gin.Context) {
		var milliseconds uint64
		var err error
		if milliseconds, err = strconv.ParseUint(c.Param("delayInMS"), 10, 64); err != nil {
			Error(c, http.StatusBadRequest, "invalid milliseconds: %s", err.Error())
			return
		}
		duration := time.Duration(milliseconds) * time.Millisecond
		sleep(duration)
		c.JSON(http.StatusOK, gin.H{"duration": duration.String()})
	})

	r.GET("/fast", func(c *gin.Context) {
		if m != nil && m.IsRunning() {
			c.JSON(http.StatusOK, gin.H{"fast": true})
		}
	})
	return r
}

func Sleep(sleepTime time.Duration) {
	time.Sleep(sleepTime)
}
