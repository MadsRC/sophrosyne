package services

import (
	"context"
	"net/url"
	"testing"

	"github.com/madsrc/sophrosyne"
	"github.com/madsrc/sophrosyne/internal/logger"
	sophrosyne2 "github.com/madsrc/sophrosyne/internal/mocks"
	"github.com/madsrc/sophrosyne/internal/rpc/jsonrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// User is successfully extracted from the context
func TestScanService_PerformScan_userExtractedSuccessfully(t *testing.T) {
	ctx := context.WithValue(context.Background(), sophrosyne.UserContextKey{}, &sophrosyne.User{})
	req := jsonrpc.Request{ID: jsonrpc.NewID("1"), Method: "scan", Params: &jsonrpc.ParamsObject{}}
	expectedResponse := []byte(`{"id":"1","result":{"result":true,"checks":{}}}`)

	logger, logBuf := logger.NewTestLogger(nil)
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
			{Name: "testCheck", UpstreamServices: []url.URL{{Scheme: "http", Host: "localhost"}, {Scheme: "http", Host: "127.0.0.1"}}},
		},
	}, nil)

	response, err := scanService.PerformScan(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, expectedResponse, response)
}
