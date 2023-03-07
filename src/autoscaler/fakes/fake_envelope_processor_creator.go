// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"
	"time"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/envelopeprocessor"
	lager "code.cloudfoundry.org/lager/v3"
)

type FakeEnvelopeProcessorCreator struct {
	NewProcessorStub        func(lager.Logger, time.Duration) envelopeprocessor.Processor
	newProcessorMutex       sync.RWMutex
	newProcessorArgsForCall []struct {
		arg1 lager.Logger
		arg2 time.Duration
	}
	newProcessorReturns struct {
		result1 envelopeprocessor.Processor
	}
	newProcessorReturnsOnCall map[int]struct {
		result1 envelopeprocessor.Processor
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeEnvelopeProcessorCreator) NewProcessor(arg1 lager.Logger, arg2 time.Duration) envelopeprocessor.Processor {
	fake.newProcessorMutex.Lock()
	ret, specificReturn := fake.newProcessorReturnsOnCall[len(fake.newProcessorArgsForCall)]
	fake.newProcessorArgsForCall = append(fake.newProcessorArgsForCall, struct {
		arg1 lager.Logger
		arg2 time.Duration
	}{arg1, arg2})
	stub := fake.NewProcessorStub
	fakeReturns := fake.newProcessorReturns
	fake.recordInvocation("NewProcessor", []interface{}{arg1, arg2})
	fake.newProcessorMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeEnvelopeProcessorCreator) NewProcessorCallCount() int {
	fake.newProcessorMutex.RLock()
	defer fake.newProcessorMutex.RUnlock()
	return len(fake.newProcessorArgsForCall)
}

func (fake *FakeEnvelopeProcessorCreator) NewProcessorCalls(stub func(lager.Logger, time.Duration) envelopeprocessor.Processor) {
	fake.newProcessorMutex.Lock()
	defer fake.newProcessorMutex.Unlock()
	fake.NewProcessorStub = stub
}

func (fake *FakeEnvelopeProcessorCreator) NewProcessorArgsForCall(i int) (lager.Logger, time.Duration) {
	fake.newProcessorMutex.RLock()
	defer fake.newProcessorMutex.RUnlock()
	argsForCall := fake.newProcessorArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeEnvelopeProcessorCreator) NewProcessorReturns(result1 envelopeprocessor.Processor) {
	fake.newProcessorMutex.Lock()
	defer fake.newProcessorMutex.Unlock()
	fake.NewProcessorStub = nil
	fake.newProcessorReturns = struct {
		result1 envelopeprocessor.Processor
	}{result1}
}

func (fake *FakeEnvelopeProcessorCreator) NewProcessorReturnsOnCall(i int, result1 envelopeprocessor.Processor) {
	fake.newProcessorMutex.Lock()
	defer fake.newProcessorMutex.Unlock()
	fake.NewProcessorStub = nil
	if fake.newProcessorReturnsOnCall == nil {
		fake.newProcessorReturnsOnCall = make(map[int]struct {
			result1 envelopeprocessor.Processor
		})
	}
	fake.newProcessorReturnsOnCall[i] = struct {
		result1 envelopeprocessor.Processor
	}{result1}
}

func (fake *FakeEnvelopeProcessorCreator) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.newProcessorMutex.RLock()
	defer fake.newProcessorMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeEnvelopeProcessorCreator) recordInvocation(key string, args []interface{}) {
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

var _ envelopeprocessor.EnvelopeProcessorCreator = new(FakeEnvelopeProcessorCreator)
