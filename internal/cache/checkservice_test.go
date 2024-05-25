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

var testCheck = sophrosyne.Check{
	ID:   "123",
	Name: "I am the test check - if you see me, something is probably wrong!",
}
var secondTestCheck = sophrosyne.Check{
	ID:   "456",
	Name: "I am the second test check - if you see me, something is probably wrong!",
}

func TestNewCheckServiceCache(t *testing.T) {
	psc := NewCheckServiceCache(
		&sophrosyne.Config{}, nil, nil)
	assert.NotNil(t, psc)
}

func TestCheckServiceCache_GetCheck(t *testing.T) {
	t.Run("retrieved from cache", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		checkServiceCache := getCheckServiceCache(t, cts)
		expectedCheck := testCheck
		checkServiceCache.cache.Set(expectedCheck.ID, expectedCheck)

		result, err := checkServiceCache.GetCheck(cts.ctx, expectedCheck.ID)

		require.NoError(t, err)
		require.Equal(t, expectedCheck, result)
		cts.checkService.AssertNotCalled(t, "GetCheck", mock.Anything, mock.Anything)
	})
	t.Run("retrieved from service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		checkServiceCache := getCheckServiceCache(t, cts)
		expectedCheck := testCheck

		cts.checkService.On("GetCheck", cts.ctx, expectedCheck.ID).Once().Return(expectedCheck, nil)

		result, err := checkServiceCache.GetCheck(cts.ctx, expectedCheck.ID)

		require.NoError(t, err)
		require.Equal(t, expectedCheck, result)

		t.Run("result was saved in cache", func(t *testing.T) {
			cacheResult, ok := checkServiceCache.cache.Get(expectedCheck.ID)
			require.True(t, ok)
			require.Equal(t, expectedCheck, cacheResult)
		})
	})
	t.Run("error retrieving from service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		checkServiceCache := getCheckServiceCache(t, cts)
		expectedCheck := sophrosyne.Check{}

		cts.checkService.On("GetCheck", cts.ctx, testCheck.ID).Once().Return(expectedCheck, assert.AnError)

		got, err := checkServiceCache.GetCheck(cts.ctx, testCheck.ID)

		require.Error(t, err)
		require.ErrorIs(t, err, assert.AnError)
		require.Equal(t, expectedCheck, got)
	})
}

func TestCheckServiceCache_GetCheckByName(t *testing.T) {
	t.Run("retrieved from cache", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		checkServiceCache := getCheckServiceCache(t, cts)
		expectedCheck := testCheck
		checkServiceCache.nameToIDCache.Set(expectedCheck.Name, expectedCheck.ID)
		checkServiceCache.cache.Set(expectedCheck.ID, expectedCheck)

		cts.tracingService.On("StartSpan", cts.ctx, mock.Anything).Once().Return(cts.ctx, cts.span)
		cts.span.On("End").Once().Return(nil)

		result, err := checkServiceCache.GetCheckByName(cts.ctx, expectedCheck.Name)

		require.NoError(t, err)
		require.Equal(t, expectedCheck, result)
		cts.checkService.AssertNotCalled(t, "GetCheckByName", mock.Anything, mock.Anything)
	})
	t.Run("retrieved from service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		checkServiceCache := getCheckServiceCache(t, cts)
		expectedCheck := testCheck

		cts.checkService.On("GetCheckByName", cts.ctx, expectedCheck.Name).Once().Return(expectedCheck, nil)

		result, err := checkServiceCache.GetCheckByName(cts.ctx, expectedCheck.Name)

		require.NoError(t, err)
		require.Equal(t, expectedCheck, result)
	})

	t.Run("error retrieving from service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		checkServiceCache := getCheckServiceCache(t, cts)
		expectedCheck := sophrosyne.Check{}

		cts.checkService.On("GetCheckByName", cts.ctx, testCheck.Name).Once().Return(expectedCheck, assert.AnError)

		result, err := checkServiceCache.GetCheckByName(cts.ctx, testCheck.Name)

		require.Error(t, err)
		require.ErrorIs(t, err, assert.AnError)
		require.Equal(t, expectedCheck, result)
	})
}

func TestCheckServiceCache_GetChecks(t *testing.T) {
	t.Run("retrieved from service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		checkServiceCache := getCheckServiceCache(t, cts)
		expectedChecks := []sophrosyne.Check{testCheck, secondTestCheck}

		cts.checkService.On("GetChecks", cts.ctx, mock.Anything).Once().Return(expectedChecks, nil)

		result, err := checkServiceCache.GetChecks(cts.ctx, nil)

		require.NoError(t, err)
		require.Equal(t, expectedChecks, result)
		cacheEntryOne, ok := checkServiceCache.cache.Get(expectedChecks[0].ID)
		require.True(t, ok)
		require.Equal(t, expectedChecks[0], cacheEntryOne)
		cacheEntryTwo, ok := checkServiceCache.cache.Get(expectedChecks[1].ID)
		require.True(t, ok)
		require.Equal(t, expectedChecks[1], cacheEntryTwo)
	})
	t.Run("error retrieving", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		checkServiceCache := getCheckServiceCache(t, cts)

		cts.checkService.On("GetChecks", cts.ctx, mock.Anything).Once().Return(nil, assert.AnError)

		result, err := checkServiceCache.GetChecks(cts.ctx, nil)
		require.Nil(t, result)
		require.ErrorIs(t, err, assert.AnError)
	})
}

