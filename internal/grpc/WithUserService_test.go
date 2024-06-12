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

func TestWithUserService(t *testing.T) {
	type testStruct struct {
		userService sophrosyne.UserService
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
			"Unknown type",
			&testStruct{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			v := sophrosyne2.NewMockUserService(t)
			option := WithUserService(v)
			option(c.target)

			require.NotNil(t, c.target)

			switch c.target.(type) {
			case *UserServiceServer:
				require.NotNil(t, c.target.(*UserServiceServer).userService)
				require.Equal(t, v, c.target.(*UserServiceServer).userService)
			default:
				require.Nil(t, c.target.(*testStruct).userService)
			}
		})
	}
}

// correctly assigns userService to UserServiceServer.
func TestWithUserService_CorrectlyAssignsUserService(t *testing.T) {
	// Create a mock UserService
	mockUserService := sophrosyne2.NewMockUserService(t)

	// Create a new UserServiceServer instance with initial values
	initialServer := &UserServiceServer{
		logger:    &slog.Logger{},
		config:    &sophrosyne.Config{},
		validator: sophrosyne2.NewMockValidator(t),
	}

	// Call the WithUserService function
	option := WithUserService(mockUserService)
	option(initialServer)

	// Assert that the userService field has been correctly assigned
	require.Equal(t, mockUserService, initialServer.userService)
}

// handles nil userService gracefully.
func TestWithUserService_HandlesNilUserServiceGracefully(t *testing.T) {
	// Create a new UserServiceServer instance with initial values
	initialServer := &UserServiceServer{
		logger:    &slog.Logger{},
		config:    &sophrosyne.Config{},
		validator: sophrosyne2.NewMockValidator(t),
	}

	// Call the WithUserService function with nil userService
	option := WithUserService(nil)
	option(initialServer)

	// Assert that the userService field is nil
	require.Nil(t, initialServer.userService)
}

// Does nothing if provided an unsupported type.
func TestWithUserService_DoesNothingIfUnsupportedType(t *testing.T) {

	// Call the WithUserService function with unsupported type
	option := WithUserService(sophrosyne2.NewMockUserService(t))
	option(nil)

	// Does not panic
}
