// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/eventgenerator/client"
	clienta "code.cloudfoundry.org/go-log-cache"
	"google.golang.org/grpc"
)

type FakeGoLogCacheClient struct {
	NewClientStub        func(string, ...clienta.ClientOption) *clienta.Client
	newClientMutex       sync.RWMutex
	newClientArgsForCall []struct {
		arg1 string
		arg2 []clienta.ClientOption
	}
	newClientReturns struct {
		result1 *clienta.Client
	}
	newClientReturnsOnCall map[int]struct {
		result1 *clienta.Client
	}
	NewOauth2HTTPClientStub        func(string, string, string, ...clienta.Oauth2Option) *clienta.Oauth2HTTPClient
	newOauth2HTTPClientMutex       sync.RWMutex
	newOauth2HTTPClientArgsForCall []struct {
		arg1 string
		arg2 string
		arg3 string
		arg4 []clienta.Oauth2Option
	}
	newOauth2HTTPClientReturns struct {
		result1 *clienta.Oauth2HTTPClient
	}
	newOauth2HTTPClientReturnsOnCall map[int]struct {
		result1 *clienta.Oauth2HTTPClient
	}
	WithHTTPClientStub        func(clienta.HTTPClient) clienta.ClientOption
	withHTTPClientMutex       sync.RWMutex
	withHTTPClientArgsForCall []struct {
		arg1 clienta.HTTPClient
	}
	withHTTPClientReturns struct {
		result1 clienta.ClientOption
	}
	withHTTPClientReturnsOnCall map[int]struct {
		result1 clienta.ClientOption
	}
	WithOauth2HTTPClientStub        func(clienta.HTTPClient) clienta.Oauth2Option
	withOauth2HTTPClientMutex       sync.RWMutex
	withOauth2HTTPClientArgsForCall []struct {
		arg1 clienta.HTTPClient
	}
	withOauth2HTTPClientReturns struct {
		result1 clienta.Oauth2Option
	}
	withOauth2HTTPClientReturnsOnCall map[int]struct {
		result1 clienta.Oauth2Option
	}
	WithViaGRPCStub        func(...grpc.DialOption) clienta.ClientOption
	withViaGRPCMutex       sync.RWMutex
	withViaGRPCArgsForCall []struct {
		arg1 []grpc.DialOption
	}
	withViaGRPCReturns struct {
		result1 clienta.ClientOption
	}
	withViaGRPCReturnsOnCall map[int]struct {
		result1 clienta.ClientOption
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeGoLogCacheClient) NewClient(arg1 string, arg2 ...clienta.ClientOption) *clienta.Client {
	fake.newClientMutex.Lock()
	ret, specificReturn := fake.newClientReturnsOnCall[len(fake.newClientArgsForCall)]
	fake.newClientArgsForCall = append(fake.newClientArgsForCall, struct {
		arg1 string
		arg2 []clienta.ClientOption
	}{arg1, arg2})
	stub := fake.NewClientStub
	fakeReturns := fake.newClientReturns
	fake.recordInvocation("NewClient", []interface{}{arg1, arg2})
	fake.newClientMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2...)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeGoLogCacheClient) NewClientCallCount() int {
	fake.newClientMutex.RLock()
	defer fake.newClientMutex.RUnlock()
	return len(fake.newClientArgsForCall)
}

func (fake *FakeGoLogCacheClient) NewClientCalls(stub func(string, ...clienta.ClientOption) *clienta.Client) {
	fake.newClientMutex.Lock()
	defer fake.newClientMutex.Unlock()
	fake.NewClientStub = stub
}

func (fake *FakeGoLogCacheClient) NewClientArgsForCall(i int) (string, []clienta.ClientOption) {
	fake.newClientMutex.RLock()
	defer fake.newClientMutex.RUnlock()
	argsForCall := fake.newClientArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeGoLogCacheClient) NewClientReturns(result1 *clienta.Client) {
	fake.newClientMutex.Lock()
	defer fake.newClientMutex.Unlock()
	fake.NewClientStub = nil
	fake.newClientReturns = struct {
		result1 *clienta.Client
	}{result1}
}

