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
