// Code generated by mockery v2.43.2. DO NOT EDIT.

package v0

import mock "github.com/stretchr/testify/mock"

// MockisCheckProviderRequest_Check is an autogenerated mock type for the isCheckProviderRequest_Check type
type MockisCheckProviderRequest_Check struct {
	mock.Mock
}

type MockisCheckProviderRequest_Check_Expecter struct {
	mock *mock.Mock
}

func (_m *MockisCheckProviderRequest_Check) EXPECT() *MockisCheckProviderRequest_Check_Expecter {
	return &MockisCheckProviderRequest_Check_Expecter{mock: &_m.Mock}
}

// isCheckProviderRequest_Check provides a mock function with given fields:
func (_m *MockisCheckProviderRequest_Check) isCheckProviderRequest_Check() {
	_m.Called()
}

// MockisCheckProviderRequest_Check_isCheckProviderRequest_Check_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'isCheckProviderRequest_Check'
type MockisCheckProviderRequest_Check_isCheckProviderRequest_Check_Call struct {
	*mock.Call
}

// isCheckProviderRequest_Check is a helper method to define mock.On call
func (_e *MockisCheckProviderRequest_Check_Expecter) isCheckProviderRequest_Check() *MockisCheckProviderRequest_Check_isCheckProviderRequest_Check_Call {
	return &MockisCheckProviderRequest_Check_isCheckProviderRequest_Check_Call{Call: _e.mock.On("isCheckProviderRequest_Check")}
}

func (_c *MockisCheckProviderRequest_Check_isCheckProviderRequest_Check_Call) Run(run func()) *MockisCheckProviderRequest_Check_isCheckProviderRequest_Check_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockisCheckProviderRequest_Check_isCheckProviderRequest_Check_Call) Return() *MockisCheckProviderRequest_Check_isCheckProviderRequest_Check_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockisCheckProviderRequest_Check_isCheckProviderRequest_Check_Call) RunAndReturn(run func()) *MockisCheckProviderRequest_Check_isCheckProviderRequest_Check_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockisCheckProviderRequest_Check creates a new instance of MockisCheckProviderRequest_Check. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockisCheckProviderRequest_Check(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockisCheckProviderRequest_Check {
	mock := &MockisCheckProviderRequest_Check{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}