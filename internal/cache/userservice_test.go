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
var secondTestUser = sophrosyne.User{
	ID:   "456",
	Name: "I am the second test user - if you see me, something is probably wrong!",
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
		expectedUser := testUser
		userServiceCache.cache.Set(expectedUser.ID, expectedUser)

		result, err := userServiceCache.GetUser(cts.ctx, expectedUser.ID)

		require.NoError(t, err)
		require.Equal(t, expectedUser, result)
		cts.userService.AssertNotCalled(t, "GetUser", mock.Anything, mock.Anything)
	})
	t.Run("retrieved from service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		userServiceCache := getUserServiceCache(t, cts)
		expectedUser := testUser

		cts.userService.On("GetUser", cts.ctx, expectedUser.ID).Once().Return(expectedUser, nil)

		result, err := userServiceCache.GetUser(cts.ctx, expectedUser.ID)

		require.NoError(t, err)
		require.Equal(t, expectedUser, result)

		t.Run("result was saved in cache", func(t *testing.T) {
			cacheResult, ok := userServiceCache.cache.Get(expectedUser.ID)
			require.True(t, ok)
			require.Equal(t, expectedUser, cacheResult)
		})
	})
	t.Run("error retrieving from service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		userServiceCache := getUserServiceCache(t, cts)
		expectedUser := sophrosyne.User{}

		cts.userService.On("GetUser", cts.ctx, testUser.ID).Once().Return(expectedUser, assert.AnError)

		got, err := userServiceCache.GetUser(cts.ctx, testUser.ID)

		require.Error(t, err)
		require.ErrorIs(t, err, assert.AnError)
		require.Equal(t, expectedUser, got)
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

func TestUserServiceCache_GetUsers(t *testing.T) {
	t.Run("retrieved from service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		userServiceCache := getUserServiceCache(t, cts)
		expectedUsers := []sophrosyne.User{testUser, secondTestUser}

		cts.userService.On("GetUsers", cts.ctx, mock.Anything).Once().Return(expectedUsers, nil)

		result, err := userServiceCache.GetUsers(cts.ctx, nil)

		require.NoError(t, err)
		require.Equal(t, expectedUsers, result)
		cacheEntryOne, ok := userServiceCache.cache.Get(expectedUsers[0].ID)
		require.True(t, ok)
		require.Equal(t, expectedUsers[0], cacheEntryOne)
		cacheEntryTwo, ok := userServiceCache.cache.Get(expectedUsers[1].ID)
		require.True(t, ok)
		require.Equal(t, expectedUsers[1], cacheEntryTwo)
	})
	t.Run("error retrieving", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		userServiceCache := getUserServiceCache(t, cts)

		cts.userService.On("GetUsers", cts.ctx, mock.Anything).Once().Return(nil, assert.AnError)

		result, err := userServiceCache.GetUsers(cts.ctx, nil)
		require.Nil(t, result)
		require.ErrorIs(t, err, assert.AnError)
	})
}

func TestUserServiceCache_CreateUser(t *testing.T) {
	t.Run("created in service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		userServiceCache := getUserServiceCache(t, cts)
		expectedUser := testUser
		input := sophrosyne.CreateUserRequest{
			Name: expectedUser.Name,
		}

		cts.userService.On("CreateUser", cts.ctx, mock.Anything).Once().Return(expectedUser, nil)

		result, err := userServiceCache.CreateUser(cts.ctx, input)

		require.NoError(t, err)
		require.Equal(t, expectedUser, result)
		cacheEntry, ok := userServiceCache.cache.Get(expectedUser.ID)
		require.True(t, ok)
		require.Equal(t, expectedUser, cacheEntry)

	})
	t.Run("error creating", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		userServiceCache := getUserServiceCache(t, cts)
		input := sophrosyne.CreateUserRequest{
			Name: testUser.Name,
		}

		cts.userService.On("CreateUser", cts.ctx, mock.Anything).Once().Return(sophrosyne.User{}, assert.AnError)

		result, err := userServiceCache.CreateUser(cts.ctx, input)

		require.Equal(t, sophrosyne.User{}, result)
		require.ErrorIs(t, err, assert.AnError)
	})
}

