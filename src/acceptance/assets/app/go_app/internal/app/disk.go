package app

import (
	"crypto/rand"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

func DiskTest(r *gin.RouterGroup, diskOccupier DiskOccupier) *gin.RouterGroup {
	r.GET("/:utilisation/:minutes", func(c *gin.Context) {
		var utilisation int64
		var minutes int64
		var err error

		utilisation, err = strconv.ParseInt(c.Param("utilisation"), 10, 64)
		if err != nil {
			Error(c, http.StatusBadRequest, "invalid utilisation: %s", err.Error())
			return
		}
		if minutes, err = strconv.ParseInt(c.Param("minutes"), 10, 64); err != nil {
			Error(c, http.StatusBadRequest, "invalid minutes: %s", err.Error())
			return
		}
		duration := time.Duration(minutes) * time.Minute
		spaceInMB := utilisation * 1000 * 1000
		if err = diskOccupier.Occupy(spaceInMB, duration); err != nil {
			Error(c, http.StatusInternalServerError, "error invoking occupation: %s", err.Error())
			return
		}
		c.JSON(http.StatusOK, gin.H{"utilisation": utilisation, "minutes": minutes})
	})

	r.GET("/close", func(c *gin.Context) {
		diskOccupier.Stop()
		c.String(http.StatusOK, "close disk test")
	})

	return r
}

//counterfeiter:generate . DiskOccupier
type DiskOccupier interface {
	Occupy(space int64, duration time.Duration) error
	Stop()
}

type defaultDiskOccupier struct {
	mu        sync.RWMutex
	isRunning bool
	filePath  string
}

func NewDefaultDiskOccupier(filePath string) *defaultDiskOccupier {
	return &defaultDiskOccupier{
		filePath: filePath,
	}
}

func (du *defaultDiskOccupier) Occupy(space int64, duration time.Duration) error {
	if err := du.checkAlreadyRunning(); err != nil {
		return err
	}

	if err := du.occupy(space); err != nil {
		return err
	}

	du.stopAfter(duration)

	return nil
}

func (du *defaultDiskOccupier) checkAlreadyRunning() error {
	du.mu.RLock()
	defer du.mu.RUnlock()

	if du.isRunning {
		return errors.New("disk space is already being occupied")
	}

	return nil
}

func (du *defaultDiskOccupier) occupy(space int64) error {
	du.mu.Lock()
	defer du.mu.Unlock()

	file, err := os.Create(du.filePath)
	if err != nil {
		return err
	}
	if _, err := io.CopyN(file, io.LimitReader(rand.Reader, space), space); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	du.isRunning = true

	return nil
}

func (du *defaultDiskOccupier) stopAfter(duration time.Duration) {
	go func() {
		time.Sleep(duration)
		du.Stop()
	}()
}

func (du *defaultDiskOccupier) Stop() {
	du.mu.Lock()
	defer du.mu.Unlock()

	if err := os.Remove(du.filePath); err == nil {
		du.isRunning = false
	}
}
