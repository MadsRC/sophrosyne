// Code generated by mockery v2.43.2. DO NOT EDIT.

package v0

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	emptypb "google.golang.org/protobuf/types/known/emptypb"

	v0 "github.com/madsrc/sophrosyne/internal/grpc/sophrosyne/v0"
)

// MockProfileServiceServer is an autogenerated mock type for the ProfileServiceServer type
type MockProfileServiceServer struct {
	mock.Mock
}

type MockProfileServiceServer_Expecter struct {
	mock *mock.Mock
}

func (_m *MockProfileServiceServer) EXPECT() *MockProfileServiceServer_Expecter {
	return &MockProfileServiceServer_Expecter{mock: &_m.Mock}
}

// CreateProfile provides a mock function with given fields: _a0, _a1
func (_m *MockProfileServiceServer) CreateProfile(_a0 context.Context, _a1 *v0.CreateProfileRequest) (*v0.CreateProfileResponse, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for CreateProfile")
	}

	var r0 *v0.CreateProfileResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v0.CreateProfileRequest) (*v0.CreateProfileResponse, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v0.CreateProfileRequest) *v0.CreateProfileResponse); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v0.CreateProfileResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v0.CreateProfileRequest) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockProfileServiceServer_CreateProfile_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateProfile'
type MockProfileServiceServer_CreateProfile_Call struct {
	*mock.Call
}

// CreateProfile is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *v0.CreateProfileRequest
func (_e *MockProfileServiceServer_Expecter) CreateProfile(_a0 interface{}, _a1 interface{}) *MockProfileServiceServer_CreateProfile_Call {
	return &MockProfileServiceServer_CreateProfile_Call{Call: _e.mock.On("CreateProfile", _a0, _a1)}
}

func (_c *MockProfileServiceServer_CreateProfile_Call) Run(run func(_a0 context.Context, _a1 *v0.CreateProfileRequest)) *MockProfileServiceServer_CreateProfile_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*v0.CreateProfileRequest))
	})
	return _c
}

func (_c *MockProfileServiceServer_CreateProfile_Call) Return(_a0 *v0.CreateProfileResponse, _a1 error) *MockProfileServiceServer_CreateProfile_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockProfileServiceServer_CreateProfile_Call) RunAndReturn(run func(context.Context, *v0.CreateProfileRequest) (*v0.CreateProfileResponse, error)) *MockProfileServiceServer_CreateProfile_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteProfile provides a mock function with given fields: _a0, _a1
func (_m *MockProfileServiceServer) DeleteProfile(_a0 context.Context, _a1 *v0.DeleteProfileRequest) (*emptypb.Empty, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for DeleteProfile")
	}

	var r0 *emptypb.Empty
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v0.DeleteProfileRequest) (*emptypb.Empty, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v0.DeleteProfileRequest) *emptypb.Empty); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*emptypb.Empty)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v0.DeleteProfileRequest) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockProfileServiceServer_DeleteProfile_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteProfile'
type MockProfileServiceServer_DeleteProfile_Call struct {
	*mock.Call
}

// DeleteProfile is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *v0.DeleteProfileRequest
func (_e *MockProfileServiceServer_Expecter) DeleteProfile(_a0 interface{}, _a1 interface{}) *MockProfileServiceServer_DeleteProfile_Call {
	return &MockProfileServiceServer_DeleteProfile_Call{Call: _e.mock.On("DeleteProfile", _a0, _a1)}
}

func (_c *MockProfileServiceServer_DeleteProfile_Call) Run(run func(_a0 context.Context, _a1 *v0.DeleteProfileRequest)) *MockProfileServiceServer_DeleteProfile_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*v0.DeleteProfileRequest))
	})
	return _c
}

func (_c *MockProfileServiceServer_DeleteProfile_Call) Return(_a0 *emptypb.Empty, _a1 error) *MockProfileServiceServer_DeleteProfile_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockProfileServiceServer_DeleteProfile_Call) RunAndReturn(run func(context.Context, *v0.DeleteProfileRequest) (*emptypb.Empty, error)) *MockProfileServiceServer_DeleteProfile_Call {
	_c.Call.Return(run)
	return _c
}

// GetProfile provides a mock function with given fields: _a0, _a1
func (_m *MockProfileServiceServer) GetProfile(_a0 context.Context, _a1 *v0.GetProfileRequest) (*v0.GetProfileResponse, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for GetProfile")
	}

	var r0 *v0.GetProfileResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v0.GetProfileRequest) (*v0.GetProfileResponse, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v0.GetProfileRequest) *v0.GetProfileResponse); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v0.GetProfileResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v0.GetProfileRequest) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockProfileServiceServer_GetProfile_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetProfile'
type MockProfileServiceServer_GetProfile_Call struct {
	*mock.Call
}

// GetProfile is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *v0.GetProfileRequest
func (_e *MockProfileServiceServer_Expecter) GetProfile(_a0 interface{}, _a1 interface{}) *MockProfileServiceServer_GetProfile_Call {
	return &MockProfileServiceServer_GetProfile_Call{Call: _e.mock.On("GetProfile", _a0, _a1)}
}

func (_c *MockProfileServiceServer_GetProfile_Call) Run(run func(_a0 context.Context, _a1 *v0.GetProfileRequest)) *MockProfileServiceServer_GetProfile_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*v0.GetProfileRequest))
	})
	return _c
}

