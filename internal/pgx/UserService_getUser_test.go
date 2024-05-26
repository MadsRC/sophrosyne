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
	"github.com/madsrc/sophrosyne"
	sophrosyne2 "github.com/madsrc/sophrosyne/internal/mocks"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/require"
	"log/slog"
	"testing"
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
