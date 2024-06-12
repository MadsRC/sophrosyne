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
func Test_CheckServiceServer_GetChecks_withoutCursor(t *testing.T) {
	user := sophrosyne.User{
		ID:   "1",
		Name: "test",
	}

	ctx := context.WithValue(context.Background(), sophrosyne.UserContextKey{}, &user)

	mockCheckService := sophrosyne2.NewMockCheckService(t)
	mockAuthzProvider := sophrosyne2.NewMockAuthorizationProvider(t)
	server := CheckServiceServer{
		checkService:  mockCheckService,
		authzProvider: mockAuthzProvider,
	}

	retChecks := []sophrosyne.Check{
		{
			ID:   "test",
			Name: "test",
		},
		{
			ID:   "test2",
			Name: "test2",
		},
	}

	mockCheckService.On("GetChecks", ctx, mock.AnythingOfType("*sophrosyne.DatabaseCursor")).Return(retChecks, nil)
	mockAuthzProvider.On("IsAuthorized", ctx, sophrosyne.AuthorizationRequest{
		Principal: &user,
		Action:    sophrosyne.AuthorizationAction("GetCheck"),
		Resource:  sophrosyne.Check{ID: retChecks[0].ID},
	}).Once().Return(true)
	mockAuthzProvider.On("IsAuthorized", ctx, sophrosyne.AuthorizationRequest{
		Principal: &user,
		Action:    sophrosyne.AuthorizationAction("GetCheck"),
		Resource:  sophrosyne.Check{ID: retChecks[1].ID},
	}).Once().Return(true)

	out, err := server.GetChecks(ctx, &v0.GetChecksRequest{})
	require.NoError(t, err)

	require.Len(t, out.Checks, 2)
	require.Equal(t, retChecks[0].Name, out.Checks[0].Name)
	require.Equal(t, retChecks[1].Name, out.Checks[1].Name)
}

// test that providing a cursor works.
func Test_CheckServiceServer_GetChecks_withCursor(t *testing.T) {
	logger, _ := log.NewTestLogger(nil)
	user := sophrosyne.User{
		ID:   "couk317om0jj94nnh130",
		Name: "test",
	}

	ctx := context.WithValue(context.Background(), sophrosyne.UserContextKey{}, &user)

	mockCheckService := sophrosyne2.NewMockCheckService(t)
	mockAuthzProvider := sophrosyne2.NewMockAuthorizationProvider(t)
	server := CheckServiceServer{
		checkService:  mockCheckService,
		authzProvider: mockAuthzProvider,
		logger:        logger,
	}

	retChecks := []sophrosyne.Check{
		{
			ID:   "cp9o4nl060j00mieonmg",
			Name: "test",
		},
		{
			ID:   "test2",
			Name: "test2",
		},
	}

	cursor := sophrosyne.NewDatabaseCursor(user.ID, retChecks[0].ID)

	mockCheckService.On("GetChecks", ctx, mock.AnythingOfType("*sophrosyne.DatabaseCursor")).Return(retChecks, nil)
	mockAuthzProvider.On("IsAuthorized", ctx, sophrosyne.AuthorizationRequest{
		Principal: &user,
		Action:    sophrosyne.AuthorizationAction("GetCheck"),
		Resource:  sophrosyne.Check{ID: retChecks[0].ID},
	}).Once().Return(true)
	mockAuthzProvider.On("IsAuthorized", ctx, sophrosyne.AuthorizationRequest{
		Principal: &user,
		Action:    sophrosyne.AuthorizationAction("GetCheck"),
		Resource:  sophrosyne.Check{ID: retChecks[1].ID},
	}).Once().Return(true)

	out, err := server.GetChecks(ctx, &v0.GetChecksRequest{
		Cursor: cursor.String(),
	})
	require.NoError(t, err)

	require.Len(t, out.Checks, 2)
	require.Equal(t, cursor.String(), out.Cursor)
	require.Equal(t, retChecks[0].Name, out.Checks[0].Name)
	require.Equal(t, retChecks[1].Name, out.Checks[1].Name)
}

// test that GetChecks fails.
func Test_CheckServiceServer_GetChecks_cursorIsNowNil(t *testing.T) {
	logger, _ := log.NewTestLogger(nil)
	user := sophrosyne.User{
		ID:   "couk317om0jj94nnh130",
		Name: "test",
	}

	ctx := context.WithValue(context.Background(), sophrosyne.UserContextKey{}, &user)

	mockCheckService := sophrosyne2.NewMockCheckService(t)
	server := CheckServiceServer{
		checkService: mockCheckService,
		logger:       logger,
	}

	mockCheckService.On("GetChecks", ctx, mock.AnythingOfType("*sophrosyne.DatabaseCursor")).Return(nil, assert.AnError)

	out, err := server.GetChecks(ctx, &v0.GetChecksRequest{})
	require.Nil(t, out)
	require.Error(t, err)
	grpcStatus, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Internal, grpcStatus.Code())
	require.Contains(t, grpcStatus.Message(), "internal error getting checks")
}

// test that invalid cursor returns error.
func Test_CheckServiceServer_GetChecks_invalidCursor(t *testing.T) {
	logger, _ := log.NewTestLogger(nil)
	user := sophrosyne.User{
		ID:   "couk317om0jj94nnh130",
		Name: "test",
	}

	ctx := context.WithValue(context.Background(), sophrosyne.UserContextKey{}, &user)

	server := CheckServiceServer{
		logger: logger,
	}

	out, err := server.GetChecks(ctx, &v0.GetChecksRequest{
		Cursor: "invalidCursor",
	})
	require.Nil(t, out)
	require.Error(t, err)
}

// test that no user in context returns authentication error.
func Test_CheckServiceServer_GetChecks_noUserContext(t *testing.T) {
	logger, _ := log.NewTestLogger(nil)

	ctx := context.Background()

	server := CheckServiceServer{
		logger: logger,
	}

	out, err := server.GetChecks(ctx, &v0.GetChecksRequest{})
	require.Nil(t, out)
	require.Error(t, err)
	grpcStatus, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Unauthenticated, grpcStatus.Code())
	require.Equal(t, InvalidTokenMsg, grpcStatus.Message())
}
