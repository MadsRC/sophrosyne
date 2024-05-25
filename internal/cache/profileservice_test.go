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

package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/madsrc/sophrosyne"
)

var testProfile = sophrosyne.Profile{
	ID:   "123",
	Name: "I am the test profile - if you see me, something is probably wrong!",
}
var secondTestProfile = sophrosyne.Profile{
	ID:   "456",
	Name: "I am the second test profile - if you see me, something is probably wrong!",
}

func TestNewProfileServiceCache(t *testing.T) {
	psc := NewProfileServiceCache(
		&sophrosyne.Config{}, nil, nil)
	assert.NotNil(t, psc)
}

func TestProfileServiceCache_GetProfile(t *testing.T) {
	t.Run("retrieved from cache", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		profileServiceCache := getProfileServiceCache(t, cts)
		expectedProfile := testProfile
		profileServiceCache.cache.Set(expectedProfile.ID, expectedProfile)

		result, err := profileServiceCache.GetProfile(cts.ctx, expectedProfile.ID)

		require.NoError(t, err)
		require.Equal(t, expectedProfile, result)
		cts.profileService.AssertNotCalled(t, "GetProfile", mock.Anything, mock.Anything)
	})

	t.Run("retrieved from service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		profileServiceCache := getProfileServiceCache(t, cts)
		expectedProfile := testProfile

		cts.profileService.On("GetProfile", cts.ctx, expectedProfile.ID).Once().Return(expectedProfile, nil)

		result, err := profileServiceCache.GetProfile(cts.ctx, expectedProfile.ID)

		require.NoError(t, err)
		require.Equal(t, expectedProfile, result)

		t.Run("result was saved in cache", func(t *testing.T) {
			cacheResult, ok := profileServiceCache.cache.Get(expectedProfile.ID)
			require.True(t, ok)
			require.Equal(t, expectedProfile, cacheResult)
		})
	})

	t.Run("error retrieving from service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		profileServiceCache := getProfileServiceCache(t, cts)
		expectedProfile := sophrosyne.Profile{}

		cts.profileService.On("GetProfile", cts.ctx, testProfile.ID).Once().Return(expectedProfile, assert.AnError)

		result, err := profileServiceCache.GetProfile(cts.ctx, testProfile.ID)

		require.Error(t, err)
		require.ErrorIs(t, err, assert.AnError)
		require.Equal(t, expectedProfile, result)
	})
}

func TestProfileServiceCache_GetProfileByName(t *testing.T) {
	t.Run("retrieved from cache", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		profileServiceCache := getProfileServiceCache(t, cts)
		expectedProfile := testProfile
		profileServiceCache.nameToIDCache.Set(expectedProfile.Name, expectedProfile.ID)
		profileServiceCache.cache.Set(expectedProfile.ID, expectedProfile)

		cts.tracingService.On("StartSpan", cts.ctx, mock.Anything).Once().Return(cts.ctx, cts.span)
		cts.span.On("End").Once().Return(nil)

		result, err := profileServiceCache.GetProfileByName(cts.ctx, expectedProfile.Name)

		require.NoError(t, err)
		require.Equal(t, expectedProfile, result)
		cts.profileService.AssertNotCalled(t, "GetProfileByName", mock.Anything, mock.Anything)
	})
	t.Run("retrieved from service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		profileServiceCache := getProfileServiceCache(t, cts)
		expectedProfile := testProfile

		cts.profileService.On("GetProfileByName", cts.ctx, expectedProfile.Name).Once().Return(expectedProfile, nil)

		result, err := profileServiceCache.GetProfileByName(cts.ctx, expectedProfile.Name)

		require.NoError(t, err)
		require.Equal(t, expectedProfile, result)
	})

	t.Run("error retrieving from service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		profileServiceCache := getProfileServiceCache(t, cts)
		expectedProfile := sophrosyne.Profile{}

		cts.profileService.On("GetProfileByName", cts.ctx, testProfile.Name).Once().Return(expectedProfile, assert.AnError)

		result, err := profileServiceCache.GetProfileByName(cts.ctx, testProfile.Name)

		require.Error(t, err)
		require.ErrorIs(t, err, assert.AnError)
		require.Equal(t, expectedProfile, result)
	})
}

func TestProfileServiceCache_GetProfiles(t *testing.T) {
	t.Run("retrieved from service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		profileServiceCache := getProfileServiceCache(t, cts)
		expectedProfiles := []sophrosyne.Profile{testProfile, secondTestProfile}

		cts.profileService.On("GetProfiles", cts.ctx, mock.Anything).Once().Return(expectedProfiles, nil)

		result, err := profileServiceCache.GetProfiles(cts.ctx, nil)

		require.NoError(t, err)
		require.Equal(t, expectedProfiles, result)
		cacheEntryOne, ok := profileServiceCache.cache.Get(expectedProfiles[0].ID)
		require.True(t, ok)
		require.Equal(t, expectedProfiles[0], cacheEntryOne)
		cacheEntryTwo, ok := profileServiceCache.cache.Get(expectedProfiles[1].ID)
		require.True(t, ok)
		require.Equal(t, expectedProfiles[1], cacheEntryTwo)
	})
	t.Run("error retrieving", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		profileServiceCache := getProfileServiceCache(t, cts)

		cts.profileService.On("GetProfiles", cts.ctx, mock.Anything).Once().Return(nil, assert.AnError)

		result, err := profileServiceCache.GetProfiles(cts.ctx, nil)
		require.Nil(t, result)
		require.ErrorIs(t, err, assert.AnError)
	})
}

