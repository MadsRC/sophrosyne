// Code generated by mockery v2.39.1. DO NOT EDIT.

package sophrosyne

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MetricService is an autogenerated mock type for the MetricService type
type MetricService struct {
	mock.Mock
}

type MetricService_Expecter struct {
	mock *mock.Mock
}

func (_m *MetricService) EXPECT() *MetricService_Expecter {
	return &MetricService_Expecter{mock: &_m.Mock}
}

// RecordPanic provides a mock function with given fields: ctx
func (_m *MetricService) RecordPanic(ctx context.Context) {
	_m.Called(ctx)
}

// MetricService_RecordPanic_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RecordPanic'
type MetricService_RecordPanic_Call struct {
	*mock.Call
}

// RecordPanic is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MetricService_Expecter) RecordPanic(ctx interface{}) *MetricService_RecordPanic_Call {
	return &MetricService_RecordPanic_Call{Call: _e.mock.On("RecordPanic", ctx)}
}

func (_c *MetricService_RecordPanic_Call) Run(run func(ctx context.Context)) *MetricService_RecordPanic_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MetricService_RecordPanic_Call) Return() *MetricService_RecordPanic_Call {
	_c.Call.Return()
	return _c
}

func (_c *MetricService_RecordPanic_Call) RunAndReturn(run func(context.Context)) *MetricService_RecordPanic_Call {
	_c.Call.Return(run)
	return _c
}

// NewMetricService creates a new instance of MetricService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMetricService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MetricService {
	mock := &MetricService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
