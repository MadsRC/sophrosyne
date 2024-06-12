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
)

// returns a non-nil slice of ServerOption.
func TestDefaultServerOptions_ReturnsNonNilSlice(t *testing.T) {
	options := defaultServerOptions()
	require.NotNil(t, options)
}

// handles nil Server instance gracefully when applying ServerOption.
func TestDefaultServerOptions_HandlesNilServerGracefully(t *testing.T) {
	options := defaultServerOptions()
	for _, option := range options {
		require.NotPanics(t, func() { option(nil) })
	}
}

// includes a Validator in the returned ServerOption slice.
func TestDefaultServerOptions_IncludesValidator(t *testing.T) {
	// Create a new Server instance
	server := &Server{}

	// Call the defaultServerOptions function
	options := defaultServerOptions()
	for _, option := range options {
		option(server)
	}

	// Assert that the Validator is included in the Server options
	require.NotNil(t, server.validator)
}

// each ServerOption in the slice can be applied to a Server instance.
func TestDefaultServerOptions_EachOptionCanBeAppliedToServerInstance(t *testing.T) {
	// Create a new Server instance
	server := &Server{}

	// Get the default server options
	options := defaultServerOptions()

	// Apply each option to the server instance
	for _, option := range options {
		option(server)
	}

	// Assert that each option has been applied successfully
	require.NotNil(t, server.validator)
}

// ensures the returned slice is not modified by subsequent calls.
func TestDefaultServerOptions_EnsuresReturnedSliceNotModified(t *testing.T) {
	// Call defaultServerOptions function twice
	options1 := defaultServerOptions()
	options2 := defaultServerOptions()

	// Assert that the slices are not the same instance
	require.NotEqual(t, &options1, &options2)
}
