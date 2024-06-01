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
	"github.com/madsrc/sophrosyne/internal/validator"
	"google.golang.org/grpc"
	"log/slog"
	"net"
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

	if s.logger != nil {
		s.logger.DebugContext(ctx, "validating server options", "options", opts, "defaults", defaultServerOptions())
	}
	err := s.validator.Validate(s)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// Serve starts the server. It is a wrapper around [grpc.Server.Serve].
func (s Server) Serve() error {
	return s.grpcServer.Serve(s.listener)
}

// defaultServerOptions returns a set of default ServerOption functions.
//
// The default ServerOption functions are:
// - WithServerValidator using [validator.NewValidator()].
func defaultServerOptions() []Option {
	return []Option{
		WithValidator(validator.NewValidator()),
	}
}
