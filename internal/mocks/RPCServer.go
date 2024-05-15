// Code generated by mockery v2.39.1. DO NOT EDIT.

package sophrosyne

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockRPCServer is an autogenerated mock type for the RPCServer type
type MockRPCServer struct {
	mock.Mock
}

type MockRPCServer_Expecter struct {
	mock *mock.Mock
}

func (_m *MockRPCServer) EXPECT() *MockRPCServer_Expecter {
	return &MockRPCServer_Expecter{mock: &_m.Mock}
}

// HandleRPCRequest provides a mock function with given fields: ctx, req
func (_m *MockRPCServer) HandleRPCRequest(ctx context.Context, req []byte) ([]byte, error) {
	ret := _m.Called(ctx, req)

	if len(ret) == 0 {
		panic("no return value specified for HandleRPCRequest")
	}

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []byte) ([]byte, error)); ok {
		return rf(ctx, req)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []byte) []byte); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []byte) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockRPCServer_HandleRPCRequest_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HandleRPCRequest'
type MockRPCServer_HandleRPCRequest_Call struct {
	*mock.Call
}

// HandleRPCRequest is a helper method to define mock.On call
//   - ctx context.Context
//   - req []byte
func (_e *MockRPCServer_Expecter) HandleRPCRequest(ctx interface{}, req interface{}) *MockRPCServer_HandleRPCRequest_Call {
	return &MockRPCServer_HandleRPCRequest_Call{Call: _e.mock.On("HandleRPCRequest", ctx, req)}
}

func (_c *MockRPCServer_HandleRPCRequest_Call) Run(run func(ctx context.Context, req []byte)) *MockRPCServer_HandleRPCRequest_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]byte))
	})
	return _c
}

func (_c *MockRPCServer_HandleRPCRequest_Call) Return(_a0 []byte, _a1 error) *MockRPCServer_HandleRPCRequest_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockRPCServer_HandleRPCRequest_Call) RunAndReturn(run func(context.Context, []byte) ([]byte, error)) *MockRPCServer_HandleRPCRequest_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockRPCServer creates a new instance of MockRPCServer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockRPCServer(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockRPCServer {
	mock := &MockRPCServer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
