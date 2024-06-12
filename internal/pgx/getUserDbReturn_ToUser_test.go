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

package pgx

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/madsrc/sophrosyne"
	sophrosyne2 "github.com/madsrc/sophrosyne/internal/mocks"
)

// Verify that all fields from getUserDbReturn are correctly mapped to User.
func TestGetUserDbReturn_ToUser_FieldMappingToUser(t *testing.T) {
	now := time.Now()
	deletedAt := now.Add(-24 * time.Hour)
	g := getUserDbReturn{
		ID:      "123",
		Name:    "John Doe",
		Email:   "john.doe@example.com",
		Token:   []byte("token"),
		IsAdmin: true,
		DefaultProfile: pgtype.Text{
			String: "someID",
			Valid:  true,
		},
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: &deletedAt,
	}

	expectedDefaultProfile := sophrosyne.Profile{
		ID:   "someID",
		Name: "someNameOfAProfile",
	}
	mockedProfileService := sophrosyne2.NewMockProfileService(t)
	mockedProfileService.On("GetProfile", context.Background(), g.DefaultProfile.String).Return(expectedDefaultProfile, nil)

	user, err := g.ToUser(context.Background(), mockedProfileService)

	require.NoError(t, err)
	require.Equal(t, g.ID, user.ID)
	require.Equal(t, g.Name, user.Name)
	require.Equal(t, g.Email, user.Email)
	require.Equal(t, g.Token, user.Token)
	require.Equal(t, g.IsAdmin, user.IsAdmin)
	require.Equal(t, expectedDefaultProfile, user.DefaultProfile)
	require.Equal(t, g.CreatedAt, user.CreatedAt)
	require.Equal(t, g.UpdatedAt, user.UpdatedAt)
	require.Equal(t, g.DeletedAt, user.DeletedAt)
}

// Test with a nil DeletedAt to ensure it is handled without errors.
func TestGetUserDbReturn_ToUser_NilDeletedAtHandling(t *testing.T) {
	now := time.Now()
	g := getUserDbReturn{
		ID:      "123",
		Name:    "John Doe",
		Email:   "john.doe@example.com",
		Token:   []byte("token"),
		IsAdmin: true,
		DefaultProfile: pgtype.Text{
			String: "someID",
			Valid:  true,
		},
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: nil,
	}

	expectedDefaultProfile := sophrosyne.Profile{
		ID:   "someID",
		Name: g.DefaultProfile.String,
	}
	mockedProfileService := sophrosyne2.NewMockProfileService(t)
	mockedProfileService.On("GetProfile", context.Background(), g.DefaultProfile.String).Return(expectedDefaultProfile, nil)

	user, err := g.ToUser(context.Background(), mockedProfileService)

	require.NoError(t, err)
	require.Equal(t, g.ID, user.ID)
	require.Equal(t, g.Name, user.Name)
	require.Equal(t, g.Email, user.Email)
	require.Equal(t, g.Token, user.Token)
	require.Equal(t, g.IsAdmin, user.IsAdmin)
	require.Equal(t, expectedDefaultProfile, user.DefaultProfile)
	require.Equal(t, g.CreatedAt, user.CreatedAt)
	require.Equal(t, g.UpdatedAt, user.UpdatedAt)
	require.Equal(t, g.DeletedAt, user.DeletedAt)
}

func TestGetUserDbReturn_ToUser_DefaultProfileStringEmpty_ReturnsDefaultProfile(t *testing.T) {
	g := getUserDbReturn{
		DefaultProfile: pgtype.Text{
			String: "",
			Valid:  true,
		},
	}

	expectedUser := sophrosyne.User{
		ID:      g.ID,
		Name:    g.Name,
		Email:   g.Email,
		Token:   g.Token,
		IsAdmin: g.IsAdmin,
		DefaultProfile: sophrosyne.Profile{
			ID:   "lol42",
			Name: DefaultProfileName,
		},
		CreatedAt: g.CreatedAt,
		UpdatedAt: g.UpdatedAt,
		DeletedAt: g.DeletedAt,
	}

	mockedProfileService := sophrosyne2.NewMockProfileService(t)
	mockedProfileService.On("GetProfileByName", context.Background(), DefaultProfileName).Return(expectedUser.DefaultProfile, nil)

	user, err := g.ToUser(context.Background(), mockedProfileService)

	require.NoError(t, err)
	require.Equal(t, expectedUser, user)
	require.True(t, mockedProfileService.AssertExpectations(t))
}

func TestGetUserDbReturn_ToUser_DefaultProfileStringEmpty_ReturnsError(t *testing.T) {
	g := getUserDbReturn{
		DefaultProfile: pgtype.Text{
			String: "",
			Valid:  true,
		},
	}

	expectedUser := sophrosyne.User{}

	mockedProfileService := sophrosyne2.NewMockProfileService(t)
	mockedProfileService.On("GetProfileByName", context.Background(), DefaultProfileName).Return(sophrosyne.Profile{}, assert.AnError)

	user, err := g.ToUser(context.Background(), mockedProfileService)

	require.ErrorIs(t, err, assert.AnError)
	require.Equal(t, expectedUser, user)
	require.True(t, mockedProfileService.AssertExpectations(t))
}

func TestGetUserDbReturn_ToUser_NotDefaultProfile_DatabaseError(t *testing.T) {
	g := getUserDbReturn{
		DefaultProfile: pgtype.Text{
			String: "someID",
			Valid:  true,
		},
	}

	expectedUser := sophrosyne.User{}

	mockedProfileService := sophrosyne2.NewMockProfileService(t)
	mockedProfileService.On("GetProfile", context.Background(), g.DefaultProfile.String).Return(sophrosyne.Profile{}, assert.AnError)

	user, err := g.ToUser(context.Background(), mockedProfileService)

	require.ErrorIs(t, err, assert.AnError)
	require.Equal(t, expectedUser, user)
	require.True(t, mockedProfileService.AssertExpectations(t))
}
