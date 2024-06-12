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
	"log/slog"
	"net"
	"strings"

	"google.golang.org/grpc"

	"github.com/madsrc/sophrosyne"
	"github.com/madsrc/sophrosyne/internal/validator"
)

const (
	InvalidTokenMsg  = "invalid token"
	InvalidCursorMsg = "invalid cursor"
)

type Server struct {
	grpcServer *grpc.Server       `validate:"required"`
	listener   net.Listener       `validate:"required"`
	config     *sophrosyne.Config `validate:"required"`
	logger     *slog.Logger       `validate:"required"`
	validator  sophrosyne.Validator
}

func NewServer(ctx context.Context, opts ...Option) (*Server, error) {
	s := &Server{}
	setOptions(s, defaultServerOptions(), opts...)

	err := s.validator.Validate(s)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// Serve starts the server. It is a wrapper around [grpc.Server.Serve].
func (s Server) Serve() error {
	s.logger.InfoContext(context.Background(), "starting server", "port", strings.Split(s.listener.Addr().String(), ":")[1])
	return s.grpcServer.Serve(s.listener)
}

// GracefulStop stops the server gracefully. It is a wrapper around
// [grpc.Server.GracefulStop].
func (s Server) GracefulStop() {
	s.grpcServer.GracefulStop()
}

// RegisterService registers the service with the gRPC server. It is a
// wrapper around [grpc.Server.RegisterService].
func (s Server) RegisterService(desc *grpc.ServiceDesc, ss interface{}) {
	s.grpcServer.RegisterService(desc, ss)
}

// GetServiceInfo returns the service info for the gRPC server. It is a
// wrapper around [grpc.Server.GetServiceInfo].
func (s Server) GetServiceInfo() map[string]grpc.ServiceInfo {
	return s.grpcServer.GetServiceInfo()
}

// defaultServerOptions returns a set of default ServerOption functions.
//
// The default ServerOption functions are:
// - WithServerValidator using [Validator.NewValidator()].
func defaultServerOptions() []Option {
	return []Option{
		WithValidator(validator.NewValidator()),
	}
}
