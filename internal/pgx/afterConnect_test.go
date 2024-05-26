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
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"log/slog"
	"testing"
)

// Verify that the function logs 'database connection established', at the debug level, when called.
func TestAfterConnect_LogsMessage(t *testing.T) {
	buf := bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	afterConnectFunc := afterConnect(logger)
	err := afterConnectFunc(context.Background(), nil)
	require.NoError(t, err)
	logEvent := map[string]interface{}{}
	var logEvents [][]byte

	scanner := bufio.NewScanner(&buf)
	for scanner.Scan() {
		line := scanner.Bytes()
		logEvents = append(logEvents, line)
	}

	require.Len(t, logEvents, 1)

	err = json.Unmarshal(logEvents[0], &logEvent)
	require.NoError(t, err)
	require.Equal(t, "database connection established", logEvent["msg"])
	require.Equal(t, "DEBUG", logEvent["level"])
}

// Check behavior when logger is nil.
func TestAfterConnect_WithNilLogger(t *testing.T) {
	afterConnectFunc := afterConnect(nil)

	require.Panics(t, func() {
		_ = afterConnectFunc(context.Background(), nil)
	})
}
