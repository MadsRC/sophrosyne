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

// Successfully extracts user from context and performs scan
func TestScanServiceServer_SuccessfullyExtractsUserAndPerformsScan(t *testing.T) {
	ctx := context.WithValue(context.Background(), sophrosyne.UserContextKey{}, &sophrosyne.User{
		DefaultProfile: sophrosyne.Profile{Name: "default"},
	})

	request := &v0.ScanRequest{Profile: "testProfile"}

	logger, _ := log.NewTestLogger(nil)

	mockProfileService := sophrosyne2.NewMockProfileService(t)
	mockProfileService.On("GetProfileByName", ctx, "testProfile").Return(sophrosyne.Profile{Name: "testProfile"}, nil)

	server := ScanServiceServer{
		logger:         logger,
		config:         &sophrosyne.Config{},
		validator:      &validator.Validator{},
		profileService: mockProfileService,
	}

	response, err := server.Scan(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, response)
}

// User extraction from context fails
func TestScanServiceServer_UserExtractionFails(t *testing.T) {
	ctx := context.Background()

	request := &v0.ScanRequest{Profile: "testProfile"}

	logger, _ := log.NewTestLogger(nil)

	server := ScanServiceServer{
		logger:         logger,
		config:         &sophrosyne.Config{},
		validator:      &validator.Validator{},
		profileService: sophrosyne2.NewMockProfileService(t),
	}

	response, err := server.Scan(ctx, request)

	require.Error(t, err)
	require.Nil(t, response)
}

// Profile lookup fails
func TestScanServiceServer_ProfileLookupFails(t *testing.T) {
	ctx := context.WithValue(context.Background(), sophrosyne.UserContextKey{}, &sophrosyne.User{
		DefaultProfile: sophrosyne.Profile{Name: "default"},
	})

	request := &v0.ScanRequest{Profile: "testProfile"}

	logger, _ := log.NewTestLogger(nil)

	mockProfileService := sophrosyne2.NewMockProfileService(t)
	mockProfileService.On("GetProfileByName", ctx, "testProfile").Return(sophrosyne.Profile{}, assert.AnError)

	server := ScanServiceServer{
		logger:         logger,
		config:         &sophrosyne.Config{},
		validator:      &validator.Validator{},
		profileService: mockProfileService,
	}

	response, err := server.Scan(ctx, request)

	require.Error(t, err)
	require.Nil(t, response)
}
