// Code generated by mockery v2.39.1. DO NOT EDIT.

package sophrosyne

import mock "github.com/stretchr/testify/mock"

// AuthorizationEntity is an autogenerated mock type for the AuthorizationEntity type
type AuthorizationEntity struct {
	mock.Mock
}

type AuthorizationEntity_Expecter struct {
	mock *mock.Mock
}

func (_m *AuthorizationEntity) EXPECT() *AuthorizationEntity_Expecter {
	return &AuthorizationEntity_Expecter{mock: &_m.Mock}
}

// EntityID provides a mock function with given fields:
func (_m *AuthorizationEntity) EntityID() string {
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

// AuthorizationEntity_EntityID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'EntityID'
type AuthorizationEntity_EntityID_Call struct {
	*mock.Call
}

// EntityID is a helper method to define mock.On call
func (_e *AuthorizationEntity_Expecter) EntityID() *AuthorizationEntity_EntityID_Call {
	return &AuthorizationEntity_EntityID_Call{Call: _e.mock.On("EntityID")}
}

func (_c *AuthorizationEntity_EntityID_Call) Run(run func()) *AuthorizationEntity_EntityID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AuthorizationEntity_EntityID_Call) Return(_a0 string) *AuthorizationEntity_EntityID_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AuthorizationEntity_EntityID_Call) RunAndReturn(run func() string) *AuthorizationEntity_EntityID_Call {
	_c.Call.Return(run)
	return _c
}

// EntityType provides a mock function with given fields:
func (_m *AuthorizationEntity) EntityType() string {
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

// AuthorizationEntity_EntityType_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'EntityType'
type AuthorizationEntity_EntityType_Call struct {
	*mock.Call
}

// EntityType is a helper method to define mock.On call
func (_e *AuthorizationEntity_Expecter) EntityType() *AuthorizationEntity_EntityType_Call {
	return &AuthorizationEntity_EntityType_Call{Call: _e.mock.On("EntityType")}
}

func (_c *AuthorizationEntity_EntityType_Call) Run(run func()) *AuthorizationEntity_EntityType_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AuthorizationEntity_EntityType_Call) Return(_a0 string) *AuthorizationEntity_EntityType_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AuthorizationEntity_EntityType_Call) RunAndReturn(run func() string) *AuthorizationEntity_EntityType_Call {
	_c.Call.Return(run)
	return _c
}

// NewAuthorizationEntity creates a new instance of AuthorizationEntity. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAuthorizationEntity(t interface {
	mock.TestingT
	Cleanup(func())
}) *AuthorizationEntity {
	mock := &AuthorizationEntity{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
