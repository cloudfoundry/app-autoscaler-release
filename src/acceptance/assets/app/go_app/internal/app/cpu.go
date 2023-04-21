package app

import (
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"golang.org/x/exp/constraints"
)

type CPUTest struct {
	mu        sync.Mutex
	isRunning bool
}

func CPUTests(logger logr.Logger, r *gin.RouterGroup, sleep func(duration time.Duration), useCPU func(utilization uint64, duration time.Duration)) *gin.RouterGroup {

	var m *CPUTest
	if sleep == nil || useCPU == nil {
		m = &CPUTest{}
		sleep = m.Sleep
		useCPU = func(utilization uint64, duration time.Duration) {
			m.UseCPU(utilization, duration)
		}
	}

	r.GET("/:utilization/:minutes", func(c *gin.Context) {
		if m != nil && m.IsRunning() {
			Error(c, http.StatusConflict, "CPU test is already running")
			return
		}
		var utilization uint64
		var minutes uint64
		var err error
		utilization, err = strconv.ParseUint(c.Param("utilization"), 10, 64)
		if err != nil {
			Error(c, http.StatusBadRequest, "invalid utilization: %s", err.Error())
			return
		}
		if minutes, err = strconv.ParseUint(c.Param("minutes"), 10, 64); err != nil {
			Error(c, http.StatusBadRequest, "invalid minutes: %s", err.Error())
			return
		}
		duration := time.Duration(minutes) * time.Minute
		go func() {
			useCPU(utilization, duration)
		}()
		c.JSON(http.StatusOK, gin.H{"utilization": utilization, "minutes": minutes})
	})

	r.GET("/close", func(c *gin.Context) {
		if m != nil && m.IsRunning() {
			logger.Info("stop CPU test")
			m.StopTest()
			c.JSON(http.StatusOK, gin.H{"status": "close cpu test"})
		} else {
			Error(c, http.StatusBadRequest, "CPU test not running")
		}
	})
	return r
}

func (m *CPUTest) UseCPU(utilisation uint64, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isRunning = true

	for utilisation > 0 {
		perProcessUtilisation := min(utilisation, 100)
		utilisation = utilisation - perProcessUtilisation

		go func(util uint64) {
			run := time.Duration(util) * time.Microsecond / 10
			sleep := time.Duration(100-util) * time.Microsecond / 10
			runtime.LockOSThread()
			for m.IsRunning() {
				begin := time.Now()
				for time.Now().Sub(begin) < run {
					// burn cpu time
				}
				time.Sleep(sleep)
			}
		}(perProcessUtilisation)
	}
	// how long
	go func() {
		time.Sleep(duration)
		m.StopTest()
	}()
}

func (m *CPUTest) Sleep(sleepTime time.Duration) {
	sleepTill := time.Now().Add(sleepTime)
	for m.IsRunning() && time.Now().Before(sleepTill) {
		time.Sleep(100 * time.Millisecond)
	}
}

func (m *CPUTest) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.isRunning
}

func (m *CPUTest) StopTest() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isRunning = false
}

func min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}
