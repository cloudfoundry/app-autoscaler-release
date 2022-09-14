package cf_test

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"sync/atomic"
	"testing"
)

var noOpHandler = func(_ http.ResponseWriter, _ *http.Request) {
	// empty function for Nop
}

func RespondWithMultiple(handlers ...http.HandlerFunc) http.HandlerFunc {
	var responseNumber int32 = 0
	if len(handlers) > 0 {
		return func(w http.ResponseWriter, req *http.Request) {
			responseNum := atomic.LoadInt32(&responseNumber)
			handlerNumber := Min(responseNum, int32(len(handlers)-1))
			handlers[handlerNumber](w, req)
			atomic.AddInt32(&responseNumber, 1)
		}
	}
	return noOpHandler
}

func RoundRobinWithMultiple(handlers ...http.HandlerFunc) http.HandlerFunc {
	var responseNumber int32 = 0

	if len(handlers) > 0 {
		return func(w http.ResponseWriter, req *http.Request) {
			handlerNumber := atomic.LoadInt32(&responseNumber) % int32(len(handlers))
			handlers[handlerNumber](w, req)
			atomic.AddInt32(&responseNumber, 1)
		}
	}
	return noOpHandler
}

func Min(one, two int32) int32 {
	if one < two {
		return one
	}
	return two
}

// ====================== TESTS

func TestRespondWithMultiple_rollsCorrectly(t *testing.T) {
	value := -1
	handlers := RespondWithMultiple(
		func(resp http.ResponseWriter, req *http.Request) { value = 1 },
		func(resp http.ResponseWriter, req *http.Request) { value = 2 },
		func(resp http.ResponseWriter, req *http.Request) { value = 3 },
		func(resp http.ResponseWriter, req *http.Request) { value = 4 },
		func(resp http.ResponseWriter, req *http.Request) { value = 5 },
	)

	handlers(nil, nil)
	assert.Equal(t, 1, value)
	handlers(nil, nil)
	assert.Equal(t, 2, value)
	handlers(nil, nil)
	assert.Equal(t, 3, value)
	handlers(nil, nil)
	assert.Equal(t, 4, value)
	handlers(nil, nil)
	assert.Equal(t, 5, value)
	handlers(nil, nil)
	assert.Equal(t, 5, value)
}

func TestRespondWithMultiple_empty(t *testing.T) {
	handlers := RespondWithMultiple()

	handlers(nil, nil)
}

func TestRoundRobinWithMultiple_rollsCorrectly(t *testing.T) {
	value := -1
	handlers := RoundRobinWithMultiple(
		func(resp http.ResponseWriter, req *http.Request) { value = 1 },
		func(resp http.ResponseWriter, req *http.Request) { value = 2 },
		func(resp http.ResponseWriter, req *http.Request) { value = 3 },
		func(resp http.ResponseWriter, req *http.Request) { value = 4 },
		func(resp http.ResponseWriter, req *http.Request) { value = 5 },
	)

	handlers(nil, nil)
	assert.Equal(t, 1, value)
	handlers(nil, nil)
	assert.Equal(t, 2, value)
	handlers(nil, nil)
	assert.Equal(t, 3, value)
	handlers(nil, nil)
	assert.Equal(t, 4, value)
	handlers(nil, nil)
	assert.Equal(t, 5, value)
	handlers(nil, nil)
	assert.Equal(t, 1, value)
}

func TestRoundRobinWithMultiple_empty(t *testing.T) {
	handlers := RoundRobinWithMultiple()

	handlers(nil, nil)
}