func (_c *MockProfileServiceServer_GetProfile_Call) Return(_a0 *v0.GetProfileResponse, _a1 error) *MockProfileServiceServer_GetProfile_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockProfileServiceServer_GetProfile_Call) RunAndReturn(run func(context.Context, *v0.GetProfileRequest) (*v0.GetProfileResponse, error)) *MockProfileServiceServer_GetProfile_Call {
	_c.Call.Return(run)
	return _c
}

// GetProfiles provides a mock function with given fields: _a0, _a1
func (_m *MockProfileServiceServer) GetProfiles(_a0 context.Context, _a1 *v0.GetProfilesRequest) (*v0.GetProfilesResponse, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for GetProfiles")
	}

	var r0 *v0.GetProfilesResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v0.GetProfilesRequest) (*v0.GetProfilesResponse, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v0.GetProfilesRequest) *v0.GetProfilesResponse); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v0.GetProfilesResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v0.GetProfilesRequest) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockProfileServiceServer_GetProfiles_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetProfiles'
type MockProfileServiceServer_GetProfiles_Call struct {
	*mock.Call
}

// GetProfiles is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *v0.GetProfilesRequest
func (_e *MockProfileServiceServer_Expecter) GetProfiles(_a0 interface{}, _a1 interface{}) *MockProfileServiceServer_GetProfiles_Call {
	return &MockProfileServiceServer_GetProfiles_Call{Call: _e.mock.On("GetProfiles", _a0, _a1)}
}

func (_c *MockProfileServiceServer_GetProfiles_Call) Run(run func(_a0 context.Context, _a1 *v0.GetProfilesRequest)) *MockProfileServiceServer_GetProfiles_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*v0.GetProfilesRequest))
	})
	return _c
}

func (_c *MockProfileServiceServer_GetProfiles_Call) Return(_a0 *v0.GetProfilesResponse, _a1 error) *MockProfileServiceServer_GetProfiles_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockProfileServiceServer_GetProfiles_Call) RunAndReturn(run func(context.Context, *v0.GetProfilesRequest) (*v0.GetProfilesResponse, error)) *MockProfileServiceServer_GetProfiles_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateProfile provides a mock function with given fields: _a0, _a1
func (_m *MockProfileServiceServer) UpdateProfile(_a0 context.Context, _a1 *v0.UpdateProfileRequest) (*v0.UpdateProfileResponse, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for UpdateProfile")
	}

	var r0 *v0.UpdateProfileResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v0.UpdateProfileRequest) (*v0.UpdateProfileResponse, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v0.UpdateProfileRequest) *v0.UpdateProfileResponse); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v0.UpdateProfileResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v0.UpdateProfileRequest) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockProfileServiceServer_UpdateProfile_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateProfile'
type MockProfileServiceServer_UpdateProfile_Call struct {
	*mock.Call
}

// UpdateProfile is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *v0.UpdateProfileRequest
func (_e *MockProfileServiceServer_Expecter) UpdateProfile(_a0 interface{}, _a1 interface{}) *MockProfileServiceServer_UpdateProfile_Call {
	return &MockProfileServiceServer_UpdateProfile_Call{Call: _e.mock.On("UpdateProfile", _a0, _a1)}
}

func (_c *MockProfileServiceServer_UpdateProfile_Call) Run(run func(_a0 context.Context, _a1 *v0.UpdateProfileRequest)) *MockProfileServiceServer_UpdateProfile_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*v0.UpdateProfileRequest))
	})
	return _c
}

func (_c *MockProfileServiceServer_UpdateProfile_Call) Return(_a0 *v0.UpdateProfileResponse, _a1 error) *MockProfileServiceServer_UpdateProfile_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockProfileServiceServer_UpdateProfile_Call) RunAndReturn(run func(context.Context, *v0.UpdateProfileRequest) (*v0.UpdateProfileResponse, error)) *MockProfileServiceServer_UpdateProfile_Call {
	_c.Call.Return(run)
	return _c
}

// mustEmbedUnimplementedProfileServiceServer provides a mock function with given fields:
func (_m *MockProfileServiceServer) mustEmbedUnimplementedProfileServiceServer() {
	_m.Called()
}

// MockProfileServiceServer_mustEmbedUnimplementedProfileServiceServer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'mustEmbedUnimplementedProfileServiceServer'
type MockProfileServiceServer_mustEmbedUnimplementedProfileServiceServer_Call struct {
	*mock.Call
}

// mustEmbedUnimplementedProfileServiceServer is a helper method to define mock.On call
func (_e *MockProfileServiceServer_Expecter) mustEmbedUnimplementedProfileServiceServer() *MockProfileServiceServer_mustEmbedUnimplementedProfileServiceServer_Call {
	return &MockProfileServiceServer_mustEmbedUnimplementedProfileServiceServer_Call{Call: _e.mock.On("mustEmbedUnimplementedProfileServiceServer")}
}

func (_c *MockProfileServiceServer_mustEmbedUnimplementedProfileServiceServer_Call) Run(run func()) *MockProfileServiceServer_mustEmbedUnimplementedProfileServiceServer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockProfileServiceServer_mustEmbedUnimplementedProfileServiceServer_Call) Return() *MockProfileServiceServer_mustEmbedUnimplementedProfileServiceServer_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockProfileServiceServer_mustEmbedUnimplementedProfileServiceServer_Call) RunAndReturn(run func()) *MockProfileServiceServer_mustEmbedUnimplementedProfileServiceServer_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockProfileServiceServer creates a new instance of MockProfileServiceServer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockProfileServiceServer(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockProfileServiceServer {
	mock := &MockProfileServiceServer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
