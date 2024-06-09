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
	"net"
	"testing"
	"time"

	"github.com/madsrc/sophrosyne/internal/log"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/madsrc/sophrosyne"
	"github.com/madsrc/sophrosyne/internal/validator"
)

// Serve starts the gRPC server successfully
func TestServe_StartsGRPCServerSuccessfully(t *testing.T) {
	// Create a mock listener
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer listener.Close()

	// Create a mock gRPC server
	grpcServer := grpc.NewServer()

	logger, _ := log.NewTestLogger(nil)
	// Create a new Server instance
	server := &Server{
		grpcServer: grpcServer,
		listener:   listener,
		config:     &sophrosyne.Config{},
		logger:     logger,
		validator:  &validator.Validator{},
	}

	// Run Serve in a separate goroutine
	go func() {
		_ = server.Serve()
	}()

	// Give some time for the server to start
	time.Sleep(100 * time.Millisecond)

	// Check if the server is serving
	conn, err := grpc.NewClient(listener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()
}
