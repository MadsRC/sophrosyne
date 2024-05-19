// Code generated by mockery v2.43.1. DO NOT EDIT.

package rpc

import (
	context "context"

	jsonrpc "github.com/madsrc/sophrosyne/internal/rpc/jsonrpc"
	mock "github.com/stretchr/testify/mock"
)

// MockService is an autogenerated mock type for the Service type
type MockService struct {
	mock.Mock
}

type MockService_Expecter struct {
	mock *mock.Mock
}

func (_m *MockService) EXPECT() *MockService_Expecter {
	return &MockService_Expecter{mock: &_m.Mock}
}

// EntityID provides a mock function with given fields:
func (_m *MockService) EntityID() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for EntityID")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockService_EntityID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'EntityID'
type MockService_EntityID_Call struct {
	*mock.Call
}

// EntityID is a helper method to define mock.On call
func (_e *MockService_Expecter) EntityID() *MockService_EntityID_Call {
	return &MockService_EntityID_Call{Call: _e.mock.On("EntityID")}
}

func (_c *MockService_EntityID_Call) Run(run func()) *MockService_EntityID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockService_EntityID_Call) Return(_a0 string) *MockService_EntityID_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockService_EntityID_Call) RunAndReturn(run func() string) *MockService_EntityID_Call {
	_c.Call.Return(run)
	return _c
}

// EntityType provides a mock function with given fields:
func (_m *MockService) EntityType() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for EntityType")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockService_EntityType_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'EntityType'
type MockService_EntityType_Call struct {
	*mock.Call
}

// EntityType is a helper method to define mock.On call
func (_e *MockService_Expecter) EntityType() *MockService_EntityType_Call {
	return &MockService_EntityType_Call{Call: _e.mock.On("EntityType")}
}

func (_c *MockService_EntityType_Call) Run(run func()) *MockService_EntityType_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockService_EntityType_Call) Return(_a0 string) *MockService_EntityType_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockService_EntityType_Call) RunAndReturn(run func() string) *MockService_EntityType_Call {
	_c.Call.Return(run)
	return _c
}

// InvokeMethod provides a mock function with given fields: ctx, req
func (_m *MockService) InvokeMethod(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	ret := _m.Called(ctx, req)

	if len(ret) == 0 {
		panic("no return value specified for InvokeMethod")
	}

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, jsonrpc.Request) ([]byte, error)); ok {
		return rf(ctx, req)
	}
	if rf, ok := ret.Get(0).(func(context.Context, jsonrpc.Request) []byte); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, jsonrpc.Request) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockService_InvokeMethod_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'InvokeMethod'
type MockService_InvokeMethod_Call struct {
	*mock.Call
}

// InvokeMethod is a helper method to define mock.On call
//   - ctx context.Context
//   - req jsonrpc.Request
func (_e *MockService_Expecter) InvokeMethod(ctx interface{}, req interface{}) *MockService_InvokeMethod_Call {
	return &MockService_InvokeMethod_Call{Call: _e.mock.On("InvokeMethod", ctx, req)}
}

func (_c *MockService_InvokeMethod_Call) Run(run func(ctx context.Context, req jsonrpc.Request)) *MockService_InvokeMethod_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(jsonrpc.Request))
	})
	return _c
}

func (_c *MockService_InvokeMethod_Call) Return(_a0 []byte, _a1 error) *MockService_InvokeMethod_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockService_InvokeMethod_Call) RunAndReturn(run func(context.Context, jsonrpc.Request) ([]byte, error)) *MockService_InvokeMethod_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockService creates a new instance of MockService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockService {
	mock := &MockService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
