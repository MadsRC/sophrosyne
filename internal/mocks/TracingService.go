// Code generated by mockery v2.39.1. DO NOT EDIT.

package sophrosyne

import (
	context "context"
	http "net/http"

	mock "github.com/stretchr/testify/mock"
)

// TracingService is an autogenerated mock type for the TracingService type
type TracingService struct {
	mock.Mock
}

type TracingService_Expecter struct {
	mock *mock.Mock
}

func (_m *TracingService) EXPECT() *TracingService_Expecter {
	return &TracingService_Expecter{mock: &_m.Mock}
}

// GetTraceID provides a mock function with given fields: ctx
func (_m *TracingService) GetTraceID(ctx context.Context) string {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetTraceID")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context) string); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// TracingService_GetTraceID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTraceID'
type TracingService_GetTraceID_Call struct {
	*mock.Call
}

// GetTraceID is a helper method to define mock.On call
//   - ctx context.Context
func (_e *TracingService_Expecter) GetTraceID(ctx interface{}) *TracingService_GetTraceID_Call {
	return &TracingService_GetTraceID_Call{Call: _e.mock.On("GetTraceID", ctx)}
}

func (_c *TracingService_GetTraceID_Call) Run(run func(ctx context.Context)) *TracingService_GetTraceID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *TracingService_GetTraceID_Call) Return(_a0 string) *TracingService_GetTraceID_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *TracingService_GetTraceID_Call) RunAndReturn(run func(context.Context) string) *TracingService_GetTraceID_Call {
	_c.Call.Return(run)
	return _c
}

// NewHTTPHandler provides a mock function with given fields: route, h
func (_m *TracingService) NewHTTPHandler(route string, h http.Handler) http.Handler {
	ret := _m.Called(route, h)

	if len(ret) == 0 {
		panic("no return value specified for NewHTTPHandler")
	}

	var r0 http.Handler
	if rf, ok := ret.Get(0).(func(string, http.Handler) http.Handler); ok {
		r0 = rf(route, h)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Handler)
		}
	}

	return r0
}

// TracingService_NewHTTPHandler_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'NewHTTPHandler'
type TracingService_NewHTTPHandler_Call struct {
	*mock.Call
}

// NewHTTPHandler is a helper method to define mock.On call
//   - route string
//   - h http.Handler
func (_e *TracingService_Expecter) NewHTTPHandler(route interface{}, h interface{}) *TracingService_NewHTTPHandler_Call {
	return &TracingService_NewHTTPHandler_Call{Call: _e.mock.On("NewHTTPHandler", route, h)}
}

func (_c *TracingService_NewHTTPHandler_Call) Run(run func(route string, h http.Handler)) *TracingService_NewHTTPHandler_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(http.Handler))
	})
	return _c
}

func (_c *TracingService_NewHTTPHandler_Call) Return(_a0 http.Handler) *TracingService_NewHTTPHandler_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *TracingService_NewHTTPHandler_Call) RunAndReturn(run func(string, http.Handler) http.Handler) *TracingService_NewHTTPHandler_Call {
	_c.Call.Return(run)
	return _c
}

// StartSpan provides a mock function with given fields: ctx, name
func (_m *TracingService) StartSpan(ctx context.Context, name string) (context.Context, Span) {
	ret := _m.Called(ctx, name)

	if len(ret) == 0 {
		panic("no return value specified for StartSpan")
	}

	var r0 context.Context
	var r1 Span
	if rf, ok := ret.Get(0).(func(context.Context, string) (context.Context, Span)); ok {
		return rf(ctx, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) context.Context); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(context.Context)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) Span); ok {
		r1 = rf(ctx, name)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(Span)
		}
	}

	return r0, r1
}

// TracingService_StartSpan_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'StartSpan'
type TracingService_StartSpan_Call struct {
	*mock.Call
}

// StartSpan is a helper method to define mock.On call
//   - ctx context.Context
//   - name string
func (_e *TracingService_Expecter) StartSpan(ctx interface{}, name interface{}) *TracingService_StartSpan_Call {
	return &TracingService_StartSpan_Call{Call: _e.mock.On("StartSpan", ctx, name)}
}

func (_c *TracingService_StartSpan_Call) Run(run func(ctx context.Context, name string)) *TracingService_StartSpan_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *TracingService_StartSpan_Call) Return(_a0 context.Context, _a1 Span) *TracingService_StartSpan_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *TracingService_StartSpan_Call) RunAndReturn(run func(context.Context, string) (context.Context, Span)) *TracingService_StartSpan_Call {
	_c.Call.Return(run)
	return _c
}

// WithRouteTag provides a mock function with given fields: route, h
func (_m *TracingService) WithRouteTag(route string, h http.Handler) http.Handler {
	ret := _m.Called(route, h)

	if len(ret) == 0 {
		panic("no return value specified for WithRouteTag")
	}

	var r0 http.Handler
	if rf, ok := ret.Get(0).(func(string, http.Handler) http.Handler); ok {
		r0 = rf(route, h)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Handler)
		}
	}

	return r0
}

// TracingService_WithRouteTag_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithRouteTag'
type TracingService_WithRouteTag_Call struct {
	*mock.Call
}

// WithRouteTag is a helper method to define mock.On call
//   - route string
//   - h http.Handler
func (_e *TracingService_Expecter) WithRouteTag(route interface{}, h interface{}) *TracingService_WithRouteTag_Call {
	return &TracingService_WithRouteTag_Call{Call: _e.mock.On("WithRouteTag", route, h)}
}

func (_c *TracingService_WithRouteTag_Call) Run(run func(route string, h http.Handler)) *TracingService_WithRouteTag_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(http.Handler))
	})
	return _c
}

func (_c *TracingService_WithRouteTag_Call) Return(_a0 http.Handler) *TracingService_WithRouteTag_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *TracingService_WithRouteTag_Call) RunAndReturn(run func(string, http.Handler) http.Handler) *TracingService_WithRouteTag_Call {
	_c.Call.Return(run)
	return _c
}

// NewTracingService creates a new instance of TracingService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewTracingService(t interface {
	mock.TestingT
	Cleanup(func())
}) *TracingService {
	mock := &TracingService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
