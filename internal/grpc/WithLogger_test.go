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
	"github.com/madsrc/sophrosyne/internal/log"
	"github.com/madsrc/sophrosyne/internal/validator"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"log/slog"
	"net"
	"os"
	"sync"
	"testing"
)

// sets the log for a valid Server instance.
func TestWithLogger_SetsLogger(t *testing.T) {
	logger, _ := log.NewTestLogger(nil)
	server := &Server{}

	option := WithLogger(logger)
	option(server)

	require.NotNil(t, server.logger)
	require.Equal(t, logger, server.logger)
}

// handles nil Server instance gracefully.
func TestWithLogger_NilServer(t *testing.T) {
	logger, _ := log.NewTestLogger(nil)

	option := WithLogger(logger)
	option(nil)

	// No panic or error should occur, and nothing to assert as the server is nil
}

// does not modify other fields of the Server instance.
func TestWithLogger_DoesNotModifyOtherFields(t *testing.T) {
	// Setup
	s := &Server{
		grpcServer: &grpc.Server{},
		listener:   &net.TCPListener{},
		config:     &sophrosyne.Config{},
		logger:     &slog.Logger{},
		validator:  &validator.Validator{},
	}
	logger := &slog.Logger{}

	// Execute
	opt := WithLogger(logger)
	opt(s)

	// Verify
	require.Equal(t, &grpc.Server{}, s.grpcServer)
	require.Equal(t, &net.TCPListener{}, s.listener)
	require.Equal(t, &sophrosyne.Config{}, s.config)
	require.Equal(t, &validator.Validator{}, s.validator)
	require.Equal(t, s.logger, logger)
}

// returns a Option function.
func TestWithLogger_ReturnsOptionFunction(t *testing.T) {
	// Setup
	logger := &slog.Logger{}

	// Execute
	Option := WithLogger(logger)

	// Verify
	require.NotNil(t, Option)
}

// can be used in conjunction with other Option functions.
func TestWithLoggerInConjunctionWithOtherOptions(t *testing.T) {
	// Setup
	logger, _ := log.NewTestLogger(nil)
	server := &ScanServiceServer{}

	// Execution
	opt1 := WithLogger(logger)
	opt2 := func(s any) {
		if s == nil {
			return
		}
		// Additional logic for another Option
	}

	opt1(server)
	opt2(server)

	// Assertion
	require.Equal(t, logger, server.logger)
}

// does not panic when log is nil.
func TestWithLogger_LoggerIsNil_DoesNotPanic(t *testing.T) {
	// Setup
	opt := WithLogger(nil)
	s := &Server{}

	// Execution
	opt(s)

	// Assertion
	require.NotNil(t, s)
	require.Nil(t, s.logger)
}

// does not modify the Server instance if the log is nil.
func TestWithLogger_LoggerIsNil(t *testing.T) {
	// Setup
	s := &Server{}
	logger := (*slog.Logger)(nil)

	// Execute
	opt := WithLogger(logger)
	opt(s)

	// Verify
	require.Nil(t, s.logger, "Server log should not be modified if log is nil")
}

// handles concurrent access to the Server instance.
func TestWithLogger_ConcurrentAccess(t *testing.T) {
	// Create a new Server instance
	server := &Server{}

	// Create a wait group to simulate concurrent access
	var wg sync.WaitGroup
	wg.Add(2)

	// Concurrently call WithLogger on the Server instance
	go func() {
		defer wg.Done()
		WithLogger(&slog.Logger{})(server)
	}()

	go func() {
		defer wg.Done()
		WithLogger(&slog.Logger{})(server)
	}()

	// Wait for goroutines to finish
	wg.Wait()

	// Assert that the log is set on the Server instance
	require.NotNil(t, server.logger)
}

// ensures log is correctly assigned even if previously set.
func TestWithLogger_LoggerPreviouslySet(t *testing.T) {
	// Create a new log
	logger, _ := log.NewTestLogger(nil)

	// Create a new server with a different log
	server := &Server{
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	// Call WithLogger with the new log
	opt := WithLogger(logger)
	opt(server)

	// Assert that the server's log is now the new log
	require.Equal(t, logger, server.logger)
}

// validates the log type before assignment.
func TestWithLogger_ValidateLoggerType(t *testing.T) {
	// Setup
	logger, _ := log.NewTestLogger(nil)
	server := &Server{}

	// Execution
	opt := WithLogger(logger)
	opt(server)

	// Assertion
	require.Equal(t, logger, server.logger)
}

// ensures idempotency when the same log is set multiple times.
func TestWithLogger_Idempotency(t *testing.T) {
	// Create a new server
	s := &Server{}

	// Create a log
	logger := &slog.Logger{}

	// Set the log using WithLogger function twice
	WithLogger(logger)(s)
	WithLogger(logger)(s)

	// Ensure that the log is set only once
	require.Equal(t, logger, s.logger)
}
