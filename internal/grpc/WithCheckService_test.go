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

func TestWithCheckService(t *testing.T) {
	type testStruct struct {
		checkService sophrosyne.CheckService
	}
	cases := []struct {
		name   string
		target any
	}{
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
			v := sophrosyne2.NewMockCheckService(t)
			option := WithCheckService(v)
			option(c.target)

			require.NotNil(t, c.target)

			switch c.target.(type) {
			case *CheckServiceServer:
				require.NotNil(t, c.target.(*CheckServiceServer).checkService)
				require.Equal(t, v, c.target.(*CheckServiceServer).checkService)
			default:
				require.Nil(t, c.target.(*testStruct).checkService)
			}
		})
	}
}

// correctly assigns checkService to CheckServiceServer.
func TestWithCheckService_CorrectlyAssignsCheckService(t *testing.T) {
	// Create a mock CheckService
	mockCheckService := sophrosyne2.NewMockCheckService(t)

	// Create a new CheckServiceServer instance with initial values
	initialServer := &CheckServiceServer{
		logger:    &slog.Logger{},
		config:    &sophrosyne.Config{},
		validator: sophrosyne2.NewMockValidator(t),
	}

	// Call the WithCheckService function
	option := WithCheckService(mockCheckService)
	option(initialServer)

	// Assert that the checkService field has been correctly assigned
	require.Equal(t, mockCheckService, initialServer.checkService)
}

// handles nil checkService gracefully.
func TestWithCheckService_HandlesNilCheckServiceGracefully(t *testing.T) {
	// Create a new CheckServiceServer instance with initial values
	initialServer := &CheckServiceServer{
		logger:    &slog.Logger{},
		config:    &sophrosyne.Config{},
		validator: sophrosyne2.NewMockValidator(t),
	}

	// Call the WithCheckService function with nil checkService
	option := WithCheckService(nil)
	option(initialServer)

	// Assert that the checkService field is nil
	require.Nil(t, initialServer.checkService)
}

// Does nothing if provided an unsupported type.
func TestWithCheckService_DoesNothingIfUnsupportedType(t *testing.T) {

	// Call the WithCheckService function with unsupported type
	option := WithCheckService(sophrosyne2.NewMockCheckService(t))
	option(nil)

	// Does not panic
}
