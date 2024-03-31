package app

import (
	"errors"
	"os"
	"sync"
	"time"
)

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
	if du.isRunning {
		return errors.New("disk space is already being occupied")
	}
	du.mu.RUnlock()

	return nil
}

func (du *defaultDiskOccupier) occupy(space int64) error {
	du.mu.Lock()
	file, err := os.Create(du.filePath)
	if err != nil {
		return err
	}
	if err := file.Truncate(space); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	du.isRunning = true
	du.mu.Unlock()

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
	if err := os.Remove(du.filePath); err == nil {
		du.isRunning = false
	}
	du.mu.Unlock()
}