func TestUserServiceCache_UpdateUser(t *testing.T) {
	t.Run("updated in service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		userServiceCache := getUserServiceCache(t, cts)
		expectedUser := testUser
		input := sophrosyne.UpdateUserRequest{
			Name: expectedUser.Name,
		}

		cts.userService.On("UpdateUser", cts.ctx, mock.Anything).Once().Return(expectedUser, nil)

		result, err := userServiceCache.UpdateUser(cts.ctx, input)

		require.NoError(t, err)
		require.Equal(t, expectedUser, result)
		cacheEntry, ok := userServiceCache.cache.Get(expectedUser.ID)
		require.True(t, ok)
		require.Equal(t, expectedUser, cacheEntry)

	})
	t.Run("error updating", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		userServiceCache := getUserServiceCache(t, cts)
		input := sophrosyne.UpdateUserRequest{
			Name: testUser.Name,
		}

		cts.userService.On("UpdateUser", cts.ctx, mock.Anything).Once().Return(sophrosyne.User{}, assert.AnError)

		result, err := userServiceCache.UpdateUser(cts.ctx, input)

		require.Equal(t, sophrosyne.User{}, result)
		require.ErrorIs(t, err, assert.AnError)
	})
}

func TestUserServiceCache_DeleteUser(t *testing.T) {
	t.Run("deleted in service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		userServiceCache := getUserServiceCache(t, cts)
		expectedUser := testUser
		input := expectedUser.ID
		userServiceCache.cache.Set(expectedUser.ID, expectedUser)
		userServiceCache.nameToIDCache.Set(expectedUser.Name, expectedUser.ID)

		cts.userService.On("GetUser", cts.ctx, input).Once().Return(expectedUser, nil)
		cts.userService.On("DeleteUser", cts.ctx, mock.Anything).Once().Return(nil)

		err := userServiceCache.DeleteUser(cts.ctx, input)

		require.NoError(t, err)
		cacheEntry, ok := userServiceCache.cache.Get(expectedUser.ID)
		require.False(t, ok)
		require.NotEqual(t, expectedUser, cacheEntry)

	})
	t.Run("error getting user", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		userServiceCache := getUserServiceCache(t, cts)
		expectedUser := testUser
		input := expectedUser.Name
		userServiceCache.cache.Set(expectedUser.ID, expectedUser)
		userServiceCache.nameToIDCache.Set(expectedUser.Name, expectedUser.ID)

		cts.userService.On("GetUser", cts.ctx, input).Once().Return(expectedUser, assert.AnError)

		err := userServiceCache.DeleteUser(cts.ctx, input)

		require.ErrorIs(t, err, assert.AnError)
		cacheEntry, ok := userServiceCache.cache.Get(expectedUser.ID)
		require.True(t, ok)
		require.Equal(t, expectedUser, cacheEntry)
	})
	t.Run("error deleting", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		userServiceCache := getUserServiceCache(t, cts)
		input := testUser.ID

		cts.userService.On("GetUser", cts.ctx, input).Once().Return(testUser, nil)
		cts.userService.On("DeleteUser", cts.ctx, mock.Anything).Once().Return(assert.AnError)

		err := userServiceCache.DeleteUser(cts.ctx, input)

		require.ErrorIs(t, err, assert.AnError)
	})
}

func TestUserServiceCache_RotateToken(t *testing.T) {
	t.Run("rotated in service", func(t *testing.T) {
		cts := setupTestStuff(t, nil)
		userServiceCache := getUserServiceCache(t, cts)
		expectedUser := testUser
		input := expectedUser.ID

		cts.userService.On("RotateToken", cts.ctx, input).Once().Return([]byte("token"), nil)

		result, err := userServiceCache.RotateToken(cts.ctx, input)

		require.NoError(t, err)
		require.Equal(t, []byte("token"), result)
	})
}

func TestUserServiceCache_Health(t *testing.T) {
	cts := setupTestStuff(t, nil)
	userServiceCache := getUserServiceCache(t, cts)
	ok, result := userServiceCache.Health(cts.ctx)

	require.True(t, ok)
	require.Equal(t, []byte(`{"ok"}`), result)
}
