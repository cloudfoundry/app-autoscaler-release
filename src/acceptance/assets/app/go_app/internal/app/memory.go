package app

import (
	"container/list"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const Megabyte = uint64(1024 * 1024)

type MemTestInfo struct {
	mu        sync.Mutex
	used      *list.List
	isRunning bool
}

var memTest = &MemTestInfo{}

func MemoryTests(logger *logrus.Logger, r *gin.RouterGroup, sleep func(duration time.Duration), useMem func(useMb uint64)) *gin.RouterGroup {
	r.GET("/:memoryMb/:minutes", func(c *gin.Context) {
		var memoryMb uint64
		var minutes uint64
		var err error
		memoryMb, err = strconv.ParseUint(c.Param("memoryMb"), 10, 64)
		if err != nil {
			Error(c, http.StatusBadRequest, "invalid memoryMb: %s", err.Error())
			return
		}
		if minutes, err = strconv.ParseUint(c.Param("minutes"), 10, 64); err != nil {
			Error(c, http.StatusBadRequest, "invalid minutes: %s", err.Error())
			return
		}
		go func() {
			useMem(memoryMb)
			sleep(time.Duration(minutes) * time.Minute)
		}()
		c.JSON(http.StatusOK, gin.H{"memoryMb": memoryMb, "minutes": minutes})
	})
	return r
}

func Error(c *gin.Context, status int, descriptionf string, args ...any) {
	c.JSON(status, gin.H{"error": gin.H{"description": fmt.Sprintf(descriptionf, args...)}})
}

const chunkSize = 4 * 1024

func (m *MemTestInfo) UseMemory(bytes uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isRunning = true
	m.used = list.New()
	memTest.IsRunning()
	used := uint64(0)
	for used <= bytes {
		m.used.PushBack(new([chunkSize]byte))
		used += chunkSize
	}
}

func (m *MemTestInfo) Sleep(sleepTime time.Duration) {
	sleepTill := time.Now().Add(sleepTime)
	for m.IsRunning() && time.Now().Before(sleepTill) {
		time.Sleep(100 * time.Millisecond)
	}
}

func (m *MemTestInfo) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.isRunning
}

func (m *MemTestInfo) StopTest() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isRunning = false
	m.used.Init()
}
