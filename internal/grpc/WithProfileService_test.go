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

// correctly assigns profileService to ScanServiceServer
func TestWithProfileService_CorrectlyAssignsProfileService(t *testing.T) {
	// Create a mock ProfileService
	mockProfileService := sophrosyne2.NewMockProfileService(t)

	// Create a new ScanServiceServer instance with initial values
	initialServer := &ScanServiceServer{
		logger:    &slog.Logger{},
		config:    &sophrosyne.Config{},
		validator: sophrosyne2.NewMockValidator(t),
	}

	// Call the WithProfileService function
	option := WithProfileService(mockProfileService)
	option(initialServer)

	// Assert that the profileService field has been correctly assigned
	require.Equal(t, mockProfileService, initialServer.profileService)
}

// handles nil profileService gracefully
func TestWithProfileService_HandlesNilProfileServiceGracefully(t *testing.T) {
	// Create a new ScanServiceServer instance with initial values
	initialServer := &ScanServiceServer{
		logger:    &slog.Logger{},
		config:    &sophrosyne.Config{},
		validator: sophrosyne2.NewMockValidator(t),
	}

	// Call the WithProfileService function with nil profileService
	option := WithProfileService(nil)
	option(initialServer)

	// Assert that the profileService field is nil
	require.Nil(t, initialServer.profileService)
}

// Does nothing if provided an unsupported type
func TestWithProfileService_DoesNothingIfUnsupportedType(t *testing.T) {

	// Call the WithProfileService function with unsupported type
	option := WithProfileService(sophrosyne2.NewMockProfileService(t))
	option(nil)

	// Does not panic
}
