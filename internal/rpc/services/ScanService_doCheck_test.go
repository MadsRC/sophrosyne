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

	response, err := doCheck(ctx, logger, check, mockCheckServiceClient)

	require.NoError(t, err)
	require.Equal(t, expected, response)
}
