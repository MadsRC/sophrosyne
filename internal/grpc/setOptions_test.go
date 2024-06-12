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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/madsrc/sophrosyne/internal/log"
)

// Applies all default options to the target.
func TestSetOptions_AppliesAllDefaultOptions(t *testing.T) {
	logger, _ := log.NewTestLogger(nil)
	target := &Server{}
	defaultOptions := []Option{
		func(s any) {
			s.(*Server).logger = logger
		},
	}
	setOptions(target, defaultOptions)
	require.Equal(t, target.logger, logger)
}

// Applies all provided options to the target.
func TestSetOptions_AppliesAllOptions(t *testing.T) {
	logger, _ := log.NewTestLogger(nil)

	opt1 := func(s any) {
		s.(*Server).logger = logger
	}
	opt2 := func(s any) {
		// Option 2 implementation
	}

	target := &Server{}

	setOptions(target, []Option{opt1}, opt2)

	require.Equal(t, target.logger, logger)
}

// Applies provided options after default options.
func TestSetOptions_AppliesProvidedOptionsAfterDefaultOptions(t *testing.T) {
	logger1, _ := log.NewTestLogger(nil)
	logger2, _ := log.NewTestLogger(nil)
	// Define default options
	defaultOption1 := func(s any) {
		s.(*Server).logger = logger1
	}

	// Define provided options
	providedOption1 := func(s any) {
		s.(*Server).logger = logger2
	}

	// Create a target object
	target := &Server{}

	// Call setOptions with target, default options, and provided options
	setOptions(target, []Option{defaultOption1}, providedOption1)

	// Verify that the target's Logger is logger2
	require.Equal(t, target.logger, logger2)
}
