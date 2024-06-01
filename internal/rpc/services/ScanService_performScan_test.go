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
	"github.com/madsrc/sophrosyne/internal/log"
	sophrosyne2 "github.com/madsrc/sophrosyne/internal/mocks"
	"github.com/madsrc/sophrosyne/internal/rpc/jsonrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// User is successfully extracted from the context
func TestScanService_PerformScan_userExtractedSuccessfully(t *testing.T) {
	ctx := context.WithValue(context.Background(), sophrosyne.UserContextKey{}, &sophrosyne.User{})
	req := jsonrpc.Request{ID: jsonrpc.NewID("1"), Method: "scan", Params: &jsonrpc.ParamsObject{
		"text": "this is some text",
	}}
	expectedResponse := []byte(`{"jsonrpc":"2.0","id":"1","result":{"result":false,"checks":{}}}`)

	logger, _ := log.NewTestLogger(nil)
	mockValidator := sophrosyne2.NewMockValidator(t)
	mockProfileService := sophrosyne2.NewMockProfileService(t)
	scanService := ScanService{
		logger:         logger,
		validator:      mockValidator,
		profileService: mockProfileService,
	}

	mockValidator.On("Validate", mock.Anything).Return(nil)
	mockProfileService.On("GetProfileByName", mock.Anything, "default").Return(sophrosyne.Profile{
		Checks: []sophrosyne.Check{
			{Name: "testCheck"},
		},
	}, nil)

	response, err := scanService.PerformScan(ctx, req)

	require.NoError(t, err)
	assert.JSONEq(t, string(expectedResponse), string(response))
}

// A check was attempted, but failed when attempting to call the upstream service.
func TestScanService_performScan_checkRunSuccessReturnNoSuccess(t *testing.T) {
	ctx := context.WithValue(context.Background(), sophrosyne.UserContextKey{}, &sophrosyne.User{})
	req := jsonrpc.Request{ID: jsonrpc.NewID("1"), Method: "scan", Params: &jsonrpc.ParamsObject{}}
	expectedResponse := []byte(`{"jsonrpc":"2.0","id":"1","result":{"result":false,"checks":{"testCheck":{"status":false, "detail":"error calling upstream service"}}}}`)

	logger, _ := log.NewTestLogger(nil)
	scanService := ScanService{
		logger: logger,
	}

	profile := sophrosyne.Profile{
		Checks: []sophrosyne.Check{
			{
				Name: "testCheck",
				UpstreamServices: []url.URL{{
					Scheme: "http",
					Host:   "127.0.0.1",
				}},
			},
		},
	}

	response, err := scanService.performScan(ctx, req, &profile, sophrosyne.PerformScanRequest{})

	require.NoError(t, err)
	assert.JSONEq(t, string(expectedResponse), string(response))
}

// Performs all checks in the profile and returns the results
func TestPerformScan_PerformsAllChecks(t *testing.T) {
	ctx := context.Background()
	logger, _ := log.NewTestLogger(nil)
	profile := &sophrosyne.Profile{
		Name: "test-profile",
		Checks: []sophrosyne.Check{
			{Name: "check1", UpstreamServices: []url.URL{{Host: "localhost:50051"}}},
			{Name: "check2", UpstreamServices: []url.URL{{Host: "localhost:50052"}}},
		},
	}
	req := jsonrpc.Request{ID: jsonrpc.NewID("1"), Method: "performScan"}
	expectedResponse := []byte(`{"jsonrpc":"2.0","id":"1","result":{"result":false,"checks":{"check1":{"status":false, "detail":"error calling upstream service"},"check2":{"status":false, "detail":"error calling upstream service"}}}}`)

	p := ScanService{logger: logger}

	result, err := p.performScan(ctx, req, profile, sophrosyne.PerformScanRequest{})
	require.NoError(t, err)

	assert.JSONEq(t, string(expectedResponse), string(result))
}

// Profile has no checks
func TestPerformScan_ProfileHasNoChecks(t *testing.T) {
	ctx := context.Background()
	logger, _ := log.NewTestLogger(nil)
	profile := &sophrosyne.Profile{
		Name:   "test-profile",
		Checks: []sophrosyne.Check{},
	}
	req := jsonrpc.Request{ID: jsonrpc.NewID("1"), Method: "performScan"}
	expectedResponse := []byte(`{"jsonrpc":"2.0","id":"1","result":{"result":false,"checks":{}}}`)

	p := ScanService{logger: logger}

	result, err := p.performScan(ctx, req, profile, sophrosyne.PerformScanRequest{})
	require.NoError(t, err)

	assert.JSONEq(t, string(expectedResponse), string(result))
}