func TestProfileServiceCache_CreateProfile(t *testing.T) {
	t.Run("created in service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		profileServiceCache := getProfileServiceCache(t, cts)
		expectedProfile := testProfile
		input := sophrosyne.CreateProfileRequest{
			Name: expectedProfile.Name,
		}

		cts.profileService.On("CreateProfile", cts.ctx, mock.Anything).Once().Return(expectedProfile, nil)

		result, err := profileServiceCache.CreateProfile(cts.ctx, input)

		require.NoError(t, err)
		require.Equal(t, expectedProfile, result)
		cacheEntry, ok := profileServiceCache.cache.Get(expectedProfile.ID)
		require.True(t, ok)
		require.Equal(t, expectedProfile, cacheEntry)

	})
	t.Run("error creating", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		profileServiceCache := getProfileServiceCache(t, cts)
		input := sophrosyne.CreateProfileRequest{
			Name: testProfile.Name,
		}

		cts.profileService.On("CreateProfile", cts.ctx, mock.Anything).Once().Return(sophrosyne.Profile{}, assert.AnError)

		result, err := profileServiceCache.CreateProfile(cts.ctx, input)

		require.Equal(t, sophrosyne.Profile{}, result)
		require.ErrorIs(t, err, assert.AnError)
	})
}

func TestProfileServiceCache_UpdateProfile(t *testing.T) {
	t.Run("updated in service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		profileServiceCache := getProfileServiceCache(t, cts)
		expectedProfile := testProfile
		input := sophrosyne.UpdateProfileRequest{
			Name: expectedProfile.Name,
		}

		cts.profileService.On("UpdateProfile", cts.ctx, mock.Anything).Once().Return(expectedProfile, nil)

		result, err := profileServiceCache.UpdateProfile(cts.ctx, input)

		require.NoError(t, err)
		require.Equal(t, expectedProfile, result)
		cacheEntry, ok := profileServiceCache.cache.Get(expectedProfile.ID)
		require.True(t, ok)
		require.Equal(t, expectedProfile, cacheEntry)

	})
	t.Run("error updating", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		profileServiceCache := getProfileServiceCache(t, cts)
		input := sophrosyne.UpdateProfileRequest{
			Name: testProfile.Name,
		}

		cts.profileService.On("UpdateProfile", cts.ctx, mock.Anything).Once().Return(sophrosyne.Profile{}, assert.AnError)

		result, err := profileServiceCache.UpdateProfile(cts.ctx, input)

		require.Equal(t, sophrosyne.Profile{}, result)
		require.ErrorIs(t, err, assert.AnError)
	})
}

func TestProfileServiceCache_DeleteProfile(t *testing.T) {
	t.Run("deleted in service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		profileServiceCache := getProfileServiceCache(t, cts)
		expectedProfile := testProfile
		input := expectedProfile.Name
		profileServiceCache.cache.Set(expectedProfile.ID, expectedProfile)
		profileServiceCache.nameToIDCache.Set(expectedProfile.Name, expectedProfile.ID)

		cts.profileService.On("GetProfileByName", cts.ctx, input).Once().Return(expectedProfile, nil)
		cts.profileService.On("DeleteProfile", cts.ctx, mock.Anything).Once().Return(nil)

		err := profileServiceCache.DeleteProfile(cts.ctx, input)

		require.NoError(t, err)
		cacheEntry, ok := profileServiceCache.cache.Get(expectedProfile.ID)
		require.False(t, ok)
		require.NotEqual(t, expectedProfile, cacheEntry)

	})
	t.Run("error getting profile", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		profileServiceCache := getProfileServiceCache(t, cts)
		expectedProfile := testProfile
		input := expectedProfile.Name
		profileServiceCache.cache.Set(expectedProfile.ID, expectedProfile)
		profileServiceCache.nameToIDCache.Set(expectedProfile.Name, expectedProfile.ID)

		cts.profileService.On("GetProfileByName", cts.ctx, input).Once().Return(expectedProfile, assert.AnError)

		err := profileServiceCache.DeleteProfile(cts.ctx, input)

		require.ErrorIs(t, err, assert.AnError)
		cacheEntry, ok := profileServiceCache.cache.Get(expectedProfile.ID)
		require.True(t, ok)
		require.Equal(t, expectedProfile, cacheEntry)
	})
	t.Run("error deleting", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		profileServiceCache := getProfileServiceCache(t, cts)
		input := testProfile.Name

		cts.profileService.On("GetProfileByName", cts.ctx, input).Once().Return(testProfile, nil)
		cts.profileService.On("DeleteProfile", cts.ctx, mock.Anything).Once().Return(assert.AnError)

		err := profileServiceCache.DeleteProfile(cts.ctx, input)

		require.ErrorIs(t, err, assert.AnError)
	})
}
