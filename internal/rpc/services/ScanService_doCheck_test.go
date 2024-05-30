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

package services

import (
	"context"
	"net/url"
	"testing"

	"github.com/madsrc/sophrosyne"
	"github.com/madsrc/sophrosyne/internal/grpc/checks"
	"github.com/madsrc/sophrosyne/internal/logger"
	checks2 "github.com/madsrc/sophrosyne/internal/mocks/internal_/grpc/checks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// User is successfully extracted from the context
func TestScanService_doCheck_checkResultNameNotEmpty(t *testing.T) {
	ctx := context.Background()
	expected := checkResult{
		Name:   "value",
		Status: false,
		Detail: "something",
	}

	logger, _ := logger.NewTestLogger(nil)
	mockCheckServiceClient := checks2.NewMockCheckServiceClient(t)
	mockCheckServiceClient.On("Check", ctx, mock.Anything).Return(&checks.CheckResponse{
		Result:  false,
		Details: "something",
	}, nil)

	check := sophrosyne.Check{
		Name: "value",
		UpstreamServices: []url.URL{{
			Host: "127.0.0.1",
		}},
	}

	response, err := doCheck(ctx, logger, check, mockCheckServiceClient, sophrosyne.PerformScanRequest{})

	require.NoError(t, err)
	require.Equal(t, expected, response)
}
