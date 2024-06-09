// Code generated by mockery v2.43.2. DO NOT EDIT.

package v0

import (
	context "context"

	grpc "google.golang.org/grpc"

	mock "github.com/stretchr/testify/mock"

	v0 "github.com/madsrc/sophrosyne/internal/grpc/sophrosyne/v0"
)

// MockScanServiceClient is an autogenerated mock type for the ScanServiceClient type
type MockScanServiceClient struct {
	mock.Mock
}

type MockScanServiceClient_Expecter struct {
	mock *mock.Mock
}

func (_m *MockScanServiceClient) EXPECT() *MockScanServiceClient_Expecter {
	return &MockScanServiceClient_Expecter{mock: &_m.Mock}
}

// Scan provides a mock function with given fields: ctx, in, opts
func (_m *MockScanServiceClient) Scan(ctx context.Context, in *v0.ScanRequest, opts ...grpc.CallOption) (*v0.ScanResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Scan")
	}

	var r0 *v0.ScanResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v0.ScanRequest, ...grpc.CallOption) (*v0.ScanResponse, error)); ok {
		return rf(ctx, in, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v0.ScanRequest, ...grpc.CallOption) *v0.ScanResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v0.ScanResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v0.ScanRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockScanServiceClient_Scan_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Scan'
type MockScanServiceClient_Scan_Call struct {
	*mock.Call
}

// Scan is a helper method to define mock.On call
//   - ctx context.Context
//   - in *v0.ScanRequest
//   - opts ...grpc.CallOption
func (_e *MockScanServiceClient_Expecter) Scan(ctx interface{}, in interface{}, opts ...interface{}) *MockScanServiceClient_Scan_Call {
	return &MockScanServiceClient_Scan_Call{Call: _e.mock.On("Scan",
		append([]interface{}{ctx, in}, opts...)...)}
}

func (_c *MockScanServiceClient_Scan_Call) Run(run func(ctx context.Context, in *v0.ScanRequest, opts ...grpc.CallOption)) *MockScanServiceClient_Scan_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]grpc.CallOption, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(grpc.CallOption)
			}
		}
		run(args[0].(context.Context), args[1].(*v0.ScanRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockScanServiceClient_Scan_Call) Return(_a0 *v0.ScanResponse, _a1 error) *MockScanServiceClient_Scan_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockScanServiceClient_Scan_Call) RunAndReturn(run func(context.Context, *v0.ScanRequest, ...grpc.CallOption) (*v0.ScanResponse, error)) *MockScanServiceClient_Scan_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockScanServiceClient creates a new instance of MockScanServiceClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockScanServiceClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockScanServiceClient {
	mock := &MockScanServiceClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}