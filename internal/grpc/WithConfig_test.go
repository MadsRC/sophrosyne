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

func TestWithConfig(t *testing.T) {
	type testStruct struct {
		config *sophrosyne.Config
	}
	cases := []struct {
		name   string
		target any
	}{
		{
			"ScanServiceServer",
			&ScanServiceServer{},
		},
		{
			"Server",
			&Server{},
		},
		{
			"UserServiceServer",
			&UserServiceServer{},
		},
		{
			"CheckServiceServer",
			&CheckServiceServer{},
		},
		{
			"ProfileServiceServer",
			&ProfileServiceServer{},
		},
		{
			"Unknown type",
			&testStruct{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			option := WithConfig(&sophrosyne.Config{
				Server: sophrosyne.ServerConfig{
					Port: 2500,
				},
			})
			option(c.target)

			require.NotNil(t, c.target)

			switch c.target.(type) {
			case *ScanServiceServer:
				require.NotNil(t, c.target.(*ScanServiceServer).Config)
				require.Equal(t, 2500, c.target.(*ScanServiceServer).Config.Server.Port)
			case *Server:
				require.NotNil(t, c.target.(*Server).config)
				require.Equal(t, 2500, c.target.(*Server).config.Server.Port)
			case *UserServiceServer:
				require.NotNil(t, c.target.(*UserServiceServer).config)
				require.Equal(t, 2500, c.target.(*UserServiceServer).config.Server.Port)
			case *CheckServiceServer:
				require.NotNil(t, c.target.(*CheckServiceServer).config)
				require.Equal(t, 2500, c.target.(*CheckServiceServer).config.Server.Port)
			case *ProfileServiceServer:
				require.NotNil(t, c.target.(*ProfileServiceServer).config)
				require.Equal(t, 2500, c.target.(*ProfileServiceServer).config.Server.Port)
			default:
				require.Nil(t, c.target.(*testStruct).config)
			}
		})
	}
}

// Sets the server's Config field when a valid Config is provided.
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

	// Assert that only the Config field has been altered
	require.Equal(t, newConfig, initialServer.config)
	require.NotNil(t, initialServer.grpcServer)
	require.NotNil(t, initialServer.listener)
	require.NotNil(t, initialServer.logger)
	require.NotNil(t, initialServer.validator)
}

// Returns a Option function that can be applied to a Server instance.
func TestWithConfig_ReturnsOption(t *testing.T) {
	// Create a dummy Config
	dummyConfig := &sophrosyne.Config{}

	// Call the WithConfig function
	option := WithConfig(dummyConfig)

	// Create a dummy Server instance
	server := &Server{}

	// Apply the Option function to the Server instance
	option(server)

	// Check if the Config field of the Server instance is set to the dummy Config
	require.Equal(t, dummyConfig, server.config)
}

// Handles a nil Config gracefully without causing a panic.
func TestWithConfig_NilConfig(t *testing.T) {
	s := &ScanServiceServer{}
	opt := WithConfig(nil)
	opt(s)

	require.Nil(t, s.Config)
}

// Ensures the server's Config field is set to nil if a nil Config is provided.
func TestWithConfig_NilConfigProvided(t *testing.T) {
	// Setup
	s := &Server{}

	// Execution
	opt := WithConfig(nil)
	opt(s)

	// Assertion
	require.Nil(t, s.config)
}

// Validates that the Config field is correctly assigned in the server struct.
func TestWithConfig_ConfigAssigned(t *testing.T) {
	// Create a new Server
	s := &Server{}

	// Create a new Config
	config := &sophrosyne.Config{}

	// Call the WithConfig function
	option := WithConfig(config)
	option(s)

	// Check if the Config field in the Server struct is correctly assigned
	require.Equal(t, config, s.config)
}

// Ensures idempotency when the same Config is applied multiple times.
func TestWithConfig_Idempotency(t *testing.T) {
	// Create a new Server instance
	server := &Server{}

	// Create a sample Config
	config := &sophrosyne.Config{}

	// Apply the Config using WithConfig function twice
	WithConfig(config)(server)
	WithConfig(config)(server)

	// Assert that the Config is applied only once
	require.Equal(t, config, server.config)
}

// Confirms that the function does not modify the input Config.
func TestWithConfig_DoesNotModifyInputConfig(t *testing.T) {
	// Create a sample Config
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

	// Call the WithConfig function with the sample Config
	option := WithConfig(sampleConfig)
	option(server)

	// Assert that the Config in the server instance is the same as the sample Config
	require.Equal(t, sampleConfig, server.config)
}
