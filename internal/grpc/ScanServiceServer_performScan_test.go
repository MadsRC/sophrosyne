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
	"time"

	"github.com/stretchr/testify/require"

	"github.com/madsrc/sophrosyne"
	v0 "github.com/madsrc/sophrosyne/internal/grpc/sophrosyne/v0"
	"github.com/madsrc/sophrosyne/internal/log"
	sophrosyne2 "github.com/madsrc/sophrosyne/internal/mocks"
	"github.com/madsrc/sophrosyne/internal/validator"
)

func TestPerformScan(t *testing.T) {
	logger, _ := log.NewTestLogger(nil)
	server := &ScanServiceServer{
		Logger:         logger,
		Config:         &sophrosyne.Config{},
		Validator:      &validator.Validator{},
		ProfileService: sophrosyne2.NewMockProfileService(t),
	}

	ctx := context.Background()

	mockProfile := &sophrosyne.Profile{
		ID:   "1",
		Name: "Mock Profile",
		Checks: []sophrosyne.Check{
			{
				ID:   "1",
				Name: "Mock Check 1",
			},
			{
				ID:   "2",
				Name: "Mock Check 2",
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	response, err := server.performScan(ctx, &v0.ScanRequest{}, mockProfile)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.False(t, response.Result)
	require.Empty(t, response.Checks)
}
