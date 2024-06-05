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
	"log/slog"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/madsrc/sophrosyne"
	"github.com/madsrc/sophrosyne/internal/validator"
)

// Sets the server's config field when a valid config is provided.
func TestWithConfig_SetsConfigField(t *testing.T) {
	config := &sophrosyne.Config{}
	server := &Server{}

	option := WithConfig(config)
	option(server)

	require.NotNil(t, server.config)
	require.Equal(t, config, server.config)
}

// Does nothing if the server instance is nil
func TestWithConfig_ServerIsNil(t *testing.T) {
	config := &sophrosyne.Config{}

	option := WithConfig(config)
	option(nil)

	// No panic or error should occur, and nothing to assert as server is nil
}

// Does not alter other fields of the server struct.
func TestWithConfig_DoesNotAlterOtherFields(t *testing.T) {
	// Create a new Server instance with some initial values
	initialServer := &Server{
		grpcServer: &grpc.Server{},
		listener:   &net.TCPListener{},
		config:     &sophrosyne.Config{},
		logger:     &slog.Logger{},
		validator:  &validator.Validator{},
	}

	// Create a new Config instance to be set
	newConfig := &sophrosyne.Config{}

	// Call the WithConfig function
	option := WithConfig(newConfig)
	option(initialServer)

	// Assert that only the config field has been altered
	require.Equal(t, newConfig, initialServer.config)
	require.NotNil(t, initialServer.grpcServer)
	require.NotNil(t, initialServer.listener)
	require.NotNil(t, initialServer.logger)
	require.NotNil(t, initialServer.validator)
}

// Returns a Option function that can be applied to a Server instance.
func TestWithConfig_ReturnsOption(t *testing.T) {
	// Create a dummy config
	dummyConfig := &sophrosyne.Config{}

	// Call the WithConfig function
	option := WithConfig(dummyConfig)

	// Create a dummy Server instance
	server := &Server{}

	// Apply the Option function to the Server instance
	option(server)

	// Check if the config field of the Server instance is set to the dummy config
	require.Equal(t, dummyConfig, server.config)
}

// Handles a nil config gracefully without causing a panic.
func TestWithConfig_NilConfig(t *testing.T) {
	s := &ScanServiceServer{}
	opt := WithConfig(nil)
	opt(s)

	require.Nil(t, s.config)
}

// Ensures the server's config field is set to nil if a nil config is provided.
func TestWithConfig_NilConfigProvided(t *testing.T) {
	// Setup
	s := &Server{}

	// Execution
	opt := WithConfig(nil)
	opt(s)

	// Assertion
	require.Nil(t, s.config)
}

// Validates that the config field is correctly assigned in the server struct.
func TestWithConfig_ConfigAssigned(t *testing.T) {
	// Create a new Server
	s := &Server{}

	// Create a new Config
	config := &sophrosyne.Config{}

	// Call the WithConfig function
	option := WithConfig(config)
	option(s)

	// Check if the config field in the Server struct is correctly assigned
	require.Equal(t, config, s.config)
}

// Ensures idempotency when the same config is applied multiple times.
func TestWithConfig_Idempotency(t *testing.T) {
	// Create a new Server instance
	server := &Server{}

	// Create a sample config
	config := &sophrosyne.Config{}

	// Apply the config using WithConfig function twice
	WithConfig(config)(server)
	WithConfig(config)(server)

	// Assert that the config is applied only once
	require.Equal(t, config, server.config)
}

// Confirms that the function does not modify the input config.
func TestWithConfig_DoesNotModifyInputConfig(t *testing.T) {
	// Create a sample config
	sampleConfig := &sophrosyne.Config{
		Principals: struct {
			Root struct {
				Name     string `key:"name" validate:"required"`
				Email    string `key:"email" validate:"required"`
				Recreate bool   `key:"recreate"`
			} `key:"root" validate:"required"`
		}{
			Root: struct {
				Name     string `key:"name" validate:"required"`
				Email    string `key:"email" validate:"required"`
				Recreate bool   `key:"recreate"`
			}{
				Name:     "John Doe",
				Email:    "john.doe@example.com",
				Recreate: false,
			},
		},
	}

	// Create a server instance
	server := &Server{}

	// Call the WithConfig function with the sample config
	option := WithConfig(sampleConfig)
	option(server)

	// Assert that the config in the server instance is the same as the sample config
	require.Equal(t, sampleConfig, server.config)
}
