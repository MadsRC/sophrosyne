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
	"github.com/jackc/pgx/v5"
	"github.com/madsrc/sophrosyne"
	sophrosyne2 "github.com/madsrc/sophrosyne/internal/mocks"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/require"
	"log/slog"
	"testing"
)

// UserService is successfully created when all inputs are valid
func TestNewUserService_CreationSuccess(t *testing.T) {
	ctx := context.Background()
	config := &sophrosyne.Config{
		Database: struct {
			User     string `key:"user" validate:"required"`
			Password string `key:"password" validate:"required"`
			Host     string `key:"host" validate:"required"`
			Port     int    `key:"port" validate:"required,min=1,max=65535"`
			Name     string `key:"name" validate:"required"`
		}{
			User: "testuser", Password: "testpass", Host: "localhost", Port: 5432, Name: "testdb",
		},
	}
	buf := bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	randomSource := rand.Reader
	profileService := &sophrosyne2.MockProfileService{}

	mockedDb, err := pgxmock.NewPool()
	require.NoError(t, err)
	mockedDb.ExpectBegin()
	mockedDb.ExpectQuery("").WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))
	mockedDb.ExpectRollback()

	userService, err := NewUserService(ctx, config, logger, randomSource, profileService, mockedDb)

	require.NoError(t, err)
	require.NotNil(t, userService)

	require.NoError(t, mockedDb.ExpectationsWereMet())
}

// Database connection fails during pool creation
func TestNewUserService_CreationDatabaseFailure(t *testing.T) {
	ctx := context.Background()
	config := &sophrosyne.Config{
		Database: struct {
			User     string `key:"user" validate:"required"`
			Password string `key:"password" validate:"required"`
			Host     string `key:"host" validate:"required"`
			Port     int    `key:"port" validate:"required,min=1,max=65535"`
			Name     string `key:"name" validate:"required"`
		}{
			User: "invaliduser", Password: "invalidpass", Host: "invalidhost", Port: -5, Name: "invaliddb",
		},
	}
	buf := bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	randomSource := rand.Reader
	profileService := &sophrosyne2.MockProfileService{}

	userService, err := NewUserService(ctx, config, logger, randomSource, profileService, nil)

	require.Error(t, err)
	require.Nil(t, userService)
}

// Database query fails when running createRootUser
func TestNewUserService_CreationFailure(t *testing.T) {
	ctx := context.Background()
	config := &sophrosyne.Config{
		Database: struct {
			User     string `key:"user" validate:"required"`
			Password string `key:"password" validate:"required"`
			Host     string `key:"host" validate:"required"`
			Port     int    `key:"port" validate:"required,min=1,max=65535"`
			Name     string `key:"name" validate:"required"`
		}{
			User: "testuser", Password: "testpass", Host: "localhost", Port: 5432, Name: "testdb",
		},
	}
	buf := bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	randomSource := rand.Reader
	profileService := &sophrosyne2.MockProfileService{}

	mockedDb, err := pgxmock.NewPool()
	require.NoError(t, err)
	mockedDb.ExpectBegin()
	mockedDb.ExpectQuery("").WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).WillReturnError(pgx.ErrNoRows)
	mockedDb.ExpectRollback()

	userService, err := NewUserService(ctx, config, logger, randomSource, profileService, mockedDb)

	require.ErrorIs(t, err, pgx.ErrNoRows)
	require.Nil(t, userService)

	require.NoError(t, mockedDb.ExpectationsWereMet())
}
