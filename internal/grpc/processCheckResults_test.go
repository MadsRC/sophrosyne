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

	"github.com/stretchr/testify/require"

	v0 "github.com/madsrc/sophrosyne/internal/grpc/sophrosyne/v0"
	"github.com/madsrc/sophrosyne/internal/log"
)

// processes multiple valid check results correctly.
func TestProcessCheckResults_ProcessesMultipleValidCheckResultsCorrectly(t *testing.T) {
	ctx := context.Background()
	logger, _ := log.NewTestLogger(nil)
	messages := make(chan *v0.CheckResult, 3)

	messages <- &v0.CheckResult{Name: "check1", Result: true}
	messages <- &v0.CheckResult{Name: "check2", Result: false}
	messages <- &v0.CheckResult{Name: "check3", Result: true}
	close(messages)

	result := processCheckResults(ctx, messages, logger)

	require.Len(t, result.Checks, 3)
	require.Equal(t, "check1", result.Checks[0].Name)
	require.True(t, result.Checks[0].Result)
	require.Equal(t, "check2", result.Checks[1].Name)
	require.False(t, result.Checks[1].Result)
	require.Equal(t, "check3", result.Checks[2].Name)
	require.True(t, result.Checks[2].Result)
	require.True(t, result.Result)
}

// ignores check results with empty names.
func TestProcessCheckResults_IgnoresCheckResultsWithEmptyNames(t *testing.T) {
	ctx := context.Background()
	logger, _ := log.NewTestLogger(nil)
	messages := make(chan *v0.CheckResult, 3)

	messages <- &v0.CheckResult{Name: "", Result: true}
	messages <- &v0.CheckResult{Name: "check2", Result: true}
	messages <- &v0.CheckResult{Name: "", Result: true}
	close(messages)

	result := processCheckResults(ctx, messages, logger)

	require.Len(t, result.Checks, 1)
	require.Equal(t, "check2", result.Checks[0].Name)
	require.True(t, result.Checks[0].Result)
	require.True(t, result.Result)
}

// does not panic when a nil message is received.
func TestProcessCheckResults_DoesNotPanicWhenNilMessageIsReceived(t *testing.T) {
	ctx := context.Background()
	logger, _ := log.NewTestLogger(nil)
	messages := make(chan *v0.CheckResult, 3)

	messages <- &v0.CheckResult{Name: "check1", Result: true}
	messages <- nil
	messages <- &v0.CheckResult{Name: "check3", Result: true}
	close(messages)

	result := processCheckResults(ctx, messages, logger)

	require.Len(t, result.Checks, 2)
	require.Equal(t, "check1", result.Checks[0].Name)
	require.True(t, result.Checks[0].Result)
	require.Equal(t, "check3", result.Checks[1].Name)
	require.True(t, result.Checks[1].Result)
	require.True(t, result.Result)
}
