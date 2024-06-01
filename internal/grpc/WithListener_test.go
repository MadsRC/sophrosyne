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
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"log/slog"
	"net"
	"testing"
)

// Server's listener is set correctly when a valid listener is provided.
func TestWithListener_SetsListener(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	server := &Server{}
	option := WithListener(listener)
	option(server)

	require.Equal(t, listener, server.listener)
}

// Function handles nil Server gracefully without causing a panic.
func TestWithListener_NilServer(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	option := WithListener(listener)
	require.NotPanics(t, func() {
		option(nil)
	})
}

// Function returns a valid Option when called with a valid listener.
func TestWithListener_ValidListener(t *testing.T) {
	// Create a dummy net.Listener
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	// Call the WithListener function
	server := &Server{}
	option := WithListener(listener)
	option(server)

	// Assert that the listener in the server is set to the provided listener.
	require.Equal(t, listener, server.listener)
}

// Function handles nil listener gracefully and sets Server's listener to nil.
func TestWithListener_NilListener(t *testing.T) {
	// Setup
	s := &Server{}

	// Call the function with nil listener
	WithListener(nil)(s)

	// Assertion
	require.Nil(t, s.listener)
}

// Server's listener is overwritten if the provided listener is nil.
func TestWithListener_RemainsUnchangedIfNil(t *testing.T) {
	// Setup
	s := &Server{
		grpcServer: &grpc.Server{},
		listener:   &net.TCPListener{},
		config:     &sophrosyne.Config{},
		logger:     &slog.Logger{},
		validator:  &validator.Validator{},
	}

	// Call the function with nil listener
	WithListener(nil)(s)

	// Assertion
	require.Nil(t, s.listener, "Server's listener should change if provided listener is nil")
}

// Verify that the function does not modify other fields of the Server.
func TestWithListener_DoesNotModifyOtherFields(t *testing.T) {
	// Create a new Server instance
	s := &Server{
		grpcServer: &grpc.Server{},
		listener:   nil,
		config:     &sophrosyne.Config{},
		logger:     &slog.Logger{},
		validator:  &validator.Validator{},
	}

	// Call the WithListener function
	newListener, _ := net.Listen("tcp", "127.0.0.1:0")
	opt := WithListener(newListener)
	opt(s)

	// Assert that only the listener field has been modified
	require.Equal(t, newListener, s.listener)
	require.Equal(t, &grpc.Server{}, s.grpcServer)
	require.Equal(t, &sophrosyne.Config{}, s.config)
	require.Equal(t, &slog.Logger{}, s.logger)
	require.Equal(t, &validator.Validator{}, s.validator)
}