func (fake *FakeGoLogCacheClient) NewClientReturnsOnCall(i int, result1 *clienta.Client) {
	fake.newClientMutex.Lock()
	defer fake.newClientMutex.Unlock()
	fake.NewClientStub = nil
	if fake.newClientReturnsOnCall == nil {
		fake.newClientReturnsOnCall = make(map[int]struct {
			result1 *clienta.Client
		})
	}
	fake.newClientReturnsOnCall[i] = struct {
		result1 *clienta.Client
	}{result1}
}

func (fake *FakeGoLogCacheClient) NewOauth2HTTPClient(arg1 string, arg2 string, arg3 string, arg4 ...clienta.Oauth2Option) *clienta.Oauth2HTTPClient {
	fake.newOauth2HTTPClientMutex.Lock()
	ret, specificReturn := fake.newOauth2HTTPClientReturnsOnCall[len(fake.newOauth2HTTPClientArgsForCall)]
	fake.newOauth2HTTPClientArgsForCall = append(fake.newOauth2HTTPClientArgsForCall, struct {
		arg1 string
		arg2 string
		arg3 string
		arg4 []clienta.Oauth2Option
	}{arg1, arg2, arg3, arg4})
	stub := fake.NewOauth2HTTPClientStub
	fakeReturns := fake.newOauth2HTTPClientReturns
	fake.recordInvocation("NewOauth2HTTPClient", []interface{}{arg1, arg2, arg3, arg4})
	fake.newOauth2HTTPClientMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4...)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeGoLogCacheClient) NewOauth2HTTPClientCallCount() int {
	fake.newOauth2HTTPClientMutex.RLock()
	defer fake.newOauth2HTTPClientMutex.RUnlock()
	return len(fake.newOauth2HTTPClientArgsForCall)
}

func (fake *FakeGoLogCacheClient) NewOauth2HTTPClientCalls(stub func(string, string, string, ...clienta.Oauth2Option) *clienta.Oauth2HTTPClient) {
	fake.newOauth2HTTPClientMutex.Lock()
	defer fake.newOauth2HTTPClientMutex.Unlock()
	fake.NewOauth2HTTPClientStub = stub
}

func (fake *FakeGoLogCacheClient) NewOauth2HTTPClientArgsForCall(i int) (string, string, string, []clienta.Oauth2Option) {
	fake.newOauth2HTTPClientMutex.RLock()
	defer fake.newOauth2HTTPClientMutex.RUnlock()
	argsForCall := fake.newOauth2HTTPClientArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *FakeGoLogCacheClient) NewOauth2HTTPClientReturns(result1 *clienta.Oauth2HTTPClient) {
	fake.newOauth2HTTPClientMutex.Lock()
	defer fake.newOauth2HTTPClientMutex.Unlock()
	fake.NewOauth2HTTPClientStub = nil
	fake.newOauth2HTTPClientReturns = struct {
		result1 *clienta.Oauth2HTTPClient
	}{result1}
}

func (fake *FakeGoLogCacheClient) NewOauth2HTTPClientReturnsOnCall(i int, result1 *clienta.Oauth2HTTPClient) {
	fake.newOauth2HTTPClientMutex.Lock()
	defer fake.newOauth2HTTPClientMutex.Unlock()
	fake.NewOauth2HTTPClientStub = nil
	if fake.newOauth2HTTPClientReturnsOnCall == nil {
		fake.newOauth2HTTPClientReturnsOnCall = make(map[int]struct {
			result1 *clienta.Oauth2HTTPClient
		})
	}
	fake.newOauth2HTTPClientReturnsOnCall[i] = struct {
		result1 *clienta.Oauth2HTTPClient
	}{result1}
}

