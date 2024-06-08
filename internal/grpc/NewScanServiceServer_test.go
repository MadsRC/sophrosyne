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
	"github.com/madsrc/sophrosyne/internal/log"
	sophrosyne2 "github.com/madsrc/sophrosyne/internal/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"log/slog"
	"testing"
)

func TestNewScanServiceServer(t *testing.T) {
	logger, _ := log.NewTestLogger(nil)
	config := &sophrosyne.Config{}
	validator := sophrosyne2.NewMockValidator(t)
	profileService := sophrosyne2.NewMockProfileService(t)

	validator.On("Validate", mock.AnythingOfType("*grpc.ScanServiceServer")).Return(nil)

	options := []Option{
		WithLogger(logger),
		WithConfig(config),
		WithValidator(validator),
		WithProfileService(profileService),
	}

	server, err := NewScanServiceServer(context.Background(), options...)

	require.NotNil(t, server)
	require.NoError(t, err)
	require.Equal(t, logger, server.Logger)
	require.Equal(t, config, server.Config)
	require.Equal(t, validator, server.Validator)
	require.Equal(t, profileService, server.ProfileService)

	require.True(t, validator.AssertExpectations(t))
}

func TestNewScanServiceServer_RequiredOptions(t *testing.T) {
	newTestLogger := func() *slog.Logger {
		logger, _ := log.NewTestLogger(nil)
		return logger
	}
	cases := []struct {
		name    string
		options []Option
	}{
		{
			name:    "no options",
			options: []Option{},
		},
		{
			name: "all but logger",
			options: []Option{
				WithConfig(&sophrosyne.Config{}),
				WithProfileService(sophrosyne2.NewMockProfileService(t)),
			},
		},
		{
			name: "all but config",
			options: []Option{
				WithLogger(newTestLogger()),
				WithProfileService(sophrosyne2.NewMockProfileService(t)),
			},
		},
		{
			name: "all but profileService",
			options: []Option{
				WithLogger(newTestLogger()),
				WithConfig(&sophrosyne.Config{}),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server, err := NewScanServiceServer(context.Background(), tc.options...)
			require.Nil(t, server)
			require.Error(t, err)
			ve := &sophrosyne.ValidationError{}
			require.ErrorAs(t, err, &ve)
		})
	}
}

// Handles nil options gracefully when passed to NewScanServiceServer function.
func TestNewScanServiceServer_HandlesNilOptionsGracefully(t *testing.T) {
	// Call the NewScanServiceServer function with nil options
	server, err := NewScanServiceServer(context.Background(), nil)

	// Assert that the server is not nil
	require.Nil(t, server)

	// Assert that validation fails because of missing options
	require.Error(t, err)
}
