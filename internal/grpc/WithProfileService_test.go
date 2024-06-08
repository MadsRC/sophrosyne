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

func TestWithProfileService(t *testing.T) {
	type testStruct struct {
		profileService sophrosyne.ProfileService
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
			v := sophrosyne2.NewMockProfileService(t)
			option := WithProfileService(v)
			option(c.target)

			require.NotNil(t, c.target)

			switch c.target.(type) {
			case *ScanServiceServer:
				require.NotNil(t, c.target.(*ScanServiceServer).ProfileService)
				require.Equal(t, v, c.target.(*ScanServiceServer).ProfileService)
			case *ProfileServiceServer:
				require.NotNil(t, c.target.(*ProfileServiceServer).profileService)
				require.Equal(t, v, c.target.(*ProfileServiceServer).profileService)
			default:
				require.Nil(t, c.target.(*testStruct).profileService)
			}
		})
	}
}

// correctly assigns ProfileService to ScanServiceServer
func TestWithProfileService_CorrectlyAssignsProfileService(t *testing.T) {
	// Create a mock ProfileService
	mockProfileService := sophrosyne2.NewMockProfileService(t)

	// Create a new ScanServiceServer instance with initial values
	initialServer := &ScanServiceServer{
		Logger:    &slog.Logger{},
		Config:    &sophrosyne.Config{},
		Validator: sophrosyne2.NewMockValidator(t),
	}

	// Call the WithProfileService function
	option := WithProfileService(mockProfileService)
	option(initialServer)

	// Assert that the ProfileService field has been correctly assigned
	require.Equal(t, mockProfileService, initialServer.ProfileService)
}

// handles nil ProfileService gracefully
func TestWithProfileService_HandlesNilProfileServiceGracefully(t *testing.T) {
	// Create a new ScanServiceServer instance with initial values
	initialServer := &ScanServiceServer{
		Logger:    &slog.Logger{},
		Config:    &sophrosyne.Config{},
		Validator: sophrosyne2.NewMockValidator(t),
	}

	// Call the WithProfileService function with nil ProfileService
	option := WithProfileService(nil)
	option(initialServer)

	// Assert that the ProfileService field is nil
	require.Nil(t, initialServer.ProfileService)
}

// Does nothing if provided an unsupported type
func TestWithProfileService_DoesNothingIfUnsupportedType(t *testing.T) {

	// Call the WithProfileService function with unsupported type
	option := WithProfileService(sophrosyne2.NewMockProfileService(t))
	option(nil)

	// Does not panic
}
