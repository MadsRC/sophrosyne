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

func TestUserServiceCache_GetUserByName(t *testing.T) {
	t.Run("retrieved from cache", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		userServiceCache := getUserServiceCache(t, cts)
		expectedUser := testUser
		userServiceCache.nameToIDCache.Set(expectedUser.Name, expectedUser.ID)
		userServiceCache.cache.Set(expectedUser.ID, expectedUser)

		cts.tracingService.On("StartSpan", cts.ctx, mock.Anything).Once().Return(cts.ctx, cts.span)
		cts.span.On("End").Once().Return(nil)

		result, err := userServiceCache.GetUserByName(cts.ctx, expectedUser.Name)

		require.NoError(t, err)
		require.Equal(t, expectedUser, result)
		cts.userService.AssertNotCalled(t, "GetUserByName", mock.Anything, mock.Anything)
	})
	t.Run("retrieved from service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		userServiceCache := getUserServiceCache(t, cts)
		expectedUser := testUser

		cts.userService.On("GetUserByName", cts.ctx, expectedUser.Name).Once().Return(expectedUser, nil)

		result, err := userServiceCache.GetUserByName(cts.ctx, expectedUser.Name)

		require.NoError(t, err)
		require.Equal(t, expectedUser, result)
	})

	t.Run("error retrieving from service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		userServiceCache := getUserServiceCache(t, cts)
		expectedUser := sophrosyne.User{}

		cts.userService.On("GetUserByName", cts.ctx, testUser.Name).Once().Return(expectedUser, assert.AnError)

		result, err := userServiceCache.GetUserByName(cts.ctx, testUser.Name)

		require.Error(t, err)
		require.ErrorIs(t, err, assert.AnError)
		require.Equal(t, expectedUser, result)
	})
}

func TestUserServiceCache_GetUserByEmail(t *testing.T) {
	t.Run("retrieved from cache", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		userServiceCache := getUserServiceCache(t, cts)
		expectedUser := testUser
		userServiceCache.emailToIDCache.Set(expectedUser.Email, expectedUser.ID)
		userServiceCache.cache.Set(expectedUser.ID, expectedUser)

		cts.tracingService.On("StartSpan", cts.ctx, mock.Anything).Once().Return(cts.ctx, cts.span)
		cts.span.On("End").Once().Return(nil)

		result, err := userServiceCache.GetUserByEmail(cts.ctx, expectedUser.Email)

		require.NoError(t, err)
		require.Equal(t, expectedUser, result)
		cts.userService.AssertNotCalled(t, "GetUserByEmail", mock.Anything, mock.Anything)
	})
	t.Run("retrieved from service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		userServiceCache := getUserServiceCache(t, cts)
		expectedUser := testUser

		cts.userService.On("GetUserByEmail", cts.ctx, expectedUser.Email).Once().Return(expectedUser, nil)

		result, err := userServiceCache.GetUserByEmail(cts.ctx, expectedUser.Email)

		require.NoError(t, err)
		require.Equal(t, expectedUser, result)
	})

	t.Run("error retrieving from service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		userServiceCache := getUserServiceCache(t, cts)
		expectedUser := sophrosyne.User{}

		cts.userService.On("GetUserByEmail", cts.ctx, testUser.Email).Once().Return(expectedUser, assert.AnError)

		result, err := userServiceCache.GetUserByEmail(cts.ctx, testUser.Email)

		require.Error(t, err)
		require.ErrorIs(t, err, assert.AnError)
		require.Equal(t, expectedUser, result)
	})
}

func TestUserServiceCache_GetUserByToken(t *testing.T) {
	t.Run("retrieved from service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		userServiceCache := getUserServiceCache(t, cts)
		expectedUser := testUser

		cts.userService.On("GetUserByToken", cts.ctx, expectedUser.Token).Once().Return(expectedUser, nil)

		result, err := userServiceCache.GetUserByToken(cts.ctx, expectedUser.Token)

		cachedUser, ok := userServiceCache.cache.Get(expectedUser.ID)

		require.True(t, ok)
		require.Equal(t, expectedUser, cachedUser)

		require.NoError(t, err)
		require.Equal(t, expectedUser, result)
	})

	t.Run("error retrieving from service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		userServiceCache := getUserServiceCache(t, cts)
		expectedUser := sophrosyne.User{}

		cts.userService.On("GetUserByToken", cts.ctx, testUser.Token).Once().Return(expectedUser, assert.AnError)

		result, err := userServiceCache.GetUserByToken(cts.ctx, testUser.Token)

		require.Error(t, err)
		require.ErrorIs(t, err, assert.AnError)
		require.Equal(t, expectedUser, result)
	})
}
