// Code generated by mockery v2.43.2. DO NOT EDIT.

package v0

import mock "github.com/stretchr/testify/mock"

// MockisGetCheckRequest_Check is an autogenerated mock type for the isGetCheckRequest_Check type
type MockisGetCheckRequest_Check struct {
	mock.Mock
}

type MockisGetCheckRequest_Check_Expecter struct {
	mock *mock.Mock
}

func (_m *MockisGetCheckRequest_Check) EXPECT() *MockisGetCheckRequest_Check_Expecter {
	return &MockisGetCheckRequest_Check_Expecter{mock: &_m.Mock}
}

// isGetCheckRequest_Check provides a mock function with given fields:
func (_m *MockisGetCheckRequest_Check) isGetCheckRequest_Check() {
	_m.Called()
}

// MockisGetCheckRequest_Check_isGetCheckRequest_Check_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'isGetCheckRequest_Check'
type MockisGetCheckRequest_Check_isGetCheckRequest_Check_Call struct {
	*mock.Call
}

// isGetCheckRequest_Check is a helper method to define mock.On call
func (_e *MockisGetCheckRequest_Check_Expecter) isGetCheckRequest_Check() *MockisGetCheckRequest_Check_isGetCheckRequest_Check_Call {
	return &MockisGetCheckRequest_Check_isGetCheckRequest_Check_Call{Call: _e.mock.On("isGetCheckRequest_Check")}
}

func (_c *MockisGetCheckRequest_Check_isGetCheckRequest_Check_Call) Run(run func()) *MockisGetCheckRequest_Check_isGetCheckRequest_Check_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockisGetCheckRequest_Check_isGetCheckRequest_Check_Call) Return() *MockisGetCheckRequest_Check_isGetCheckRequest_Check_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockisGetCheckRequest_Check_isGetCheckRequest_Check_Call) RunAndReturn(run func()) *MockisGetCheckRequest_Check_isGetCheckRequest_Check_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockisGetCheckRequest_Check creates a new instance of MockisGetCheckRequest_Check. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockisGetCheckRequest_Check(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockisGetCheckRequest_Check {
	mock := &MockisGetCheckRequest_Check{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}