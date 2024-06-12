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

// returns CheckProviderRequest with image when ScanRequest contains image data.
func TestCheckProviderRequestFromScanRequest_ReturnsCheckProviderRequestWithImage(t *testing.T) {
	ctx := context.Background()
	logger, _ := log.NewTestLogger(nil)
	imageData := []byte{0x01, 0x02, 0x03}
	req := &v0.ScanRequest{
		Kind: &v0.ScanRequest_Image{Image: imageData},
	}

	result := checkProviderRequestFromScanRequest(ctx, logger, req)

	require.NotNil(t, result)
	require.IsType(t, &v0.CheckProviderRequest_Image{}, result.Check)
	require.Equal(t, imageData, result.GetImage())
}

// handles empty ScanRequest gracefully.
func TestCheckProviderRequestFromScanRequest_HandlesEmptyScanRequestGracefully(t *testing.T) {
	ctx := context.Background()
	logger, _ := log.NewTestLogger(nil)
	req := &v0.ScanRequest{}

	result := checkProviderRequestFromScanRequest(ctx, logger, req)

	require.Equal(t, "", result.GetText())
}

// returns CheckProviderRequest with text when ScanRequest contains text data.
func TestCheckProviderRequestFromScanRequest_ReturnsCheckProviderRequestWithTextNewLogger(t *testing.T) {
	ctx := context.Background()
	logger, _ := log.NewTestLogger(nil)
	text := "sample text"
	req := &v0.ScanRequest{
		Profile: "sample profile",
		Kind:    &v0.ScanRequest_Text{Text: text},
	}

	result := checkProviderRequestFromScanRequest(ctx, logger, req)

	require.NotNil(t, result)
	require.IsType(t, &v0.CheckProviderRequest{}, result)
	require.IsType(t, &v0.CheckProviderRequest_Text{}, result.GetCheck())
	require.Equal(t, text, result.GetCheck().(*v0.CheckProviderRequest_Text).Text)
}
