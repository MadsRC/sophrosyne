// Sophrosyne
//   Copyright (C) 2024  Mads R. Havmand
//
// This program is free software: you can redistribute it and/or modify
//   it under the terms of the GNU Affero General Public License as published by
//   the Free Software Foundation, either version 3 of the License, or
//   (at your option) any later version.
//
//   This program is distributed in the hope that it will be useful,
//   but WITHOUT ANY WARRANTY; without even the implied warranty of
//   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//   GNU Affero General Public License for more details.
//
//   You should have received a copy of the GNU Affero General Public License
//   along with this program.  If not, see <http://www.gnu.org/licenses/>.

package grpc

import (
	"context"
	"github.com/madsrc/sophrosyne"
	v0 "github.com/madsrc/sophrosyne/internal/grpc/sophrosyne/v0"
	sophrosyne2 "github.com/madsrc/sophrosyne/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

// GetCheck runs successfully with name.
func Test_CheckServiceServer_GetCheck_withName(t *testing.T) {
	mockCheckService := sophrosyne2.NewMockCheckService(t)
	mockAuthzProvider := sophrosyne2.NewMockAuthorizationProvider(t)
	server := CheckServiceServer{
		checkService:  mockCheckService,
		authzProvider: mockAuthzProvider,
	}

	user := sophrosyne.User{
		ID: "test",
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, sophrosyne.UserContextKey{}, &user)

	request := &v0.GetCheckRequest{
		Check: &v0.GetCheckRequest_Name{
			Name: "test",
		},
	}

	mockCheckService.On("GetCheckByName", ctx, request.GetName()).Return(sophrosyne.Check{
		ID:   "test",
		Name: "test",
	}, nil)
	mockAuthzProvider.On("IsAuthorized", ctx, sophrosyne.AuthorizationRequest{
		Principal: &user,
		Action:    sophrosyne.AuthorizationAction("GetProfile"),
		Resource:  sophrosyne.Profile{ID: "test"},
	}).Return(true)

	response, err := server.GetCheck(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.True(t, mockCheckService.AssertExpectations(t))
	require.True(t, mockAuthzProvider.AssertExpectations(t))
}

// GetCheck runs successfully with ID.
func Test_CheckServiceServer_GetCheck_withID(t *testing.T) {
	mockCheckService := sophrosyne2.NewMockCheckService(t)
	mockAuthzProvider := sophrosyne2.NewMockAuthorizationProvider(t)
	server := CheckServiceServer{
		checkService:  mockCheckService,
		authzProvider: mockAuthzProvider,
	}

	user := sophrosyne.User{
		ID: "test",
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, sophrosyne.UserContextKey{}, &user)

	request := &v0.GetCheckRequest{
		Check: &v0.GetCheckRequest_Id{
			Id: "someID",
		},
	}

	mockCheckService.On("GetCheck", ctx, request.GetId()).Return(sophrosyne.Check{
		ID:   "test",
		Name: "test",
	}, nil)
	mockAuthzProvider.On("IsAuthorized", ctx, sophrosyne.AuthorizationRequest{
		Principal: &user,
		Action:    sophrosyne.AuthorizationAction("GetProfile"),
		Resource:  sophrosyne.Profile{ID: "test"},
	}).Return(true)

	response, err := server.GetCheck(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.True(t, mockCheckService.AssertExpectations(t))
	require.True(t, mockAuthzProvider.AssertExpectations(t))
}

// GetCheck returns an error if the user is not authorized.
func Test_CheckServiceServer_GetCheck_UserNotAuthorized(t *testing.T) {
	mockCheckService := sophrosyne2.NewMockCheckService(t)
	mockAuthzProvider := sophrosyne2.NewMockAuthorizationProvider(t)
	server := CheckServiceServer{
		checkService:  mockCheckService,
		authzProvider: mockAuthzProvider,
	}

	user := sophrosyne.User{
		ID: "test",
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, sophrosyne.UserContextKey{}, &user)

	request := &v0.GetCheckRequest{
		Check: &v0.GetCheckRequest_Name{
			Name: "test",
		},
	}

	mockCheckService.On("GetCheckByName", ctx, request.GetName()).Return(sophrosyne.Check{
		ID:   "test",
		Name: "test",
	}, nil)
	mockAuthzProvider.On("IsAuthorized", ctx, sophrosyne.AuthorizationRequest{
		Principal: &user,
		Action:    sophrosyne.AuthorizationAction("GetProfile"),
		Resource:  sophrosyne.Profile{ID: "test"},
	}).Return(false)

	response, err := server.GetCheck(ctx, request)

	require.Error(t, err)
	grpcStatus, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.PermissionDenied, grpcStatus.Code())
	require.Equal(t, "unauthorized", grpcStatus.Message())
	require.Nil(t, response)
	require.True(t, mockCheckService.AssertExpectations(t))
	require.True(t, mockAuthzProvider.AssertExpectations(t))
}

// GetCheck returns an error if the check is not found.
func Test_CheckServiceServer_GetCheck_CheckNotFound(t *testing.T) {
	mockCheckService := sophrosyne2.NewMockCheckService(t)
	server := CheckServiceServer{
		checkService: mockCheckService,
	}

	user := sophrosyne.User{
		ID: "test",
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, sophrosyne.UserContextKey{}, &user)

	request := &v0.GetCheckRequest{
		Check: &v0.GetCheckRequest_Name{
			Name: "test",
		},
	}

	mockCheckService.On("GetCheckByName", ctx, request.GetName()).Return(sophrosyne.Check{}, assert.AnError)

	response, err := server.GetCheck(ctx, request)

	require.Error(t, err)
	grpcStatus, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Internal, grpcStatus.Code())
	require.Contains(t, grpcStatus.Message(), "error getting check:")
	require.Nil(t, response)
	require.True(t, mockCheckService.AssertExpectations(t))
}

// GetCheck returns an authentication error if the user is not attached to the context.
func Test_CheckServiceServer_GetCheck_NoUserInContext(t *testing.T) {
	server := CheckServiceServer{}

	response, err := server.GetCheck(context.Background(), &v0.GetCheckRequest{})

	require.Error(t, err)
	grpcStatus, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Unauthenticated, grpcStatus.Code())
	require.Equal(t, InvalidTokenMsg, grpcStatus.Message())
	require.Nil(t, response)
}
