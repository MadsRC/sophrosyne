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

	"github.com/madsrc/sophrosyne/internal/validator"
)

// returns a slice of Option with a validator
func TestDefaultScanServiceServerOptions_ReturnsSliceWithValidator(t *testing.T) {
	options := defaultScanServiceServerOptions()

	require.NotNil(t, options)
	require.Len(t, options, 1)

	scanServiceServer := &ScanServiceServer{}
	options[0](scanServiceServer)

	require.NotNil(t, scanServiceServer.validator)
}

// NewValidator returns nil
func TestNewValidator_ReturnsNonNil(t *testing.T) {
	validator := validator.NewValidator()

	require.NotNil(t, validator)
}
