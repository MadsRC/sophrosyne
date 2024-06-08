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
	"github.com/madsrc/sophrosyne"
	v0 "github.com/madsrc/sophrosyne/internal/grpc/sophrosyne/v0"
	"github.com/madsrc/sophrosyne/internal/log"
	v02 "github.com/madsrc/sophrosyne/internal/mocks/internal_/grpc/sophrosyne/v0"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
	"time"
)

// Successfully connects to upstream service and returns a valid CheckResult.
func TestDoCheck_SuccessfullyConnectsToUpstreamServiceAndReturnsValidCheckResult(t *testing.T) {
	// Create a mock logger
	logger, _ := log.NewTestLogger(nil)

	// Create a mock Check instance
	check := sophrosyne.Check{
		ID:       "123",
		Name:     "TestCheck",
		Profiles: []sophrosyne.Profile{},
		UpstreamServices: []url.URL{
			{Host: "localhost"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: nil,
	}

	// Create a mock CheckProviderServiceClient
	client := v02.NewMockCheckProviderServiceClient(t)

	// Create a mock ScanRequest
	req := &v0.ScanRequest{
		Profile: "test_profile",
		Kind: &v0.ScanRequest_Text{
			Text: "test_text",
		},
	}

	client.On("Check", mock.Anything, mock.Anything).Return(&v0.CheckProviderResponse{
		Result: true,
	}, nil)

	// Call the doCheck function
	result, err := doCheck(context.Background(), logger, check, client, req)

	// Assert that the result is as expected
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, check.Name, result.Name)
	require.True(t, result.Result)
	require.True(t, client.AssertExpectations(t))
}

// Upstream service returns error.
func TestDoCheck_UpstreamServiceReturnsError(t *testing.T) {
	// Create a mock logger
	logger, _ := log.NewTestLogger(nil)

	// Create a mock Check instance
	check := sophrosyne.Check{
		ID:       "123",
		Name:     "TestCheck",
		Profiles: []sophrosyne.Profile{},
		UpstreamServices: []url.URL{
			{Host: "localhost"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: nil,
	}

	// Create a mock CheckProviderServiceClient
	client := v02.NewMockCheckProviderServiceClient(t)

	// Create a mock ScanRequest
	req := &v0.ScanRequest{
		Profile: "test_profile",
		Kind: &v0.ScanRequest_Text{
			Text: "test_text",
		},
	}

	client.On("Check", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	// Call the doCheck function
	result, err := doCheck(context.Background(), logger, check, client, req)

	// Assert that the result is as expected
	require.Error(t, err)
	require.ErrorIs(t, err, assert.AnError)
	require.Nil(t, result)
	require.True(t, client.AssertExpectations(t))
}

// When not provided a client, a client is constructed.
func TestDoCheck_NoClientProvided(t *testing.T) {
	// Create a mock logger
	logger, _ := log.NewTestLogger(nil)

	// Create a mock Check instance
	check := sophrosyne.Check{
		ID:       "123",
		Name:     "TestCheck",
		Profiles: []sophrosyne.Profile{},
		UpstreamServices: []url.URL{
			{Host: ""},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: nil,
	}
	// Create a mock ScanRequest
	req := &v0.ScanRequest{
		Profile: "test_profile",
		Kind: &v0.ScanRequest_Text{
			Text: "test_text",
		},
	}

	// Call the doCheck function
	result, err := doCheck(context.Background(), logger, check, nil, req)

	// Assert that the result is as expected
	require.Error(t, err)
	require.Nil(t, result)
}

// No upstream services.
func TestDoCheck_NoUpstreamServices(t *testing.T) {
	// Create a mock logger
	logger, _ := log.NewTestLogger(nil)

	// Create a mock Check instance
	check := sophrosyne.Check{
		ID:               "123",
		Name:             "TestCheck",
		Profiles:         []sophrosyne.Profile{},
		UpstreamServices: []url.URL{},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		DeletedAt:        nil,
	}
	// Create a mock ScanRequest
	req := &v0.ScanRequest{
		Profile: "test_profile",
		Kind: &v0.ScanRequest_Text{
			Text: "test_text",
		},
	}

	// Call the doCheck function
	result, err := doCheck(context.Background(), logger, check, nil, req)

	// Assert that the result is as expected
	require.Error(t, err)
	require.Nil(t, result)
}
