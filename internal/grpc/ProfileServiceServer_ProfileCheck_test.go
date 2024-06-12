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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/madsrc/sophrosyne"
	v0 "github.com/madsrc/sophrosyne/internal/grpc/sophrosyne/v0"
	sophrosyne2 "github.com/madsrc/sophrosyne/internal/mocks"
)

// GetProfile runs successfully with name.
func Test_ProfileServiceServer_GetProfile_withName(t *testing.T) {
	mockProfileService := sophrosyne2.NewMockProfileService(t)
	mockAuthzProvider := sophrosyne2.NewMockAuthorizationProvider(t)
	server := ProfileServiceServer{
		profileService: mockProfileService,
		authzProvider:  mockAuthzProvider,
	}

	user := sophrosyne.User{
		ID: "test",
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, sophrosyne.UserContextKey{}, &user)

	request := &v0.GetProfileRequest{
		Profile: &v0.GetProfileRequest_Name{
			Name: "test",
		},
	}

	mockProfileService.On("GetProfileByName", ctx, request.GetName()).Return(sophrosyne.Profile{
		ID:   "test",
		Name: "test",
	}, nil)
	mockAuthzProvider.On("IsAuthorized", ctx, sophrosyne.AuthorizationRequest{
		Principal: &user,
		Action:    sophrosyne.AuthorizationAction("GetProfile"),
		Resource:  sophrosyne.Profile{ID: "test"},
	}).Return(true)

	response, err := server.GetProfile(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.True(t, mockProfileService.AssertExpectations(t))
	require.True(t, mockAuthzProvider.AssertExpectations(t))
}

// GetProfile runs successfully with ID.
func Test_ProfileServiceServer_GetProfile_withID(t *testing.T) {
	mockProfileService := sophrosyne2.NewMockProfileService(t)
	mockAuthzProvider := sophrosyne2.NewMockAuthorizationProvider(t)
	server := ProfileServiceServer{
		profileService: mockProfileService,
		authzProvider:  mockAuthzProvider,
	}

	user := sophrosyne.User{
		ID: "test",
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, sophrosyne.UserContextKey{}, &user)

	request := &v0.GetProfileRequest{
		Profile: &v0.GetProfileRequest_Id{
			Id: "someID",
		},
	}

	mockProfileService.On("GetProfile", ctx, request.GetId()).Return(sophrosyne.Profile{
		ID:   "test",
		Name: "test",
	}, nil)
	mockAuthzProvider.On("IsAuthorized", ctx, sophrosyne.AuthorizationRequest{
		Principal: &user,
		Action:    sophrosyne.AuthorizationAction("GetProfile"),
		Resource:  sophrosyne.Profile{ID: "test"},
	}).Return(true)

	response, err := server.GetProfile(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.True(t, mockProfileService.AssertExpectations(t))
	require.True(t, mockAuthzProvider.AssertExpectations(t))
}

// GetProfile returns an error if the user is not authorized.
func Test_ProfileServiceServer_GetProfile_UserNotAuthorized(t *testing.T) {
	mockProfileService := sophrosyne2.NewMockProfileService(t)
	mockAuthzProvider := sophrosyne2.NewMockAuthorizationProvider(t)
	server := ProfileServiceServer{
		profileService: mockProfileService,
		authzProvider:  mockAuthzProvider,
	}

	user := sophrosyne.User{
		ID: "test",
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, sophrosyne.UserContextKey{}, &user)

	request := &v0.GetProfileRequest{
		Profile: &v0.GetProfileRequest_Name{
			Name: "test",
		},
	}

	mockProfileService.On("GetProfileByName", ctx, request.GetName()).Return(sophrosyne.Profile{
		ID:   "test",
		Name: "test",
	}, nil)
	mockAuthzProvider.On("IsAuthorized", ctx, sophrosyne.AuthorizationRequest{
		Principal: &user,
		Action:    sophrosyne.AuthorizationAction("GetProfile"),
		Resource:  sophrosyne.Profile{ID: "test"},
	}).Return(false)

	response, err := server.GetProfile(ctx, request)

	require.Error(t, err)
	grpcStatus, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.PermissionDenied, grpcStatus.Code())
	require.Equal(t, "unauthorized", grpcStatus.Message())
	require.Nil(t, response)
	require.True(t, mockProfileService.AssertExpectations(t))
	require.True(t, mockAuthzProvider.AssertExpectations(t))
}

// GetProfile returns an error if the profile is not found.
func Test_ProfileServiceServer_GetProfile_ProfileNotFound(t *testing.T) {
	mockProfileService := sophrosyne2.NewMockProfileService(t)
	server := ProfileServiceServer{
		profileService: mockProfileService,
	}

	user := sophrosyne.User{
		ID: "test",
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, sophrosyne.UserContextKey{}, &user)

	request := &v0.GetProfileRequest{
		Profile: &v0.GetProfileRequest_Name{
			Name: "test",
		},
	}

	mockProfileService.On("GetProfileByName", ctx, request.GetName()).Return(sophrosyne.Profile{}, assert.AnError)

	response, err := server.GetProfile(ctx, request)

	require.Error(t, err)
	grpcStatus, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Internal, grpcStatus.Code())
	require.Contains(t, grpcStatus.Message(), "error getting profile:")
	require.Nil(t, response)
	require.True(t, mockProfileService.AssertExpectations(t))
}

// GetProfile returns an authentication error if the user is not attached to the context.
func Test_ProfileServiceServer_GetProfile_NoUserInContext(t *testing.T) {
	server := ProfileServiceServer{}

	response, err := server.GetProfile(context.Background(), &v0.GetProfileRequest{})

	require.Error(t, err)
	grpcStatus, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Unauthenticated, grpcStatus.Code())
	require.Equal(t, InvalidTokenMsg, grpcStatus.Message())
	require.Nil(t, response)
}
