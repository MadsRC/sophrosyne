// Code generated by mockery v2.43.2. DO NOT EDIT.

package v0

import mock "github.com/stretchr/testify/mock"

// MockUnsafeCheckProviderServiceServer is an autogenerated mock type for the UnsafeCheckProviderServiceServer type
type MockUnsafeCheckProviderServiceServer struct {
	mock.Mock
}

type MockUnsafeCheckProviderServiceServer_Expecter struct {
	mock *mock.Mock
}

func (_m *MockUnsafeCheckProviderServiceServer) EXPECT() *MockUnsafeCheckProviderServiceServer_Expecter {
	return &MockUnsafeCheckProviderServiceServer_Expecter{mock: &_m.Mock}
}

// mustEmbedUnimplementedCheckProviderServiceServer provides a mock function with given fields:
func (_m *MockUnsafeCheckProviderServiceServer) mustEmbedUnimplementedCheckProviderServiceServer() {
	_m.Called()
}

// MockUnsafeCheckProviderServiceServer_mustEmbedUnimplementedCheckProviderServiceServer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'mustEmbedUnimplementedCheckProviderServiceServer'
type MockUnsafeCheckProviderServiceServer_mustEmbedUnimplementedCheckProviderServiceServer_Call struct {
	*mock.Call
}

// mustEmbedUnimplementedCheckProviderServiceServer is a helper method to define mock.On call
func (_e *MockUnsafeCheckProviderServiceServer_Expecter) mustEmbedUnimplementedCheckProviderServiceServer() *MockUnsafeCheckProviderServiceServer_mustEmbedUnimplementedCheckProviderServiceServer_Call {
	return &MockUnsafeCheckProviderServiceServer_mustEmbedUnimplementedCheckProviderServiceServer_Call{Call: _e.mock.On("mustEmbedUnimplementedCheckProviderServiceServer")}
}

func (_c *MockUnsafeCheckProviderServiceServer_mustEmbedUnimplementedCheckProviderServiceServer_Call) Run(run func()) *MockUnsafeCheckProviderServiceServer_mustEmbedUnimplementedCheckProviderServiceServer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockUnsafeCheckProviderServiceServer_mustEmbedUnimplementedCheckProviderServiceServer_Call) Return() *MockUnsafeCheckProviderServiceServer_mustEmbedUnimplementedCheckProviderServiceServer_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockUnsafeCheckProviderServiceServer_mustEmbedUnimplementedCheckProviderServiceServer_Call) RunAndReturn(run func()) *MockUnsafeCheckProviderServiceServer_mustEmbedUnimplementedCheckProviderServiceServer_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockUnsafeCheckProviderServiceServer creates a new instance of MockUnsafeCheckProviderServiceServer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockUnsafeCheckProviderServiceServer(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockUnsafeCheckProviderServiceServer {
	mock := &MockUnsafeCheckProviderServiceServer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
