// Code generated by mockery v2.39.1. DO NOT EDIT.

package jsonrpc

import mock "github.com/stretchr/testify/mock"

// Params is an autogenerated mock type for the Params type
type Params struct {
	mock.Mock
}

type Params_Expecter struct {
	mock *mock.Mock
}

func (_m *Params) EXPECT() *Params_Expecter {
	return &Params_Expecter{mock: &_m.Mock}
}

// isParams provides a mock function with given fields:
func (_m *Params) isParams() {
	_m.Called()
}

// Params_isParams_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'isParams'
type Params_isParams_Call struct {
	*mock.Call
}

// isParams is a helper method to define mock.On call
func (_e *Params_Expecter) isParams() *Params_isParams_Call {
	return &Params_isParams_Call{Call: _e.mock.On("isParams")}
}

func (_c *Params_isParams_Call) Run(run func()) *Params_isParams_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Params_isParams_Call) Return() *Params_isParams_Call {
	_c.Call.Return()
	return _c
}

func (_c *Params_isParams_Call) RunAndReturn(run func()) *Params_isParams_Call {
	_c.Call.Return(run)
	return _c
}

// NewParams creates a new instance of Params. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewParams(t interface {
	mock.TestingT
	Cleanup(func())
}) *Params {
	mock := &Params{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
