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

package services

import (
	"context"
	"testing"

	"github.com/madsrc/sophrosyne"
	"github.com/madsrc/sophrosyne/internal/log"
	sophrosyne2 "github.com/madsrc/sophrosyne/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Profile lookup when a specific profile is provided in the request parameters
func TestScanService_lookupProfile_withExistingProfile(t *testing.T) {
	ctx := context.Background()
	logger, _ := log.NewTestLogger(nil)

	expectedProfile := sophrosyne.Profile{Name: "testProfile"}
	params := sophrosyne.PerformScanRequest{Profile: "testProfile"}
	curUser := &sophrosyne.User{}

	mockProfileService := sophrosyne2.NewMockProfileService(t)
	mockProfileService.On("GetProfileByName", ctx, "testProfile").Return(expectedProfile, nil)

	scanService := ScanService{
		profileService: mockProfileService,
		logger:         logger,
	}

	profile, err := scanService.lookupProfile(ctx, params, curUser)

	require.NoError(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, expectedProfile.Name, profile.Name)
	mockProfileService.AssertExpectations(t)
}
