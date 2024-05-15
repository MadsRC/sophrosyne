// Code generated by mockery v2.39.1. DO NOT EDIT.

package sophrosyne

import mock "github.com/stretchr/testify/mock"

// MockHttpService is an autogenerated mock type for the HttpService type
type MockHttpService struct {
	mock.Mock
}

type MockHttpService_Expecter struct {
	mock *mock.Mock
}

func (_m *MockHttpService) EXPECT() *MockHttpService_Expecter {
	return &MockHttpService_Expecter{mock: &_m.Mock}
}

// Start provides a mock function with given fields:
func (_m *MockHttpService) Start() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Start")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockHttpService_Start_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Start'
type MockHttpService_Start_Call struct {
	*mock.Call
}

// Start is a helper method to define mock.On call
func (_e *MockHttpService_Expecter) Start() *MockHttpService_Start_Call {
	return &MockHttpService_Start_Call{Call: _e.mock.On("Start")}
}

func (_c *MockHttpService_Start_Call) Run(run func()) *MockHttpService_Start_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockHttpService_Start_Call) Return(_a0 error) *MockHttpService_Start_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockHttpService_Start_Call) RunAndReturn(run func() error) *MockHttpService_Start_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockHttpService creates a new instance of MockHttpService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockHttpService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockHttpService {
	mock := &MockHttpService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
