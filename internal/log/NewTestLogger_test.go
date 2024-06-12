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

package log

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

// Verify function returns a *slog.Logger and *bytes.Buffer
func TestNewTestLogger_test_ReturnsLoggerAndBuffer(t *testing.T) {
	logger, buffer := NewTestLogger(nil)
	require.NotNil(t, logger, "Expected non-nil *slog.Logger")
	require.NotNil(t, buffer, "Expected non-nil *bytes.Buffer")
}

// Test with nil opts to confirm default settings are applied
func TestNewTestLogger_test_NilOptsDefaultSettings(t *testing.T) {
	logger, _ := NewTestLogger(nil)
	require.True(t, logger.Handler().Enabled(context.Background(), slog.LevelDebug), "Expected log level to be debug")
}

// Check if JSON handler is correctly attached to the log
func TestNewTestLogger_test_JSONHandlerAttached(t *testing.T) {
	logger, buffer := NewTestLogger(nil)
	require.NotNil(t, logger, "Expected non-nil *slog.Logger")
	require.NotNil(t, buffer, "Expected non-nil *bytes.Buffer")
	// Check if JSON handler is attached
	_, ok := logger.Handler().(*slog.JSONHandler)
	require.True(t, ok, "Expected JSON handler to be attached to the log")
}

// Confirm that the buffer is empty upon initialization
func TestNewTestLogger_test_BufferEmptyUponInitialization(t *testing.T) {
	logger, buffer := NewTestLogger(nil)
	require.NotNil(t, logger, "Expected non-nil *slog.Logger")
	require.NotNil(t, buffer, "Expected non-nil *bytes.Buffer")
	require.Empty(t, buffer.String(), "Expected empty buffer upon initialization")
}

// Validate that provided opts are used when not nil
func TestNewTestLogger_test_OptsUsedWhenNotNil(t *testing.T) {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	logger, _ := NewTestLogger(opts)
	require.True(t, logger.Handler().Enabled(context.Background(), slog.LevelInfo), "Expected log level to be info")
}

// Monitor for correct JSON formatting in the buffer output
func TestNewTestLogger_test_JSONFormattingInBuffer(t *testing.T) {
	logger, buffer := NewTestLogger(nil)

	logger.Info("this is a message")

	var parsedLog map[string]interface{}
	err := json.Unmarshal(buffer.Bytes(), &parsedLog)

	require.NoError(t, err, "Expected no error while unmarshalling JSON")
	require.Contains(t, buffer.String(), "this is a message", "Expected buffer to contain log message")
}

// Check for any side effects when reusing the buffer
func TestNewTestLogger_test_ReusingBuffer(t *testing.T) {
	logger, buffer := NewTestLogger(nil)
	require.NotNil(t, logger, "Expected non-nil *slog.Logger")
	require.NotNil(t, buffer, "Expected non-nil *bytes.Buffer")

	// Write to the buffer
	logger.Info("Test message")
	require.Contains(t, buffer.String(), "Test message", "Expected buffer to contain test message")

	// Reuse the buffer
	logger.Info("Another message")
	require.Contains(t, buffer.String(), "Another message", "Expected buffer to contain another message")
}
