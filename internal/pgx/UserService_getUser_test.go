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

package pgx

import (
	"bytes"
	"context"
	"crypto/rand"
	"log/slog"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/madsrc/sophrosyne"
	sophrosyne2 "github.com/madsrc/sophrosyne/internal/mocks"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Verify correct user is returned when queried by email
func TestGetUser_ByEmail(t *testing.T) {
	ctx := context.Background()
	config := &sophrosyne.Config{}
	pool, err := pgxmock.NewPool()
	require.NoError(t, err)
	logger := slog.New(slog.NewJSONHandler(&bytes.Buffer{}, &slog.HandlerOptions{Level: slog.LevelInfo}))
	randomSource := rand.Reader
	profileService := sophrosyne2.NewMockProfileService(t)
	userService := UserService{
		config:         config,
		pool:           pool,
		logger:         logger,
		randomSource:   randomSource,
		profileService: profileService,
	}
	email := "test@example.com"
	expectedUser := sophrosyne.User{
		ID:    "123",
		Name:  "Test User",
		Email: email,
		DefaultProfile: sophrosyne.Profile{
			ID:   "someID",
			Name: "defaultProfile",
		},
	}

	pool.ExpectQuery("").WithArgs([]byte(email)).WillReturnRows(pgxmock.NewRows([]string{
		"id",
		"name",
		"email",
		"token",
		"is_admin",
		"default_profile",
		"created_at",
		"updated_at",
		"deleted_at",
	}).AddRow(
		expectedUser.ID,
		expectedUser.Name,
		expectedUser.Email,
		[]byte("token"),
		true,
		expectedUser.DefaultProfile.ID,
		expectedUser.CreatedAt,
		expectedUser.UpdatedAt,
		nil,
	))
	profileService.On("GetProfile", ctx, expectedUser.DefaultProfile.ID).Return(expectedUser.DefaultProfile, nil)

	user, err := userService.getUser(ctx, "email", []byte(email))
	require.NoError(t, err)
	require.Equal(t, expectedUser.Email, user.Email)

	require.NoError(t, pool.ExpectationsWereMet())
	require.True(t, profileService.AssertExpectations(t))
}

// Column not found in getUserQueryMap
func TestGetUser_ColumnNotFound(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxmock.NewPool()
	require.NoError(t, err)
	profileService := sophrosyne2.NewMockProfileService(t)
	userService := UserService{
		pool:           pool,
		profileService: profileService,
	}
	column := "invalid_column"
	input := []byte("test@example.com")

	_, err = userService.getUser(ctx, column, input)
	var uce *sophrosyne.UnreachableCodeError
	require.ErrorAs(t, err, &uce)
	require.NoError(t, pool.ExpectationsWereMet())
	require.True(t, profileService.AssertExpectations(t))
}

// Valid column and input returns a correct user with additional fields
func TestGetUser_ValidColumnAndInput_WithMoreFields(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxmock.NewPool()
	require.NoError(t, err)
	profileService := sophrosyne2.NewMockProfileService(t)
	userService := UserService{
		pool:           pool,
		profileService: profileService,
	}
	column := "email"
	input := []byte("test@example.com")
	expectedUser := sophrosyne.User{
		ID:      "123",
		Name:    "Test User",
		Email:   string(input),
		Token:   []byte("token123"),
		IsAdmin: true,
		DefaultProfile: sophrosyne.Profile{
			ID:   "someID",
			Name: "someName",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: nil,
	}
	pool.ExpectQuery("").WithArgs(input).WillReturnRows(pgxmock.NewRows([]string{
		"id",
		"name",
		"email",
		"token",
		"is_admin",
		"default_profile",
		"created_at",
		"updated_at",
		"deleted_at",
	}).AddRow(
		expectedUser.ID,
		expectedUser.Name,
		expectedUser.Email,
		expectedUser.Token,
		expectedUser.IsAdmin,
		expectedUser.DefaultProfile.ID,
		expectedUser.CreatedAt,
		expectedUser.UpdatedAt,
		expectedUser.DeletedAt,
	))
	profileService.On("GetProfile", ctx, expectedUser.DefaultProfile.ID).Return(expectedUser.DefaultProfile, nil)

	user, err := userService.getUser(ctx, column, input)
	require.NoError(t, err)
	require.Equal(t, expectedUser, user)

	require.NoError(t, pool.ExpectationsWereMet())
	require.True(t, profileService.AssertExpectations(t))
}

// Query execution returns exactly one row
func TestGetUser_WithValidQueryResult(t *testing.T) {
	ctx := context.Background()
	config := &sophrosyne.Config{}
	pool, err := pgxmock.NewPool()
	require.NoError(t, err)
	logger := slog.New(slog.NewJSONHandler(&bytes.Buffer{}, &slog.HandlerOptions{Level: slog.LevelInfo}))
	randomSource := rand.Reader
	profileService := sophrosyne2.NewMockProfileService(t)
	userService := UserService{
		config:         config,
		pool:           pool,
		logger:         logger,
		randomSource:   randomSource,
		profileService: profileService,
	}
	column := "email"
	input := []byte("test@example.com")
	expectedUser := sophrosyne.User{
		ID:    "123",
		Name:  "Test User",
		Email: "test@example.com",
		DefaultProfile: sophrosyne.Profile{
			ID:   "someID",
			Name: "defaultProfile",
		},
	}

	pool.ExpectQuery("").WithArgs(input).WillReturnRows(pgxmock.NewRows([]string{
		"id",
		"name",
		"email",
		"token",
		"is_admin",
		"default_profile",
		"created_at",
		"updated_at",
		"deleted_at",
	}).AddRow(
		expectedUser.ID,
		expectedUser.Name,
		expectedUser.Email,
		[]byte("token"),
		true,
		expectedUser.DefaultProfile.ID,
		expectedUser.CreatedAt,
		expectedUser.UpdatedAt,
		nil,
	))
	profileService.On("GetProfile", ctx, expectedUser.DefaultProfile.ID).Return(expectedUser.DefaultProfile, nil)

	user, err := userService.getUser(ctx, column, input)
	require.NoError(t, err)
	require.Equal(t, expectedUser.Email, user.Email)

	require.NoError(t, pool.ExpectationsWereMet())
	require.True(t, profileService.AssertExpectations(t))
}

func TestGetUser_QueryNoRows(t *testing.T) {
	ctx := context.Background()
	column := "email"
	input := []byte("test@example.com")
	config := &sophrosyne.Config{}
	pool, err := pgxmock.NewPool()
	require.NoError(t, err)
	logger := slog.New(slog.NewJSONHandler(&bytes.Buffer{}, &slog.HandlerOptions{Level: slog.LevelInfo}))
	randomSource := rand.Reader
	profileService := sophrosyne2.NewMockProfileService(t)
	userService := UserService{
		config:         config,
		pool:           pool,
		logger:         logger,
		randomSource:   randomSource,
		profileService: profileService,
	}
	pool.ExpectQuery("").WithArgs(input).WillReturnError(pgx.ErrNoRows)

	user, err := userService.getUser(ctx, column, input)
	require.ErrorIs(t, err, sophrosyne.ErrNotFound)
	require.Equal(t, sophrosyne.User{}, user)

	require.NoError(t, pool.ExpectationsWereMet())
	require.True(t, profileService.AssertExpectations(t))
}

func TestGetUser_BadError(t *testing.T) {
	ctx := context.Background()
	column := "email"
	input := []byte("test@example.com")
	config := &sophrosyne.Config{}
	pool, err := pgxmock.NewPool()
	require.NoError(t, err)
	logger := slog.New(slog.NewJSONHandler(&bytes.Buffer{}, &slog.HandlerOptions{Level: slog.LevelInfo}))
	randomSource := rand.Reader
	profileService := sophrosyne2.NewMockProfileService(t)
	userService := UserService{
		config:         config,
		pool:           pool,
		logger:         logger,
		randomSource:   randomSource,
		profileService: profileService,
	}
	pool.ExpectQuery("").WithArgs(input).WillReturnError(assert.AnError)

	user, err := userService.getUser(ctx, column, input)
	require.ErrorIs(t, err, assert.AnError)
	require.Equal(t, sophrosyne.User{}, user)

	require.NoError(t, pool.ExpectationsWereMet())
	require.True(t, profileService.AssertExpectations(t))
}
