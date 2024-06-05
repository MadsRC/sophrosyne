// Code generated by mockery v2.43.2. DO NOT EDIT.

package v0

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	emptypb "google.golang.org/protobuf/types/known/emptypb"

	v0 "github.com/madsrc/sophrosyne/internal/grpc/sophrosyne/v0"
)

// MockUserServiceServer is an autogenerated mock type for the UserServiceServer type
type MockUserServiceServer struct {
	mock.Mock
}

type MockUserServiceServer_Expecter struct {
	mock *mock.Mock
}

func (_m *MockUserServiceServer) EXPECT() *MockUserServiceServer_Expecter {
	return &MockUserServiceServer_Expecter{mock: &_m.Mock}
}

// CreateUser provides a mock function with given fields: _a0, _a1
func (_m *MockUserServiceServer) CreateUser(_a0 context.Context, _a1 *v0.CreateUserRequest) (*v0.CreateUserResponse, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for CreateUser")
	}

	var r0 *v0.CreateUserResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v0.CreateUserRequest) (*v0.CreateUserResponse, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v0.CreateUserRequest) *v0.CreateUserResponse); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v0.CreateUserResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v0.CreateUserRequest) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockUserServiceServer_CreateUser_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateUser'
type MockUserServiceServer_CreateUser_Call struct {
	*mock.Call
}

// CreateUser is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *v0.CreateUserRequest
func (_e *MockUserServiceServer_Expecter) CreateUser(_a0 interface{}, _a1 interface{}) *MockUserServiceServer_CreateUser_Call {
	return &MockUserServiceServer_CreateUser_Call{Call: _e.mock.On("CreateUser", _a0, _a1)}
}

func (_c *MockUserServiceServer_CreateUser_Call) Run(run func(_a0 context.Context, _a1 *v0.CreateUserRequest)) *MockUserServiceServer_CreateUser_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*v0.CreateUserRequest))
	})
	return _c
}

func (_c *MockUserServiceServer_CreateUser_Call) Return(_a0 *v0.CreateUserResponse, _a1 error) *MockUserServiceServer_CreateUser_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockUserServiceServer_CreateUser_Call) RunAndReturn(run func(context.Context, *v0.CreateUserRequest) (*v0.CreateUserResponse, error)) *MockUserServiceServer_CreateUser_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteUser provides a mock function with given fields: _a0, _a1
func (_m *MockUserServiceServer) DeleteUser(_a0 context.Context, _a1 *v0.DeleteUserRequest) (*emptypb.Empty, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for DeleteUser")
	}

	var r0 *emptypb.Empty
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v0.DeleteUserRequest) (*emptypb.Empty, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v0.DeleteUserRequest) *emptypb.Empty); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*emptypb.Empty)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v0.DeleteUserRequest) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockUserServiceServer_DeleteUser_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteUser'
type MockUserServiceServer_DeleteUser_Call struct {
	*mock.Call
}

// DeleteUser is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *v0.DeleteUserRequest
func (_e *MockUserServiceServer_Expecter) DeleteUser(_a0 interface{}, _a1 interface{}) *MockUserServiceServer_DeleteUser_Call {
	return &MockUserServiceServer_DeleteUser_Call{Call: _e.mock.On("DeleteUser", _a0, _a1)}
}

func (_c *MockUserServiceServer_DeleteUser_Call) Run(run func(_a0 context.Context, _a1 *v0.DeleteUserRequest)) *MockUserServiceServer_DeleteUser_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*v0.DeleteUserRequest))
	})
	return _c
}

func (_c *MockUserServiceServer_DeleteUser_Call) Return(_a0 *emptypb.Empty, _a1 error) *MockUserServiceServer_DeleteUser_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockUserServiceServer_DeleteUser_Call) RunAndReturn(run func(context.Context, *v0.DeleteUserRequest) (*emptypb.Empty, error)) *MockUserServiceServer_DeleteUser_Call {
	_c.Call.Return(run)
	return _c
}