func (fake *FakeGoLogCacheClient) WithHTTPClient(arg1 clienta.HTTPClient) clienta.ClientOption {
	fake.withHTTPClientMutex.Lock()
	ret, specificReturn := fake.withHTTPClientReturnsOnCall[len(fake.withHTTPClientArgsForCall)]
	fake.withHTTPClientArgsForCall = append(fake.withHTTPClientArgsForCall, struct {
		arg1 clienta.HTTPClient
	}{arg1})
	stub := fake.WithHTTPClientStub
	fakeReturns := fake.withHTTPClientReturns
	fake.recordInvocation("WithHTTPClient", []interface{}{arg1})
	fake.withHTTPClientMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeGoLogCacheClient) WithHTTPClientCallCount() int {
	fake.withHTTPClientMutex.RLock()
	defer fake.withHTTPClientMutex.RUnlock()
	return len(fake.withHTTPClientArgsForCall)
}

func (fake *FakeGoLogCacheClient) WithHTTPClientCalls(stub func(clienta.HTTPClient) clienta.ClientOption) {
	fake.withHTTPClientMutex.Lock()
	defer fake.withHTTPClientMutex.Unlock()
	fake.WithHTTPClientStub = stub
}

func (fake *FakeGoLogCacheClient) WithHTTPClientArgsForCall(i int) clienta.HTTPClient {
	fake.withHTTPClientMutex.RLock()
	defer fake.withHTTPClientMutex.RUnlock()
	argsForCall := fake.withHTTPClientArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeGoLogCacheClient) WithHTTPClientReturns(result1 clienta.ClientOption) {
	fake.withHTTPClientMutex.Lock()
	defer fake.withHTTPClientMutex.Unlock()
	fake.WithHTTPClientStub = nil
	fake.withHTTPClientReturns = struct {
		result1 clienta.ClientOption
	}{result1}
}

func (fake *FakeGoLogCacheClient) WithHTTPClientReturnsOnCall(i int, result1 clienta.ClientOption) {
	fake.withHTTPClientMutex.Lock()
	defer fake.withHTTPClientMutex.Unlock()
	fake.WithHTTPClientStub = nil
	if fake.withHTTPClientReturnsOnCall == nil {
		fake.withHTTPClientReturnsOnCall = make(map[int]struct {
			result1 clienta.ClientOption
		})
	}
	fake.withHTTPClientReturnsOnCall[i] = struct {
		result1 clienta.ClientOption
	}{result1}
}

func (fake *FakeGoLogCacheClient) WithOauth2HTTPClient(arg1 clienta.HTTPClient) clienta.Oauth2Option {
	fake.withOauth2HTTPClientMutex.Lock()
	ret, specificReturn := fake.withOauth2HTTPClientReturnsOnCall[len(fake.withOauth2HTTPClientArgsForCall)]
	fake.withOauth2HTTPClientArgsForCall = append(fake.withOauth2HTTPClientArgsForCall, struct {
		arg1 clienta.HTTPClient
	}{arg1})
	stub := fake.WithOauth2HTTPClientStub
	fakeReturns := fake.withOauth2HTTPClientReturns
	fake.recordInvocation("WithOauth2HTTPClient", []interface{}{arg1})
	fake.withOauth2HTTPClientMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeGoLogCacheClient) WithOauth2HTTPClientCallCount() int {
	fake.withOauth2HTTPClientMutex.RLock()
	defer fake.withOauth2HTTPClientMutex.RUnlock()
	return len(fake.withOauth2HTTPClientArgsForCall)
}

func (fake *FakeGoLogCacheClient) WithOauth2HTTPClientCalls(stub func(clienta.HTTPClient) clienta.Oauth2Option) {
	fake.withOauth2HTTPClientMutex.Lock()
	defer fake.withOauth2HTTPClientMutex.Unlock()
	fake.WithOauth2HTTPClientStub = stub
}

func (fake *FakeGoLogCacheClient) WithOauth2HTTPClientArgsForCall(i int) clienta.HTTPClient {
	fake.withOauth2HTTPClientMutex.RLock()
	defer fake.withOauth2HTTPClientMutex.RUnlock()
	argsForCall := fake.withOauth2HTTPClientArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeGoLogCacheClient) WithOauth2HTTPClientReturns(result1 clienta.Oauth2Option) {
	fake.withOauth2HTTPClientMutex.Lock()
	defer fake.withOauth2HTTPClientMutex.Unlock()
	fake.WithOauth2HTTPClientStub = nil
	fake.withOauth2HTTPClientReturns = struct {
		result1 clienta.Oauth2Option
	}{result1}
}

