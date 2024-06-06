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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/madsrc/sophrosyne"
	sophrosyne2 "github.com/madsrc/sophrosyne/internal/mocks"
)

func TestWithAuthorizationProvider(t *testing.T) {
	type testStruct struct {
		authzProvider sophrosyne.AuthorizationProvider
	}
	cases := []struct {
		name   string
		target any
	}{
		{
			"UserServiceServer",
			&UserServiceServer{},
		},
		{
			"ProfileServiceServer",
			&ProfileServiceServer{},
		},
		{
			"CheckServiceServer",
			&CheckServiceServer{},
		},
		{
			"Unknown type",
			&testStruct{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			v := sophrosyne2.NewMockAuthorizationProvider(t)
			option := WithAuthorizationProvider(v)
			option(c.target)

			require.NotNil(t, c.target)

			switch c.target.(type) {
			case *UserServiceServer:
				require.NotNil(t, c.target.(*UserServiceServer).authzProvider)
				require.Equal(t, v, c.target.(*UserServiceServer).authzProvider)
			case *ProfileServiceServer:
				require.NotNil(t, c.target.(*ProfileServiceServer).authzProvider)
				require.Equal(t, v, c.target.(*ProfileServiceServer).authzProvider)
			case *CheckServiceServer:
				require.NotNil(t, c.target.(*CheckServiceServer).authzProvider)
				require.Equal(t, v, c.target.(*CheckServiceServer).authzProvider)
			default:
				require.Nil(t, c.target.(*testStruct).authzProvider)
			}
		})
	}
}

// correctly assigns checkService to CheckServiceServer
func TestWithAuthorizationProvider_CorrectlyAssignsCheckService(t *testing.T) {
	// Create a mock CheckService
	mockAuthzProvider := sophrosyne2.NewMockAuthorizationProvider(t)

	// Create a new CheckServiceServer instance with initial values
	initialServer := &CheckServiceServer{
		logger:    &slog.Logger{},
		config:    &sophrosyne.Config{},
		validator: sophrosyne2.NewMockValidator(t),
	}

	// Call the WithAuthorizationProvider function
	option := WithAuthorizationProvider(mockAuthzProvider)
	option(initialServer)

	// Assert that the authzProvider field has been correctly assigned
	require.Equal(t, mockAuthzProvider, initialServer.authzProvider)
}

// handles nil authzProvider gracefully
func TestWithAuthorizationProvider_HandlesNilCheckServiceGracefully(t *testing.T) {
	// Create a new CheckServiceServer instance with initial values
	initialServer := &CheckServiceServer{
		logger:    &slog.Logger{},
		config:    &sophrosyne.Config{},
		validator: sophrosyne2.NewMockValidator(t),
	}

	// Call the WithCheckService function with nil checkService
	option := WithAuthorizationProvider(nil)
	option(initialServer)

	// Assert that the checkService field is nil
	require.Nil(t, initialServer.checkService)
}

// Does nothing if provided an unsupported type
func TestWithAuthorizationProvider_DoesNothingIfUnsupportedType(t *testing.T) {

	// Call the WithCheckService function with unsupported type
	option := WithAuthorizationProvider(sophrosyne2.NewMockAuthorizationProvider(t))
	option(nil)

	// Does not panic
}
