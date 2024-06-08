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

package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/madsrc/sophrosyne"
	sophrosyne2 "github.com/madsrc/sophrosyne/internal/mocks"
	"github.com/madsrc/sophrosyne/internal/validator"
)

// Tests that a default Validator is attached to the server
func TestNewServer_HasDefaultValidator(t *testing.T) {
	server, _ := NewServer(context.Background())
	require.NotNil(t, server.validator)
}

// Test that a provided Validator replaces the default Validator
func TestNewServer_UsesProvidedValidator(t *testing.T) {
	v := validator.NewValidator()
	server, _ := NewServer(context.Background(), WithValidator(v))
	require.Equal(t, v, server.validator)
}

// Test that no grpcServer is provided, validation fails
func TestNewServer_ReturnsErrorIfNoGrpcServer(t *testing.T) {
	// TODO: Fix this - The "required" tag on the grpcServer pointer should cause an error when validated and the pointer is nil.
	t.Skip()
	server, err := NewServer(context.Background())
	require.Nil(t, server)
	require.Error(t, err)
	var ve *sophrosyne.ValidationError
	require.ErrorAs(t, err, &ve)
}

// Test that if validation fails, an error is returned
func TestNewServer_ReturnsErrorIfValidationFails(t *testing.T) {
	mockValidator := sophrosyne2.NewMockValidator(t)
	mockValidator.On("Validate", mock.Anything).Return(assert.AnError)
	server, err := NewServer(context.Background(), WithValidator(mockValidator))
	require.Nil(t, server)
	require.Error(t, err)
	require.ErrorIs(t, err, assert.AnError)
}
