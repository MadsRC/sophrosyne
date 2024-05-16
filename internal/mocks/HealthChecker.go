// Code generated by mockery v2.43.0. DO NOT EDIT.

package sophrosyne

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockHealthChecker is an autogenerated mock type for the HealthChecker type
type MockHealthChecker struct {
	mock.Mock
}

type MockHealthChecker_Expecter struct {
	mock *mock.Mock
}

func (_m *MockHealthChecker) EXPECT() *MockHealthChecker_Expecter {
	return &MockHealthChecker_Expecter{mock: &_m.Mock}
}

// Health provides a mock function with given fields: ctx
func (_m *MockHealthChecker) Health(ctx context.Context) (bool, []byte) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Health")
	}

	var r0 bool
	var r1 []byte
	if rf, ok := ret.Get(0).(func(context.Context) (bool, []byte)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) bool); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context) []byte); ok {
		r1 = rf(ctx)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([]byte)
		}
	}

	return r0, r1
}

// MockHealthChecker_Health_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Health'
type MockHealthChecker_Health_Call struct {
	*mock.Call
}

// Health is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockHealthChecker_Expecter) Health(ctx interface{}) *MockHealthChecker_Health_Call {
	return &MockHealthChecker_Health_Call{Call: _e.mock.On("Health", ctx)}
}

func (_c *MockHealthChecker_Health_Call) Run(run func(ctx context.Context)) *MockHealthChecker_Health_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockHealthChecker_Health_Call) Return(_a0 bool, _a1 []byte) *MockHealthChecker_Health_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockHealthChecker_Health_Call) RunAndReturn(run func(context.Context) (bool, []byte)) *MockHealthChecker_Health_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockHealthChecker creates a new instance of MockHealthChecker. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockHealthChecker(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockHealthChecker {
	mock := &MockHealthChecker{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