func TestCheckServiceCache_CreateCheck(t *testing.T) {
	t.Run("created in service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		checkServiceCache := getCheckServiceCache(t, cts)
		expectedCheck := testCheck
		input := sophrosyne.CreateCheckRequest{
			Name: expectedCheck.Name,
		}

		cts.checkService.On("CreateCheck", cts.ctx, mock.Anything).Once().Return(expectedCheck, nil)

		result, err := checkServiceCache.CreateCheck(cts.ctx, input)

		require.NoError(t, err)
		require.Equal(t, expectedCheck, result)
		cacheEntry, ok := checkServiceCache.cache.Get(expectedCheck.ID)
		require.True(t, ok)
		require.Equal(t, expectedCheck, cacheEntry)

	})
	t.Run("error creating", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		checkServiceCache := getCheckServiceCache(t, cts)
		input := sophrosyne.CreateCheckRequest{
			Name: testCheck.Name,
		}

		cts.checkService.On("CreateCheck", cts.ctx, mock.Anything).Once().Return(sophrosyne.Check{}, assert.AnError)

		result, err := checkServiceCache.CreateCheck(cts.ctx, input)

		require.Equal(t, sophrosyne.Check{}, result)
		require.ErrorIs(t, err, assert.AnError)
	})
}

func TestCheckServiceCache_UpdateCheck(t *testing.T) {
	t.Run("updated in service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		checkServiceCache := getCheckServiceCache(t, cts)
		expectedCheck := testCheck
		input := sophrosyne.UpdateCheckRequest{
			Name: expectedCheck.Name,
		}

		cts.checkService.On("UpdateCheck", cts.ctx, mock.Anything).Once().Return(expectedCheck, nil)

		result, err := checkServiceCache.UpdateCheck(cts.ctx, input)

		require.NoError(t, err)
		require.Equal(t, expectedCheck, result)
		cacheEntry, ok := checkServiceCache.cache.Get(expectedCheck.ID)
		require.True(t, ok)
		require.Equal(t, expectedCheck, cacheEntry)

	})
	t.Run("error updating", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		checkServiceCache := getCheckServiceCache(t, cts)
		input := sophrosyne.UpdateCheckRequest{
			Name: testCheck.Name,
		}

		cts.checkService.On("UpdateCheck", cts.ctx, mock.Anything).Once().Return(sophrosyne.Check{}, assert.AnError)

		result, err := checkServiceCache.UpdateCheck(cts.ctx, input)

		require.Equal(t, sophrosyne.Check{}, result)
		require.ErrorIs(t, err, assert.AnError)
	})
}

func TestCheckServiceCache_DeleteCheck(t *testing.T) {
	t.Run("deleted in service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		checkServiceCache := getCheckServiceCache(t, cts)
		expectedCheck := testCheck
		input := expectedCheck.ID
		checkServiceCache.cache.Set(expectedCheck.ID, expectedCheck)
		checkServiceCache.nameToIDCache.Set(expectedCheck.Name, expectedCheck.ID)

		cts.checkService.On("GetCheck", cts.ctx, input).Once().Return(expectedCheck, nil)
		cts.checkService.On("DeleteCheck", cts.ctx, mock.Anything).Once().Return(nil)

		err := checkServiceCache.DeleteCheck(cts.ctx, input)

		require.NoError(t, err)
		cacheEntry, ok := checkServiceCache.cache.Get(expectedCheck.ID)
		require.False(t, ok)
		require.NotEqual(t, expectedCheck, cacheEntry)

	})
	t.Run("error getting check", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		checkServiceCache := getCheckServiceCache(t, cts)
		expectedCheck := testCheck
		input := expectedCheck.ID
		checkServiceCache.cache.Set(expectedCheck.ID, expectedCheck)
		checkServiceCache.nameToIDCache.Set(expectedCheck.Name, expectedCheck.ID)

		cts.checkService.On("GetCheck", cts.ctx, input).Once().Return(expectedCheck, assert.AnError)

		err := checkServiceCache.DeleteCheck(cts.ctx, input)

		require.ErrorIs(t, err, assert.AnError)
		cacheEntry, ok := checkServiceCache.cache.Get(expectedCheck.ID)
		require.True(t, ok)
		require.Equal(t, expectedCheck, cacheEntry)
	})
	t.Run("error deleting", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		checkServiceCache := getCheckServiceCache(t, cts)
		input := testCheck.ID

		cts.checkService.On("GetCheck", cts.ctx, input).Once().Return(testCheck, nil)
		cts.checkService.On("DeleteCheck", cts.ctx, mock.Anything).Once().Return(assert.AnError)

		err := checkServiceCache.DeleteCheck(cts.ctx, input)

		require.ErrorIs(t, err, assert.AnError)
	})
}
