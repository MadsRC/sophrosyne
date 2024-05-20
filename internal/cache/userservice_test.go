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

var testUser = sophrosyne.User{
	ID:   "123",
	Name: "I am the test user - if you see me, something is probably wrong!",
}

func TestNewUserServiceCache(t *testing.T) {
	psc := NewUserServiceCache(
		&sophrosyne.Config{}, nil, nil)
	assert.NotNil(t, psc)
}

func TestUserServiceCache_GetUser(t *testing.T) {
	t.Run("retrieved from cache", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		userServiceCache := getUserServiceCache(t, cts)
		expectedCheck := testUser
		userServiceCache.cache.Set(expectedCheck.ID, expectedCheck)

		result, err := userServiceCache.GetUser(cts.ctx, expectedCheck.ID)

		require.NoError(t, err)
		require.Equal(t, expectedCheck, result)
		cts.userService.AssertNotCalled(t, "GetUser", mock.Anything, mock.Anything)
	})
	t.Run("retrieved from service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		userServiceCache := getUserServiceCache(t, cts)
		expectedCheck := testUser

		cts.userService.On("GetUser", cts.ctx, expectedCheck.ID).Once().Return(expectedCheck, nil)

		result, err := userServiceCache.GetUser(cts.ctx, expectedCheck.ID)

		require.NoError(t, err)
		require.Equal(t, expectedCheck, result)

		t.Run("result was saved in cache", func(t *testing.T) {
			cacheResult, ok := userServiceCache.cache.Get(expectedCheck.ID)
			require.True(t, ok)
			require.Equal(t, expectedCheck, cacheResult)
		})
	})
	t.Run("error retrieving from service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		userServiceCache := getUserServiceCache(t, cts)
		expectedCheck := sophrosyne.User{}

		cts.userService.On("GetUser", cts.ctx, testUser.ID).Once().Return(expectedCheck, assert.AnError)

		got, err := userServiceCache.GetUser(cts.ctx, testUser.ID)

		require.Error(t, err)
		require.ErrorIs(t, err, assert.AnError)
		require.Equal(t, expectedCheck, got)
	})
}