func (fake *FakeGoLogCacheClient) WithOauth2HTTPClientReturnsOnCall(i int, result1 clienta.Oauth2Option) {
	fake.withOauth2HTTPClientMutex.Lock()
	defer fake.withOauth2HTTPClientMutex.Unlock()
	fake.WithOauth2HTTPClientStub = nil
	if fake.withOauth2HTTPClientReturnsOnCall == nil {
		fake.withOauth2HTTPClientReturnsOnCall = make(map[int]struct {
			result1 clienta.Oauth2Option
		})
	}
	fake.withOauth2HTTPClientReturnsOnCall[i] = struct {
		result1 clienta.Oauth2Option
	}{result1}
}

func (fake *FakeGoLogCacheClient) WithViaGRPC(arg1 ...grpc.DialOption) clienta.ClientOption {
	fake.withViaGRPCMutex.Lock()
	ret, specificReturn := fake.withViaGRPCReturnsOnCall[len(fake.withViaGRPCArgsForCall)]
	fake.withViaGRPCArgsForCall = append(fake.withViaGRPCArgsForCall, struct {
		arg1 []grpc.DialOption
	}{arg1})
	stub := fake.WithViaGRPCStub
	fakeReturns := fake.withViaGRPCReturns
	fake.recordInvocation("WithViaGRPC", []interface{}{arg1})
	fake.withViaGRPCMutex.Unlock()
	if stub != nil {
		return stub(arg1...)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeGoLogCacheClient) WithViaGRPCCallCount() int {
	fake.withViaGRPCMutex.RLock()
	defer fake.withViaGRPCMutex.RUnlock()
	return len(fake.withViaGRPCArgsForCall)
}

func (fake *FakeGoLogCacheClient) WithViaGRPCCalls(stub func(...grpc.DialOption) clienta.ClientOption) {
	fake.withViaGRPCMutex.Lock()
	defer fake.withViaGRPCMutex.Unlock()
	fake.WithViaGRPCStub = stub
}

func (fake *FakeGoLogCacheClient) WithViaGRPCArgsForCall(i int) []grpc.DialOption {
	fake.withViaGRPCMutex.RLock()
	defer fake.withViaGRPCMutex.RUnlock()
	argsForCall := fake.withViaGRPCArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeGoLogCacheClient) WithViaGRPCReturns(result1 clienta.ClientOption) {
	fake.withViaGRPCMutex.Lock()
	defer fake.withViaGRPCMutex.Unlock()
	fake.WithViaGRPCStub = nil
	fake.withViaGRPCReturns = struct {
		result1 clienta.ClientOption
	}{result1}
}

func (fake *FakeGoLogCacheClient) WithViaGRPCReturnsOnCall(i int, result1 clienta.ClientOption) {
	fake.withViaGRPCMutex.Lock()
	defer fake.withViaGRPCMutex.Unlock()
	fake.WithViaGRPCStub = nil
	if fake.withViaGRPCReturnsOnCall == nil {
		fake.withViaGRPCReturnsOnCall = make(map[int]struct {
			result1 clienta.ClientOption
		})
	}
	fake.withViaGRPCReturnsOnCall[i] = struct {
		result1 clienta.ClientOption
	}{result1}
}

func (fake *FakeGoLogCacheClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.newClientMutex.RLock()
	defer fake.newClientMutex.RUnlock()
	fake.newOauth2HTTPClientMutex.RLock()
	defer fake.newOauth2HTTPClientMutex.RUnlock()
	fake.withHTTPClientMutex.RLock()
	defer fake.withHTTPClientMutex.RUnlock()
	fake.withOauth2HTTPClientMutex.RLock()
	defer fake.withOauth2HTTPClientMutex.RUnlock()
	fake.withViaGRPCMutex.RLock()
	defer fake.withViaGRPCMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeGoLogCacheClient) recordInvocation(key string, args []interface{}) {
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

var _ client.GoLogCacheClient = new(FakeGoLogCacheClient)
