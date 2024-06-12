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
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/madsrc/sophrosyne"
)

// Verify successful pool creation with valid configuration.
func TestNewPool_Success(t *testing.T) {
	ctx := context.Background()
	config := &sophrosyne.Config{
		Database: struct {
			User     string `key:"user" validate:"required"`
			Password string `key:"password" validate:"required"`
			Host     string `key:"host" validate:"required"`
			Port     int    `key:"port" validate:"required,min=1,max=65535"`
			Name     string `key:"name" validate:"required"`
		}{
			User:     "testuser",
			Password: "testpass",
			Host:     "localhost",
			Port:     5432,
			Name:     "testdb",
		},
	}
	pool, err := newPool(ctx, config, nil)
	require.NoError(t, err)
	require.NotNil(t, pool)
}

// Verify failure pool creation with invalid configuration.
func TestNewPool_Failure(t *testing.T) {
	ctx := context.Background()
	config := &sophrosyne.Config{}
	pool, err := newPool(ctx, config, nil)
	require.Error(t, err)
	require.Nil(t, pool)
}
