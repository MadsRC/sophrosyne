// Code generated by mockery v2.43.2. DO NOT EDIT.

package sophrosyne

import mock "github.com/stretchr/testify/mock"

// MockValidator is an autogenerated mock type for the Validator type
type MockValidator struct {
	mock.Mock
}

type MockValidator_Expecter struct {
	mock *mock.Mock
}

func (_m *MockValidator) EXPECT() *MockValidator_Expecter {
	return &MockValidator_Expecter{mock: &_m.Mock}
}

// Validate provides a mock function with given fields: _a0
func (_m *MockValidator) Validate(_a0 interface{}) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Validate")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockValidator_Validate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Validate'
type MockValidator_Validate_Call struct {
	*mock.Call
}

// Validate is a helper method to define mock.On call
//   - _a0 interface{}
func (_e *MockValidator_Expecter) Validate(_a0 interface{}) *MockValidator_Validate_Call {
	return &MockValidator_Validate_Call{Call: _e.mock.On("Validate", _a0)}
}

func (_c *MockValidator_Validate_Call) Run(run func(_a0 interface{})) *MockValidator_Validate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interface{}))
	})
	return _c
}

func (_c *MockValidator_Validate_Call) Return(_a0 error) *MockValidator_Validate_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockValidator_Validate_Call) RunAndReturn(run func(interface{}) error) *MockValidator_Validate_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockValidator creates a new instance of MockValidator. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockValidator(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockValidator {
	mock := &MockValidator{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
