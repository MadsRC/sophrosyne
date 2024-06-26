// Code generated by mockery v2.43.1. DO NOT EDIT.

package sophrosyne

import mock "github.com/stretchr/testify/mock"

// MockAuthorizationEntity is an autogenerated mock type for the AuthorizationEntity type
type MockAuthorizationEntity struct {
	mock.Mock
}

type MockAuthorizationEntity_Expecter struct {
	mock *mock.Mock
}

func (_m *MockAuthorizationEntity) EXPECT() *MockAuthorizationEntity_Expecter {
	return &MockAuthorizationEntity_Expecter{mock: &_m.Mock}
}

// EntityID provides a mock function with given fields:
func (_m *MockAuthorizationEntity) EntityID() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for EntityID")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockAuthorizationEntity_EntityID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'EntityID'
type MockAuthorizationEntity_EntityID_Call struct {
	*mock.Call
}

// EntityID is a helper method to define mock.On call
func (_e *MockAuthorizationEntity_Expecter) EntityID() *MockAuthorizationEntity_EntityID_Call {
	return &MockAuthorizationEntity_EntityID_Call{Call: _e.mock.On("EntityID")}
}

func (_c *MockAuthorizationEntity_EntityID_Call) Run(run func()) *MockAuthorizationEntity_EntityID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockAuthorizationEntity_EntityID_Call) Return(_a0 string) *MockAuthorizationEntity_EntityID_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockAuthorizationEntity_EntityID_Call) RunAndReturn(run func() string) *MockAuthorizationEntity_EntityID_Call {
	_c.Call.Return(run)
	return _c
}

// EntityType provides a mock function with given fields:
func (_m *MockAuthorizationEntity) EntityType() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for EntityType")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockAuthorizationEntity_EntityType_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'EntityType'
type MockAuthorizationEntity_EntityType_Call struct {
	*mock.Call
}

// EntityType is a helper method to define mock.On call
func (_e *MockAuthorizationEntity_Expecter) EntityType() *MockAuthorizationEntity_EntityType_Call {
	return &MockAuthorizationEntity_EntityType_Call{Call: _e.mock.On("EntityType")}
}

func (_c *MockAuthorizationEntity_EntityType_Call) Run(run func()) *MockAuthorizationEntity_EntityType_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockAuthorizationEntity_EntityType_Call) Return(_a0 string) *MockAuthorizationEntity_EntityType_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockAuthorizationEntity_EntityType_Call) RunAndReturn(run func() string) *MockAuthorizationEntity_EntityType_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockAuthorizationEntity creates a new instance of MockAuthorizationEntity. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockAuthorizationEntity(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockAuthorizationEntity {
	mock := &MockAuthorizationEntity{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
