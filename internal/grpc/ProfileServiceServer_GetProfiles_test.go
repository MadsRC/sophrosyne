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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/madsrc/sophrosyne"
	v0 "github.com/madsrc/sophrosyne/internal/grpc/sophrosyne/v0"
	"github.com/madsrc/sophrosyne/internal/log"
	sophrosyne2 "github.com/madsrc/sophrosyne/internal/mocks"
)

// test that not providing a cursor works.
func Test_ProfileServiceServer_GetProfiles_withoutCursor(t *testing.T) {
	user := sophrosyne.User{
		ID:   "1",
		Name: "test",
	}

	ctx := context.WithValue(context.Background(), sophrosyne.UserContextKey{}, &user)

	mockProfileService := sophrosyne2.NewMockProfileService(t)
	mockAuthzProvider := sophrosyne2.NewMockAuthorizationProvider(t)
	server := ProfileServiceServer{
		profileService: mockProfileService,
		authzProvider:  mockAuthzProvider,
	}

	retProfiles := []sophrosyne.Profile{
		{
			ID:   "test",
			Name: "test",
		},
		{
			ID:   "test2",
			Name: "test2",
		},
	}

	mockProfileService.On("GetProfiles", ctx, mock.AnythingOfType("*sophrosyne.DatabaseCursor")).Return(retProfiles, nil)
	mockAuthzProvider.On("IsAuthorized", ctx, sophrosyne.AuthorizationRequest{
		Principal: &user,
		Action:    sophrosyne.AuthorizationAction("GetProfile"),
		Resource:  sophrosyne.Profile{ID: retProfiles[0].ID},
	}).Once().Return(true)
	mockAuthzProvider.On("IsAuthorized", ctx, sophrosyne.AuthorizationRequest{
		Principal: &user,
		Action:    sophrosyne.AuthorizationAction("GetProfile"),
		Resource:  sophrosyne.Profile{ID: retProfiles[1].ID},
	}).Once().Return(true)

	out, err := server.GetProfiles(ctx, &v0.GetProfilesRequest{})
	require.NoError(t, err)

	require.Len(t, out.Profiles, 2)
	require.Equal(t, retProfiles[0].Name, out.Profiles[0].Name)
	require.Equal(t, retProfiles[1].Name, out.Profiles[1].Name)
}

// test that providing a cursor works.
func Test_ProfileServiceServer_GetProfiles_withCursor(t *testing.T) {
	logger, _ := log.NewTestLogger(nil)
	user := sophrosyne.User{
		ID:   "couk317om0jj94nnh130",
		Name: "test",
	}

	ctx := context.WithValue(context.Background(), sophrosyne.UserContextKey{}, &user)

	mockProfileService := sophrosyne2.NewMockProfileService(t)
	mockAuthzProvider := sophrosyne2.NewMockAuthorizationProvider(t)
	server := ProfileServiceServer{
		profileService: mockProfileService,
		authzProvider:  mockAuthzProvider,
		logger:         logger,
	}

	retProfiles := []sophrosyne.Profile{
		{
			ID:   "cp9o4nl060j00mieonmg",
			Name: "test",
		},
		{
			ID:   "test2",
			Name: "test2",
		},
	}

	cursor := sophrosyne.NewDatabaseCursor(user.ID, retProfiles[0].ID)

	mockProfileService.On("GetProfiles", ctx, mock.AnythingOfType("*sophrosyne.DatabaseCursor")).Return(retProfiles, nil)
	mockAuthzProvider.On("IsAuthorized", ctx, sophrosyne.AuthorizationRequest{
		Principal: &user,
		Action:    sophrosyne.AuthorizationAction("GetProfile"),
		Resource:  sophrosyne.Profile{ID: retProfiles[0].ID},
	}).Once().Return(true)
	mockAuthzProvider.On("IsAuthorized", ctx, sophrosyne.AuthorizationRequest{
		Principal: &user,
		Action:    sophrosyne.AuthorizationAction("GetProfile"),
		Resource:  sophrosyne.Profile{ID: retProfiles[1].ID},
	}).Once().Return(true)

	out, err := server.GetProfiles(ctx, &v0.GetProfilesRequest{
		Cursor: cursor.String(),
	})
	require.NoError(t, err)

	require.Len(t, out.Profiles, 2)
	require.Equal(t, cursor.String(), out.Cursor)
	require.Equal(t, retProfiles[0].Name, out.Profiles[0].Name)
	require.Equal(t, retProfiles[1].Name, out.Profiles[1].Name)
}

// test that GetProfiles fails.
func Test_ProfileServiceServer_GetProfiles_cursorIsNowNil(t *testing.T) {
	logger, _ := log.NewTestLogger(nil)
	user := sophrosyne.User{
		ID:   "couk317om0jj94nnh130",
		Name: "test",
	}

	ctx := context.WithValue(context.Background(), sophrosyne.UserContextKey{}, &user)

	mockProfileService := sophrosyne2.NewMockProfileService(t)
	server := ProfileServiceServer{
		profileService: mockProfileService,
		logger:         logger,
	}

	mockProfileService.On("GetProfiles", ctx, mock.AnythingOfType("*sophrosyne.DatabaseCursor")).Return(nil, assert.AnError)

	out, err := server.GetProfiles(ctx, &v0.GetProfilesRequest{})
	require.Nil(t, out)
	require.Error(t, err)
	grpcStatus, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Internal, grpcStatus.Code())
	require.Contains(t, grpcStatus.Message(), "internal error getting profiles")
}

// test that invalid cursor returns error.
func Test_ProfileServiceServer_GetProfiles_invalidCursor(t *testing.T) {
	logger, _ := log.NewTestLogger(nil)
	user := sophrosyne.User{
		ID:   "couk317om0jj94nnh130",
		Name: "test",
	}

	ctx := context.WithValue(context.Background(), sophrosyne.UserContextKey{}, &user)

	server := ProfileServiceServer{
		logger: logger,
	}

	out, err := server.GetProfiles(ctx, &v0.GetProfilesRequest{
		Cursor: "invalidCursor",
	})
	require.Nil(t, out)
	require.Error(t, err)
}

// test that no user in context returns authentication error.
func Test_ProfileServiceServer_GetProfiles_noUserContext(t *testing.T) {
	logger, _ := log.NewTestLogger(nil)

	ctx := context.Background()

	server := ProfileServiceServer{
		logger: logger,
	}

	out, err := server.GetProfiles(ctx, &v0.GetProfilesRequest{})
	require.Nil(t, out)
	require.Error(t, err)
	grpcStatus, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Unauthenticated, grpcStatus.Code())
	require.Equal(t, InvalidTokenMsg, grpcStatus.Message())
}
