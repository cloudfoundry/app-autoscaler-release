package app

import (
	"bytes"
	"container/list"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/prometheus/procfs"
)

const (
	_ = 1 << (10 * iota)
	Kibi
	Mebi
)

type MemTest struct {
	mu        sync.RWMutex
	used      *list.List
	isRunning bool
}

func MemoryTests(logger logr.Logger, r *gin.RouterGroup, sleep func(duration time.Duration), useMem func(useMb uint64)) *gin.RouterGroup {

	var m *MemTest
	if sleep == nil || useMem == nil {
		m = &MemTest{}
		sleep = m.Sleep
		useMem = func(useMb uint64) {
			m.UseMemory(useMb * Mebi)
		}
	}

	r.GET("/:memoryMiB/:minutes", func(c *gin.Context) {
		if m != nil && m.IsRunning() {
			Error(c, http.StatusConflict, "memory test is already running")
			return
		}
		var memoryMiB uint64
		var minutes uint64
		var err error
		memoryMiB, err = strconv.ParseUint(c.Param("memoryMiB"), 10, 64)
		if err != nil {
			Error(c, http.StatusBadRequest, "invalid memoryMiB: %s", err.Error())
			return
		}
		if minutes, err = strconv.ParseUint(c.Param("minutes"), 10, 64); err != nil {
			Error(c, http.StatusBadRequest, "invalid minutes: %s", err.Error())
			return
		}
		duration := time.Duration(minutes) * time.Minute
		logger := logger.WithValues("memoryMiB", memoryMiB, "duration", duration)
		go func() {
			logMemoryUsage(logger, "before memory test")
			useMem(memoryMiB)
			logMemoryUsage(logger, "after allocating memory")
			sleep(duration)
		}()
		c.JSON(http.StatusOK, gin.H{"memoryMiB": memoryMiB, "minutes": minutes})
	})

	r.GET("/close", func(c *gin.Context) {
		if m != nil && m.IsRunning() {
			logger.Info("stop mem test")
			m.StopTest()
			logMemoryUsage(logger, "after freeing memory")
			c.JSON(http.StatusOK, gin.H{"status": "close memory test"})
		} else {
			Error(c, http.StatusBadRequest, "memory test not running")
		}
	})
	return r
}

func logMemoryUsage(logger logr.Logger, action string) {
	logger = logger.WithValues("action", action)
	memoryUsage, err := getTotalMemoryUsage()
	if err == nil {
		logger.Info("memory usage", "usage", memoryUsage/Mebi)
	} else {
		logger.Error(err, "could not determine memory usage")
	}
}

func Error(c *gin.Context, status int, descriptionf string, args ...any) {
	c.JSON(status, gin.H{"error": gin.H{"description": fmt.Sprintf(descriptionf, args...)}})
}

const chunkSize = 4 * Kibi

func (m *MemTest) UseMemory(numBytes uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isRunning = true
	m.used = list.New()
	used := uint64(0)
	for used <= numBytes {
		m.used.PushBack(bytes.Repeat([]byte("X"), chunkSize)) // The bytes need to be non-zero to force memory allocation
		used += chunkSize
	}
}

func (m *MemTest) Sleep(sleepTime time.Duration) {
	sleepTill := time.Now().Add(sleepTime)
	for m.IsRunning() && time.Now().Before(sleepTill) {
		time.Sleep(100 * time.Millisecond)
	}
}

func (m *MemTest) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isRunning
}

func (m *MemTest) StopTest() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isRunning = false
	m.used.Init()
	runtime.GC()
}

func getTotalMemoryUsage() (uint64, error) {
	fs, err := procfs.NewFS("/proc")
	if err != nil {
		return 0, err
	}

	proc, err := fs.Proc(os.Getpid())
	if err != nil {
		return 0, err
	}

	stat, err := proc.NewStatus()
	if err != nil {
		return 0, err
	}

	result := stat.VmRSS + stat.VmSwap

	return result, nil
}
