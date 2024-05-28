package services

import (
	"context"
	"testing"

	"github.com/madsrc/sophrosyne"
	"github.com/madsrc/sophrosyne/internal/logger"
	sophrosyne2 "github.com/madsrc/sophrosyne/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Profile lookup when a specific profile is provided in the request parameters
func TestScanService_lookupProfile_withExistingProfile(t *testing.T) {
	ctx := context.Background()
	logger, _ := logger.NewTestLogger(nil)

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