// GetUser provides a mock function with given fields: _a0, _a1
func (_m *MockUserServiceServer) GetUser(_a0 context.Context, _a1 *v0.GetUserRequest) (*v0.GetUserResponse, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for GetUser")
	}

	var r0 *v0.GetUserResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v0.GetUserRequest) (*v0.GetUserResponse, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v0.GetUserRequest) *v0.GetUserResponse); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v0.GetUserResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v0.GetUserRequest) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockUserServiceServer_GetUser_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetUser'
type MockUserServiceServer_GetUser_Call struct {
	*mock.Call
}

// GetUser is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *v0.GetUserRequest
func (_e *MockUserServiceServer_Expecter) GetUser(_a0 interface{}, _a1 interface{}) *MockUserServiceServer_GetUser_Call {
	return &MockUserServiceServer_GetUser_Call{Call: _e.mock.On("GetUser", _a0, _a1)}
}

func (_c *MockUserServiceServer_GetUser_Call) Run(run func(_a0 context.Context, _a1 *v0.GetUserRequest)) *MockUserServiceServer_GetUser_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*v0.GetUserRequest))
	})
	return _c
}

func (_c *MockUserServiceServer_GetUser_Call) Return(_a0 *v0.GetUserResponse, _a1 error) *MockUserServiceServer_GetUser_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockUserServiceServer_GetUser_Call) RunAndReturn(run func(context.Context, *v0.GetUserRequest) (*v0.GetUserResponse, error)) *MockUserServiceServer_GetUser_Call {
	_c.Call.Return(run)
	return _c
}

// GetUsers provides a mock function with given fields: _a0, _a1
func (_m *MockUserServiceServer) GetUsers(_a0 context.Context, _a1 *v0.GetUsersRequest) (*v0.GetUsersResponse, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for GetUsers")
	}

	var r0 *v0.GetUsersResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v0.GetUsersRequest) (*v0.GetUsersResponse, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v0.GetUsersRequest) *v0.GetUsersResponse); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v0.GetUsersResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v0.GetUsersRequest) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockUserServiceServer_GetUsers_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetUsers'
type MockUserServiceServer_GetUsers_Call struct {
	*mock.Call
}

// GetUsers is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *v0.GetUsersRequest
func (_e *MockUserServiceServer_Expecter) GetUsers(_a0 interface{}, _a1 interface{}) *MockUserServiceServer_GetUsers_Call {
	return &MockUserServiceServer_GetUsers_Call{Call: _e.mock.On("GetUsers", _a0, _a1)}
}

func (_c *MockUserServiceServer_GetUsers_Call) Run(run func(_a0 context.Context, _a1 *v0.GetUsersRequest)) *MockUserServiceServer_GetUsers_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*v0.GetUsersRequest))
	})
	return _c
}

func (_c *MockUserServiceServer_GetUsers_Call) Return(_a0 *v0.GetUsersResponse, _a1 error) *MockUserServiceServer_GetUsers_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockUserServiceServer_GetUsers_Call) RunAndReturn(run func(context.Context, *v0.GetUsersRequest) (*v0.GetUsersResponse, error)) *MockUserServiceServer_GetUsers_Call {
	_c.Call.Return(run)
	return _c
}

// RotateToken provides a mock function with given fields: _a0, _a1
func (_m *MockUserServiceServer) RotateToken(_a0 context.Context, _a1 *v0.RotateTokenRequest) (*v0.RotateTokenResponse, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for RotateToken")
	}

	var r0 *v0.RotateTokenResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v0.RotateTokenRequest) (*v0.RotateTokenResponse, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v0.RotateTokenRequest) *v0.RotateTokenResponse); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v0.RotateTokenResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v0.RotateTokenRequest) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockUserServiceServer_RotateToken_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RotateToken'
type MockUserServiceServer_RotateToken_Call struct {
	*mock.Call
}

// RotateToken is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *v0.RotateTokenRequest
func (_e *MockUserServiceServer_Expecter) RotateToken(_a0 interface{}, _a1 interface{}) *MockUserServiceServer_RotateToken_Call {
	return &MockUserServiceServer_RotateToken_Call{Call: _e.mock.On("RotateToken", _a0, _a1)}
}

