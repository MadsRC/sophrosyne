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
	"github.com/madsrc/sophrosyne"
	"github.com/madsrc/sophrosyne/internal/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"log/slog"
	"net"
	"testing"
)

// Assigns the provided validator to the server's validator field.
func TestWithValidator_AssignsValidator(t *testing.T) {
	v := validator.NewValidator()
	s := &Server{}

	option := WithValidator(v)
	option(s)

	require.NotNil(t, s.validator)
	assert.Equal(t, v, s.validator)
}

// Handles nil Server instance gracefully without panicking.
func TestWithValidator_HandlesNilServer(t *testing.T) {
	v := validator.NewValidator()

	option := WithValidator(v)

	require.NotPanics(t, func() {
		option(nil)
	})
}

// Does not alter other fields of the server struct.
func TestWithValidator_DoesNotAlterOtherFields(t *testing.T) {
	// Create a new validator
	v := &validator.Validator{}

	// Create a new server with some initial values
	s := &Server{
		grpcServer: &grpc.Server{},
		listener:   &net.TCPListener{},
		config:     &sophrosyne.Config{},
		logger:     &slog.Logger{},
		validator:  &validator.Validator{},
	}

	// Call the WithValidator function
	opt := WithValidator(v)
	opt(s)

	// Validate that only the validator field was altered
	require.Equal(t, v, s.validator)
	require.NotNil(t, s.grpcServer)
	require.NotNil(t, s.listener)
	require.NotNil(t, s.config)
	require.NotNil(t, s.logger)
}

// Works correctly when a valid Server instance is passed.
func TestWithValidator_ValidServerInstance(t *testing.T) {
	// Create a new validator
	v := validator.NewValidator()

	// Create a new Server instance
	s := &Server{
		grpcServer: &grpc.Server{},
		listener:   &net.TCPListener{},
		config:     &sophrosyne.Config{},
		logger:     &slog.Logger{},
		validator:  nil,
	}

	// Call the WithValidator function with the validator
	opt := WithValidator(v)
	opt(s)

	// Assert that the validator in the Server instance is set to the created validator
	require.Equal(t, v, s.validator)
}

// Returns a Option function that can be executed without errors.
func TestWithValidator_ReturnsValidFunction(t *testing.T) {
	// Create a new validator
	v := validator.NewValidator()

	// Create a new Server
	s := &Server{}

	// Call the WithValidator function
	opt := WithValidator(v)

	// Execute the Option function returned
	opt(s)

	// Assert that the validator in the Server is set to the created validator
	require.Equal(t, v, s.validator)
}

// Works correctly when the provided validator is nil.
func TestWithValidator_NilValidator(t *testing.T) {
	// Setup
	var s Server

	// Execution
	opt := WithValidator(nil)
	opt(&s)

	// Assertion
	require.Nil(t, s.validator)
}

// Ensures no side effects when the server is nil.
func TestWithValidator_NoSideEffectsWhenServerIsNil(t *testing.T) {
	// Setup
	var validator *validator.Validator
	server := &Server{}

	// Execution
	opt := WithValidator(validator)
	opt(nil)

	// Validation
	require.Nil(t, server.validator, "Validator should not be set when server is nil")
}

// Does not modify the server if the validator is already set.
func TestWithValidator_ValidatorAlreadySet(t *testing.T) {
	// Setup
	validator := &validator.Validator{}
	server := &Server{
		validator: validator,
	}

	// Call the function
	opt := WithValidator(validator)
	opt(server)

	// Validate
	require.Equal(t, validator, server.validator, "Validator should not be modified if already set")
}

// Ensures idempotency when the Option function is called multiple times.
func TestWithValidator_Idempotency(t *testing.T) {
	// Create a new Server
	s := &Server{}

	// Create a Validator
	v := validator.NewValidator()

	// Call WithValidator function twice with the same Validator.
	opt1 := WithValidator(v)
	opt2 := WithValidator(v)

	// Apply the options to the Server
	opt1(s)
	opt2(s)

	// Validate that the validator is set only once
	require.Equal(t, v, s.validator)
}

// Validates that the Option function can be chained with other Option functions.
func TestWithValidator_Chained(t *testing.T) {
	require := require.New(t)

	// Create a new Server
	s := &Server{}

	// Create a Validator
	v := validator.NewValidator()

	// Create a Option with the Validator
	Option := WithValidator(v)

	// Create another Option
	anotherOption := func(s *Server) {
		// Do something else with the server
	}

	// Chain the Options
	Option(s)
	anotherOption(s)

	// Assert that the Validator is set in the Server
	require.Equal(v, s.validator)
}

// Checks that the Option function does not introduce memory leaks.
func TestWithValidator_NoMemoryLeaks(t *testing.T) {
	// Create a mock validator
	mockValidator := validator.NewValidator()

	// Create a new Server
	server := &ScanServiceServer{}

	// Call the WithValidator function with the mock validator
	opt := WithValidator(mockValidator)
	opt(server)

	// Assert that the validator in the server is the same as the mock validator
	require.Equal(t, mockValidator, server.validator)
}
