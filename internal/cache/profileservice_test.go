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
