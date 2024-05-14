// Code generated by mockery v2.39.1. DO NOT EDIT.

package sophrosyne

import mock "github.com/stretchr/testify/mock"

// Span is an autogenerated mock type for the Span type
type Span struct {
	mock.Mock
}

type Span_Expecter struct {
	mock *mock.Mock
}

func (_m *Span) EXPECT() *Span_Expecter {
	return &Span_Expecter{mock: &_m.Mock}
}

// End provides a mock function with given fields:
func (_m *Span) End() {
	_m.Called()
}

// Span_End_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'End'
type Span_End_Call struct {
	*mock.Call
}

// End is a helper method to define mock.On call
func (_e *Span_Expecter) End() *Span_End_Call {
	return &Span_End_Call{Call: _e.mock.On("End")}
}

func (_c *Span_End_Call) Run(run func()) *Span_End_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Span_End_Call) Return() *Span_End_Call {
	_c.Call.Return()
	return _c
}

func (_c *Span_End_Call) RunAndReturn(run func()) *Span_End_Call {
	_c.Call.Return(run)
	return _c
}

// NewSpan creates a new instance of Span. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSpan(t interface {
	mock.TestingT
	Cleanup(func())
}) *Span {
	mock := &Span{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
