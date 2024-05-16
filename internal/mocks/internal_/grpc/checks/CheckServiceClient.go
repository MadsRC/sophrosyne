// Code generated by mockery v2.43.0. DO NOT EDIT.

package checks

import (
	context "context"

	checks "github.com/madsrc/sophrosyne/internal/grpc/checks"

	grpc "google.golang.org/grpc"

	mock "github.com/stretchr/testify/mock"
)

// MockCheckServiceClient is an autogenerated mock type for the CheckServiceClient type
type MockCheckServiceClient struct {
	mock.Mock
}

type MockCheckServiceClient_Expecter struct {
	mock *mock.Mock
}

func (_m *MockCheckServiceClient) EXPECT() *MockCheckServiceClient_Expecter {
	return &MockCheckServiceClient_Expecter{mock: &_m.Mock}
}

// Check provides a mock function with given fields: ctx, in, opts
func (_m *MockCheckServiceClient) Check(ctx context.Context, in *checks.CheckRequest, opts ...grpc.CallOption) (*checks.CheckResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Check")
	}

	var r0 *checks.CheckResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *checks.CheckRequest, ...grpc.CallOption) (*checks.CheckResponse, error)); ok {
		return rf(ctx, in, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *checks.CheckRequest, ...grpc.CallOption) *checks.CheckResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*checks.CheckResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *checks.CheckRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockCheckServiceClient_Check_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Check'
type MockCheckServiceClient_Check_Call struct {
	*mock.Call
}

// Check is a helper method to define mock.On call
//   - ctx context.Context
//   - in *checks.CheckRequest
//   - opts ...grpc.CallOption
func (_e *MockCheckServiceClient_Expecter) Check(ctx interface{}, in interface{}, opts ...interface{}) *MockCheckServiceClient_Check_Call {
	return &MockCheckServiceClient_Check_Call{Call: _e.mock.On("Check",
		append([]interface{}{ctx, in}, opts...)...)}
}

func (_c *MockCheckServiceClient_Check_Call) Run(run func(ctx context.Context, in *checks.CheckRequest, opts ...grpc.CallOption)) *MockCheckServiceClient_Check_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]grpc.CallOption, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(grpc.CallOption)
			}
		}
		run(args[0].(context.Context), args[1].(*checks.CheckRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockCheckServiceClient_Check_Call) Return(_a0 *checks.CheckResponse, _a1 error) *MockCheckServiceClient_Check_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockCheckServiceClient_Check_Call) RunAndReturn(run func(context.Context, *checks.CheckRequest, ...grpc.CallOption) (*checks.CheckResponse, error)) *MockCheckServiceClient_Check_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockCheckServiceClient creates a new instance of MockCheckServiceClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockCheckServiceClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockCheckServiceClient {
	mock := &MockCheckServiceClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
