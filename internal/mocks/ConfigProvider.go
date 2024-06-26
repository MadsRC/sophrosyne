// Code generated by mockery v2.43.1. DO NOT EDIT.

package sophrosyne

import (
	sophrosyne "github.com/madsrc/sophrosyne"
	mock "github.com/stretchr/testify/mock"
)

// MockConfigProvider is an autogenerated mock type for the ConfigProvider type
type MockConfigProvider struct {
	mock.Mock
}

type MockConfigProvider_Expecter struct {
	mock *mock.Mock
}

func (_m *MockConfigProvider) EXPECT() *MockConfigProvider_Expecter {
	return &MockConfigProvider_Expecter{mock: &_m.Mock}
}

// Get provides a mock function with given fields:
func (_m *MockConfigProvider) Get() *sophrosyne.Config {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 *sophrosyne.Config
	if rf, ok := ret.Get(0).(func() *sophrosyne.Config); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*sophrosyne.Config)
		}
	}

	return r0
}

// MockConfigProvider_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type MockConfigProvider_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
func (_e *MockConfigProvider_Expecter) Get() *MockConfigProvider_Get_Call {
	return &MockConfigProvider_Get_Call{Call: _e.mock.On("Get")}
}

func (_c *MockConfigProvider_Get_Call) Run(run func()) *MockConfigProvider_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockConfigProvider_Get_Call) Return(_a0 *sophrosyne.Config) *MockConfigProvider_Get_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockConfigProvider_Get_Call) RunAndReturn(run func() *sophrosyne.Config) *MockConfigProvider_Get_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockConfigProvider creates a new instance of MockConfigProvider. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockConfigProvider(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockConfigProvider {
	mock := &MockConfigProvider{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
