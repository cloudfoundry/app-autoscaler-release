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

	"github.com/go-logr/logr"
	"github.com/prometheus/procfs"
)

const (
	_ = 1 << (10 * iota)
	Kibi
	Mebi
)

//counterfeiter:generate . MemoryGobbler
type MemoryGobbler interface {
	UseMemory(numBytes int64)
	Sleep(sleepTime time.Duration)
	IsRunning() bool
	StopTest()
}

type ListBasedMemoryGobbler struct {
	mu        sync.RWMutex
	used      *list.List
	isRunning bool
}

var _ MemoryGobbler = &ListBasedMemoryGobbler{}

func MemoryTests(logger logr.Logger, mux *http.ServeMux, memoryTest MemoryGobbler) {
	mux.HandleFunc("GET /memory/{memoryMiB}/{minutes}", func(w http.ResponseWriter, r *http.Request) {
		if memoryTest.IsRunning() {
			Error(w, http.StatusConflict, "memory test is already running")
			return
		}
		var memoryMiB int64
		var minutes int64
		var err error
		memoryMiB, err = strconv.ParseInt(r.PathValue("memoryMiB"), 10, 64)
		if err != nil {
			Error(w, http.StatusBadRequest, "invalid memoryMiB: %s", err.Error())
			return
		}
		minutes, err = strconv.ParseInt(r.PathValue("minutes"), 10, 64)
		if err != nil {
			Error(w, http.StatusBadRequest, "invalid minutes: %s", err.Error())
			return
		}
		if memoryMiB < 1 {
			Error(w, http.StatusBadRequest, "memoryMiB must be > 0")
			return
		}
		if minutes < 1 {
			Error(w, http.StatusBadRequest, "minutes must be > 0")
			return
		}
		duration := time.Duration(minutes) * time.Minute
		numBytes := memoryMiB * Mebi
		logger.Info("Starting memory test",
			"memoryMiB", memoryMiB,
			"minutes", minutes,
			"bytes", numBytes)
		go func() {
			memoryTest.UseMemory(numBytes)
			memoryTest.Sleep(duration)
			memoryTest.StopTest()
		}()
		writeJSON(w, http.StatusOK, JSONResponse{
			"memoryMiB": memoryMiB,
			"minutes":   minutes,
		})
	})

	mux.HandleFunc("GET /memory/stop", func(w http.ResponseWriter, r *http.Request) {
		if !memoryTest.IsRunning() {
			Error(w, http.StatusConflict, "memory test is not running")
			return
		}
		memoryTest.StopTest()
		writeJSON(w, http.StatusOK, JSONResponse{"message": "Stopped memory test"})
	})

	mux.HandleFunc("GET /memory/usage", func(w http.ResponseWriter, r *http.Request) {
		pid := os.Getpid()
		proc, err := procfs.NewProc(pid)
		if err != nil {
			Error(w, http.StatusInternalServerError, "failed to get process info: %s", err.Error())
			return
		}
		stat, err := proc.Stat()
		if err != nil {
			Error(w, http.StatusInternalServerError, "failed to get process stats: %s", err.Error())
			return
		}
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		writeJSON(w, http.StatusOK, JSONResponse{
			"rss":         stat.RSS * 4096,
			"vms":         stat.VSize,
			"alloc":       m.Alloc,
			"total_alloc": m.TotalAlloc,
			"sys":         m.Sys,
			"num_gc":      m.NumGC,
		})
	})
}

// Error writes an error response
func Error(w http.ResponseWriter, statusCode int, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	writeJSON(w, statusCode, JSONResponse{"error": JSONResponse{"description": message}})
}

const chunkSize = 4 * Kibi

func (m *ListBasedMemoryGobbler) UseMemory(numBytes int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isRunning = true
	m.used = list.New()
	used := int64(0)
	for used <= numBytes {
		m.used.PushBack(bytes.Repeat([]byte("X"), chunkSize)) // The bytes need to be non-zero to force memory allocation
		used += chunkSize
	}
}

func (m *ListBasedMemoryGobbler) Sleep(sleepTime time.Duration) {
	sleepTill := time.Now().Add(sleepTime)
	for m.IsRunning() && time.Now().Before(sleepTill) {
		time.Sleep(100 * time.Millisecond)
	}
}

func (m *ListBasedMemoryGobbler) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isRunning
}

func (m *ListBasedMemoryGobbler) StopTest() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isRunning = false
	m.used.Init()
	runtime.GC()
}