func (_c *MockUserServiceServer_RotateToken_Call) Run(run func(_a0 context.Context, _a1 *v0.RotateTokenRequest)) *MockUserServiceServer_RotateToken_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*v0.RotateTokenRequest))
	})
	return _c
}

func (_c *MockUserServiceServer_RotateToken_Call) Return(_a0 *v0.RotateTokenResponse, _a1 error) *MockUserServiceServer_RotateToken_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockUserServiceServer_RotateToken_Call) RunAndReturn(run func(context.Context, *v0.RotateTokenRequest) (*v0.RotateTokenResponse, error)) *MockUserServiceServer_RotateToken_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateUser provides a mock function with given fields: _a0, _a1
func (_m *MockUserServiceServer) UpdateUser(_a0 context.Context, _a1 *v0.UpdateUserRequest) (*v0.UpdateUserResponse, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for UpdateUser")
	}

	var r0 *v0.UpdateUserResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v0.UpdateUserRequest) (*v0.UpdateUserResponse, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v0.UpdateUserRequest) *v0.UpdateUserResponse); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v0.UpdateUserResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v0.UpdateUserRequest) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockUserServiceServer_UpdateUser_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateUser'
type MockUserServiceServer_UpdateUser_Call struct {
	*mock.Call
}

// UpdateUser is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *v0.UpdateUserRequest
func (_e *MockUserServiceServer_Expecter) UpdateUser(_a0 interface{}, _a1 interface{}) *MockUserServiceServer_UpdateUser_Call {
	return &MockUserServiceServer_UpdateUser_Call{Call: _e.mock.On("UpdateUser", _a0, _a1)}
}

func (_c *MockUserServiceServer_UpdateUser_Call) Run(run func(_a0 context.Context, _a1 *v0.UpdateUserRequest)) *MockUserServiceServer_UpdateUser_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*v0.UpdateUserRequest))
	})
	return _c
}

func (_c *MockUserServiceServer_UpdateUser_Call) Return(_a0 *v0.UpdateUserResponse, _a1 error) *MockUserServiceServer_UpdateUser_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockUserServiceServer_UpdateUser_Call) RunAndReturn(run func(context.Context, *v0.UpdateUserRequest) (*v0.UpdateUserResponse, error)) *MockUserServiceServer_UpdateUser_Call {
	_c.Call.Return(run)
	return _c
}

// mustEmbedUnimplementedUserServiceServer provides a mock function with given fields:
func (_m *MockUserServiceServer) mustEmbedUnimplementedUserServiceServer() {
	_m.Called()
}

// MockUserServiceServer_mustEmbedUnimplementedUserServiceServer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'mustEmbedUnimplementedUserServiceServer'
type MockUserServiceServer_mustEmbedUnimplementedUserServiceServer_Call struct {
	*mock.Call
}

// mustEmbedUnimplementedUserServiceServer is a helper method to define mock.On call
func (_e *MockUserServiceServer_Expecter) mustEmbedUnimplementedUserServiceServer() *MockUserServiceServer_mustEmbedUnimplementedUserServiceServer_Call {
	return &MockUserServiceServer_mustEmbedUnimplementedUserServiceServer_Call{Call: _e.mock.On("mustEmbedUnimplementedUserServiceServer")}
}

func (_c *MockUserServiceServer_mustEmbedUnimplementedUserServiceServer_Call) Run(run func()) *MockUserServiceServer_mustEmbedUnimplementedUserServiceServer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockUserServiceServer_mustEmbedUnimplementedUserServiceServer_Call) Return() *MockUserServiceServer_mustEmbedUnimplementedUserServiceServer_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockUserServiceServer_mustEmbedUnimplementedUserServiceServer_Call) RunAndReturn(run func()) *MockUserServiceServer_mustEmbedUnimplementedUserServiceServer_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockUserServiceServer creates a new instance of MockUserServiceServer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockUserServiceServer(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockUserServiceServer {
	mock := &MockUserServiceServer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
