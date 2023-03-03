// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"
	"time"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/envelopeprocessor"
	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/eventgenerator/client"
	"code.cloudfoundry.org/lager/v3"
)

type FakeLogCacheClientCreator struct {
	NewLogCacheClientStub        func(lager.Logger, func() time.Time, envelopeprocessor.EnvelopeProcessor, string) client.MetricClient
	newLogCacheClientMutex       sync.RWMutex
	newLogCacheClientArgsForCall []struct {
		arg1 lager.Logger
		arg2 func() time.Time
		arg3 envelopeprocessor.EnvelopeProcessor
		arg4 string
	}
	newLogCacheClientReturns struct {
		result1 client.MetricClient
	}
	newLogCacheClientReturnsOnCall map[int]struct {
		result1 client.MetricClient
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeLogCacheClientCreator) NewLogCacheClient(arg1 lager.Logger, arg2 func() time.Time, arg3 envelopeprocessor.EnvelopeProcessor, arg4 string) client.MetricClient {
	fake.newLogCacheClientMutex.Lock()
	ret, specificReturn := fake.newLogCacheClientReturnsOnCall[len(fake.newLogCacheClientArgsForCall)]
	fake.newLogCacheClientArgsForCall = append(fake.newLogCacheClientArgsForCall, struct {
		arg1 lager.Logger
		arg2 func() time.Time
		arg3 envelopeprocessor.EnvelopeProcessor
		arg4 string
	}{arg1, arg2, arg3, arg4})
	stub := fake.NewLogCacheClientStub
	fakeReturns := fake.newLogCacheClientReturns
	fake.recordInvocation("NewLogCacheClient", []interface{}{arg1, arg2, arg3, arg4})
	fake.newLogCacheClientMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeLogCacheClientCreator) NewLogCacheClientCallCount() int {
	fake.newLogCacheClientMutex.RLock()
	defer fake.newLogCacheClientMutex.RUnlock()
	return len(fake.newLogCacheClientArgsForCall)
}

func (fake *FakeLogCacheClientCreator) NewLogCacheClientCalls(stub func(lager.Logger, func() time.Time, envelopeprocessor.EnvelopeProcessor, string) client.MetricClient) {
	fake.newLogCacheClientMutex.Lock()
	defer fake.newLogCacheClientMutex.Unlock()
	fake.NewLogCacheClientStub = stub
}

func (fake *FakeLogCacheClientCreator) NewLogCacheClientArgsForCall(i int) (lager.Logger, func() time.Time, envelopeprocessor.EnvelopeProcessor, string) {
	fake.newLogCacheClientMutex.RLock()
	defer fake.newLogCacheClientMutex.RUnlock()
	argsForCall := fake.newLogCacheClientArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *FakeLogCacheClientCreator) NewLogCacheClientReturns(result1 client.MetricClient) {
	fake.newLogCacheClientMutex.Lock()
	defer fake.newLogCacheClientMutex.Unlock()
	fake.NewLogCacheClientStub = nil
	fake.newLogCacheClientReturns = struct {
		result1 client.MetricClient
	}{result1}
}

func (fake *FakeLogCacheClientCreator) NewLogCacheClientReturnsOnCall(i int, result1 client.MetricClient) {
	fake.newLogCacheClientMutex.Lock()
	defer fake.newLogCacheClientMutex.Unlock()
	fake.NewLogCacheClientStub = nil
	if fake.newLogCacheClientReturnsOnCall == nil {
		fake.newLogCacheClientReturnsOnCall = make(map[int]struct {
			result1 client.MetricClient
		})
	}
	fake.newLogCacheClientReturnsOnCall[i] = struct {
		result1 client.MetricClient
	}{result1}
}

func (fake *FakeLogCacheClientCreator) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.newLogCacheClientMutex.RLock()
	defer fake.newLogCacheClientMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeLogCacheClientCreator) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ client.LogCacheClientCreator = new(FakeLogCacheClientCreator)
