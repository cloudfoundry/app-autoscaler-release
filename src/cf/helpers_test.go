package cf_test

import (
	"errors"
	"fmt"
	"github.com/onsi/ginkgo/v2"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type ConnectionWatcher struct {
	maxActive   int32
	connections sync.Map
	wrap        func(net.Conn, http.ConnState)
}

func NewConnectionWatcher(wrapping func(net.Conn, http.ConnState)) *ConnectionWatcher {
	return &ConnectionWatcher{wrap: wrapping}
}

// OnStateChange records open connections in response to connection
// state changes. Set net/http Server.ConnState to this method
// as value.
func (cw *ConnectionWatcher) OnStateChange(conn net.Conn, state http.ConnState) {
	switch state {
	case http.StateClosed, http.StateHijacked:
		cw.Remove(conn)
	default:
		cw.Add(conn, state)
	}
	if cw.wrap != nil {
		cw.wrap(conn, state)
	}
}

// Count returns the number of connections at the time
// the call.
func (cw *ConnectionWatcher) GetStates() map[string]int {
	result := map[string]int{}
	cw.connections.Range(func(key, value any) bool {
		state := value.(http.ConnState).String()
		count, ok := result[state]
		if ok {
			result[state] = count + 1
		} else {
			result[state] = 1
		}
		return true
	})

	return result
}

// Add adds c to the number of active connections.
func (cw *ConnectionWatcher) Add(c net.Conn, state http.ConnState) {
	cw.connections.Store(c, state)
	done := false
	for !done {
		prev := atomic.LoadInt32(&cw.maxActive)
		count := cw.Count()
		if count > prev {
			done = atomic.CompareAndSwapInt32(&cw.maxActive, prev, count)
		} else {
			done = true
		}
	}
}

func (cw *ConnectionWatcher) Remove(c net.Conn) {
	cw.connections.Delete(c)
}

func (cw *ConnectionWatcher) MaxOpenConnections() int32 {
	return atomic.LoadInt32(&cw.maxActive)
}
func (cw *ConnectionWatcher) Count() int32 {
	count := int32(0)
	cw.connections.Range(func(key, value any) bool {
		count++
		return true
	})
	return count
}

func (cw *ConnectionWatcher) printStats(title string) {
	GinkgoWriter.Printf("\n# %s\n", title)
	for key, value := range cw.GetStates() {
		GinkgoWriter.Printf("\t%s:\t%d\n", key, value)
	}
}

func LoadFile(filename string) string {
	file, err := os.ReadFile(filename)
	if err != nil {
		file, err = os.ReadFile("testdata/" + filename)
	}
	FailOnError("Could not read file", err)
	return string(file)
}

func FailOnError(message string, err error) {
	if err != nil {
		ginkgo.Fail(fmt.Sprintf("%s: %s", message, err.Error()))
	}
}

func ParseDate(date string) time.Time {
	updated, err := time.Parse(time.RFC3339, date)
	Expect(err).NotTo(HaveOccurred())
	return updated
}

func IsUrlNetOpError(err error) {
	var urlErr *url.Error
	Expect(errors.As(err, &urlErr)).To(BeTrue(), fmt.Sprintf("Expected a (*url.Error) error in the chan got, %T: %+v", err, err))

	var netOpErr *net.OpError
	Expect(errors.As(err, &netOpErr) || errors.Is(err, io.EOF)).
		To(BeTrue(), fmt.Sprintf("Expected a (*net.OpError) or io.EOF error in the chan got, %T: %+v", err, err))
}
