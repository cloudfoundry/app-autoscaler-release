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

//counterfeiter:generate . CPUWaster
type CPUWaster interface {
	UseCPU(utilisation uint64, duration time.Duration)
	IsRunning() bool
	StopTest()
}

type ConcurrentBusyLoopCPUWaster struct {
	mu        sync.Mutex
	isRunning bool
}

var _ CPUWaster = &ConcurrentBusyLoopCPUWaster{}

func CPUTests(logger logr.Logger, r *gin.RouterGroup, cpuTest CPUWaster) *gin.RouterGroup {
	r.GET("/:utilization/:minutes", func(c *gin.Context) {
		if cpuTest.IsRunning() {
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
			cpuTest.UseCPU(utilization, duration)
		}()
		c.JSON(http.StatusOK, gin.H{"utilization": utilization, "minutes": minutes})
	})

	r.GET("/close", func(c *gin.Context) {
		if cpuTest.IsRunning() {
			logger.Info("stop CPU test")
			cpuTest.StopTest()
			c.JSON(http.StatusOK, gin.H{"status": "close cpu test"})
		} else {
			Error(c, http.StatusBadRequest, "CPU test not running")
		}
	})
	return r
}

func (m *ConcurrentBusyLoopCPUWaster) UseCPU(utilisation uint64, duration time.Duration) {
	m.StartTest()

	for utilisation > 0 {
		perGoRoutineUtilisation := min(utilisation, 100)
		utilisation = utilisation - perGoRoutineUtilisation

		go func(util uint64) {
			run := time.Duration(util) * time.Second / 100
			sleep := time.Duration(100-util) * time.Second / 100
			runtime.LockOSThread()
			for m.IsRunning() {
				begin := time.Now()
				for time.Since(begin) < run {
					// burn cpu time
				}
				time.Sleep(sleep)
			}
			runtime.UnlockOSThread()
		}(perGoRoutineUtilisation)
	}

	// how long
	go func() {
		time.Sleep(duration)
		m.StopTest()
	}()
}

func (m *ConcurrentBusyLoopCPUWaster) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.isRunning
}

func (m *ConcurrentBusyLoopCPUWaster) StopTest() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isRunning = false
}

func (m *ConcurrentBusyLoopCPUWaster) StartTest() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isRunning = true
}

func min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}
