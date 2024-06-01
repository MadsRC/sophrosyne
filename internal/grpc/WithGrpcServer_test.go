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
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"testing"
)

// Assigns grpcServer to Server when Server is not nil.
func TestWithGrpcServer_AssignsGrpcServer(t *testing.T) {
	grpcServer := &grpc.Server{}
	server := &Server{}

	option := WithGrpcServer(grpcServer)
	option(server)

	require.Equal(t, grpcServer, server.grpcServer, "grpcServer should be assigned to Server")
}

// Server is nil, function should handle gracefully.
func TestWithGrpcServer_ServerIsNil(t *testing.T) {
	grpcServer := &grpc.Server{}

	option := WithGrpcServer(grpcServer)
	option(nil)

	// No panic or error should occur, so just pass the test
	require.True(t, true, "Function should handle nil Server gracefully")
}

// Option function modifies Server as expected.
func TestWithGrpcServer_ModifiesServerAsExpected(t *testing.T) {
	// Setup
	s := &Server{}
	grpcServer := &grpc.Server{}

	// Call the function
	opt := WithGrpcServer(grpcServer)
	opt(s)

	// Validate the modification
	require.Equal(t, grpcServer, s.grpcServer)
}

// Returns a valid Option function.
func TestWithGrpcServer_ReturnsValidOption(t *testing.T) {
	// Create a new Server
	s := &Server{}

	// Create a new grpc.Server
	grpcServer := grpc.NewServer()

	// Call the WithGrpcServer function
	option := WithGrpcServer(grpcServer)

	// Apply the option to the Server
	option(s)

	// Check if the grpcServer field in the Server is set to the provided grpc.Server
	require.Equal(t, grpcServer, s.grpcServer)
}

// grpcServer is nil, function should still assign nil to Server
func TestWithGrpcServer_GrpcServerIsNil(t *testing.T) {
	// Setup
	var s Server
	var grpcServer *grpc.Server = nil

	// Execute
	opt := WithGrpcServer(grpcServer)
	opt(&s)

	// Verify
	require.Nil(t, s.grpcServer)
}

// Server already has a grpcServer assigned, function should overwrite it
func TestWithGrpcServer_OverwriteGrpcServer(t *testing.T) {
	// Setup
	s := &Server{
		grpcServer: grpc.NewServer(),
	}
	newGrpcServer := grpc.NewServer()

	// Execute
	opt := WithGrpcServer(newGrpcServer)
	opt(s)

	// Validate
	require.Equal(t, newGrpcServer, s.grpcServer)
}

// Ensure no other fields in Server are modified
func TestWithGrpcServer_NoOtherFieldsModified(t *testing.T) {
	// Setup
	s := &Server{
		grpcServer: nil,
		listener:   nil,
		config:     nil,
		logger:     nil,
		validator:  nil,
	}

	grpcServer := &grpc.Server{}

	// Execute
	opt := WithGrpcServer(grpcServer)
	opt(s)

	// Verify
	require.Equal(t, grpcServer, s.grpcServer)
	require.Nil(t, s.listener)
	require.Nil(t, s.config)
	require.Nil(t, s.logger)
	require.Nil(t, s.validator)
}

// Validate that the returned Option is a function
func TestWithGrpcServer_ReturnedFunction(t *testing.T) {
	// Create a dummy grpc server
	dummyGrpcServer := grpc.NewServer()

	// Call the WithGrpcServer function
	option := WithGrpcServer(dummyGrpcServer)

	// Check if the returned option is a function
	require.NotNil(t, option)
	require.IsType(t, Option(nil), option)
}

// Verify that the function does not panic with invalid inputs
func TestWithGrpcServer_InvalidInputs(t *testing.T) {
	// Create a dummy server
	s := &Server{}

	// Call the function with nil input
	opt := WithGrpcServer(nil)
	opt(s)

	// Assert that the grpcServer field is still nil
	require.Nil(t, s.grpcServer)
}
