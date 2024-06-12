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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/madsrc/sophrosyne"
	v0 "github.com/madsrc/sophrosyne/internal/grpc/sophrosyne/v0"
	"github.com/madsrc/sophrosyne/internal/log"
	sophrosyne2 "github.com/madsrc/sophrosyne/internal/mocks"
	"github.com/madsrc/sophrosyne/internal/validator"
)

// Returns profile from request when profile name is provided.
func Test_LookupProfile_ReturnsProfileFromRequestWhenProfileNameIsProvided(t *testing.T) {
	logger, _ := log.NewTestLogger(nil)
	mockProfileService := sophrosyne2.NewMockProfileService(t)
	// Create a new instance of ScanServiceServer
	server := &ScanServiceServer{
		Logger:         logger,
		Config:         &sophrosyne.Config{},
		Validator:      validator.NewValidator(),
		ProfileService: mockProfileService,
	}

	// Create a context
	ctx := context.Background()

	// Create a mock user with a default profile
	user := &sophrosyne.User{
		DefaultProfile: sophrosyne.Profile{Name: "default"},
	}

	mockProfileService.On("GetProfileByName", ctx, "test_profile").Return(sophrosyne.Profile{Name: "test_profile"}, nil)

	// Create a mock ScanRequest with a profile name
	req := &v0.ScanRequest{Profile: "test_profile"}

	// Call the lookupProfile method
	profile, err := server.lookupProfile(ctx, req, user)

	// Assert that the profile returned is from the request
	require.NoError(t, err)
	require.NotNil(t, profile)
	require.Equal(t, "test_profile", profile.Name)
}

// Returns service-wide default profile when user has no default profile.
func Test_LookupProfile_ReturnsServiceWideDefaultProfileWhenUserHasNoDefaultProfile(t *testing.T) {
	logger, _ := log.NewTestLogger(nil)
	mockProfileService := sophrosyne2.NewMockProfileService(t)
	// Create a new instance of ScanServiceServer
	server := &ScanServiceServer{
		Logger:         logger,
		Config:         &sophrosyne.Config{},
		Validator:      &validator.Validator{},
		ProfileService: mockProfileService,
	}

	// Create a context
	ctx := context.Background()

	// Create a mock user with no default profile
	curUser := &sophrosyne.User{}

	// Create a mock request with empty profile
	req := &v0.ScanRequest{}

	mockProfileService.On("GetProfileByName", ctx, "default").Return(sophrosyne.Profile{Name: "default"}, nil)

	// Call the lookupProfile method
	profile, err := server.lookupProfile(ctx, req, curUser)

	// Assert that the profile returned is the service-wide default profile
	require.NoError(t, err)
	require.NotNil(t, profile)
	require.Equal(t, "default", profile.Name)
}

// Handles nil request gracefully.
func Test_LookupProfile_HandlesNilRequestGracefully(t *testing.T) {
	logger, _ := log.NewTestLogger(nil)
	// Create a new instance of ScanServiceServer
	server := &ScanServiceServer{
		Logger:         logger,
		Config:         &sophrosyne.Config{},
		Validator:      validator.NewValidator(),
		ProfileService: sophrosyne2.NewMockProfileService(t),
	}

	// Call lookupProfile with nil request and a valid user
	profile, err := server.lookupProfile(context.Background(), nil, &sophrosyne.User{})

	// Assert that the profile is nil and error is not nil
	require.Nil(t, profile)
	require.Error(t, err)
}

// Handles nil current user gracefully.
func Test_LookupProfile_HandlesNilCurrentUserGracefully(t *testing.T) {
	logger, _ := log.NewTestLogger(nil)
	// Create a new instance of ScanServiceServer
	server := &ScanServiceServer{
		Logger:         logger,
		Config:         &sophrosyne.Config{},
		Validator:      &validator.Validator{},
		ProfileService: sophrosyne2.NewMockProfileService(t),
	}

	// Call the lookupProfile method with nil current user
	profile, err := server.lookupProfile(context.Background(), &v0.ScanRequest{Profile: "test_profile"}, nil)

	// Assert that the error is not nil and contains the expected message
	require.Error(t, err)
	require.Contains(t, err.Error(), "curUser cannot be nil")

	// Assert that the profile is nil
	require.Nil(t, profile)
}

// Handles error when profile name from request does not exist.
func Test_HandleErrorWhenProfileNameDoesNotExist(t *testing.T) {
	logger, _ := log.NewTestLogger(nil)
	mockProfileService := sophrosyne2.NewMockProfileService(t)
	// Create a new instance of ScanServiceServer
	server := &ScanServiceServer{
		Logger:         logger,
		ProfileService: mockProfileService,
	}

	// Create a mock context and user
	ctx := context.Background()
	user := &sophrosyne.User{
		DefaultProfile: sophrosyne.Profile{Name: "default"},
	}

	// Create a mock profile service that returns an error
	mockProfileService.On("GetProfileByName", ctx, "non_existing_profile").Return(sophrosyne.Profile{}, assert.AnError)

	// Create a request with a non-existing profile name
	req := &v0.ScanRequest{Profile: "non_existing_profile"}

	// Call the lookupProfile method
	_, err := server.lookupProfile(ctx, req, user)

	// Assert that an error is returned
	require.Error(t, err)
}

// Handles empty profile name in request.
func Test_LookupProfile_HandlesEmptyProfileNameInRequest(t *testing.T) {
	logger, _ := log.NewTestLogger(nil)
	mockProfileService := sophrosyne2.NewMockProfileService(t)
	// Create a new instance of ScanServiceServer
	server := &ScanServiceServer{
		Logger:         logger,
		Config:         &sophrosyne.Config{},
		Validator:      validator.NewValidator(),
		ProfileService: mockProfileService,
	}

	// Create a context
	ctx := context.Background()

	// Create a new ScanRequest with empty profile name
	req := &v0.ScanRequest{
		Profile: "",
	}

	// Create a new User with default profile name as empty
	curUser := &sophrosyne.User{
		DefaultProfile: sophrosyne.Profile{
			Name: "the default profile",
		},
	}

	// Call the lookupProfile method
	profile, err := server.lookupProfile(ctx, req, curUser)

	// Assert that the profile is not nil
	require.NotNil(t, profile)
	// Assert that the profile is the default profile of the current user
	require.Equal(t, curUser.DefaultProfile, *profile)
	// Assert that there is no error returned
	require.NoError(t, err)
}

// Error getting service wide default profile from database.
func Test_LookupProfile_HandleErrorGettingServiceWideDefaultProfile(t *testing.T) {
	logger, _ := log.NewTestLogger(nil)
	mockProfileService := sophrosyne2.NewMockProfileService(t)
	// Create a new instance of ScanServiceServer
	server := &ScanServiceServer{
		Logger:         logger,
		Config:         &sophrosyne.Config{},
		Validator:      &validator.Validator{},
		ProfileService: mockProfileService,
	}

	// Create a context
	ctx := context.Background()

	// Create a mock user with no default profile
	curUser := &sophrosyne.User{}

	// Create a mock request with empty profile
	req := &v0.ScanRequest{}

	mockProfileService.On("GetProfileByName", ctx, "default").Return(sophrosyne.Profile{}, assert.AnError)

	// Call the lookupProfile method
	profile, err := server.lookupProfile(ctx, req, curUser)

	// Assert that the profile returned is the service-wide default profile
	require.Error(t, err)
	require.Nil(t, profile)
}
