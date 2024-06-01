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
	"github.com/madsrc/sophrosyne"
	"google.golang.org/grpc"
	"log/slog"
	"net"
)

type Option func(s any)

// setOptions applies all provided Option functions to the provided instance.
func setOptions(target any, defaults []Option, opts ...Option) {
	for _, opt := range defaults {
		opt(target)
	}

	for _, opt := range opts {
		if opt != nil {
			opt(target)
		}
	}
}

// WithOptions returns an Option function that applies
// all provided Option functions to the provided instance.
func WithOptions(opts ...Option) Option {
	return func(target any) {
		if target == nil {
			return
		}
		for _, opt := range opts {
			if opt != nil {
				opt(target)
			}
		}
	}
}

// WithLogger returns an Option function that sets the
// provided logger to instance.
//
// The Option function can only be applied to the following types:
// - *ScanServiceServer
// - *Server
//
// If the type is not one of the above, the Option function does nothing.
func WithLogger(logger *slog.Logger) Option {
	return func(target any) {
		switch s := target.(type) {
		case *ScanServiceServer:
			s.logger = logger
		case *Server:
			s.logger = logger
		default:
			return
		}
	}
}

// WithConfig returns an Option function that sets the
// provided config to instance.
//
// The Option function can only be applied to the following types:
// - *ScanServiceServer
// - *Server
//
// If the type is not one of the above, the Option function does nothing.
func WithConfig(config *sophrosyne.Config) Option {
	return func(target any) {
		switch s := target.(type) {
		case *ScanServiceServer:
			s.config = config
		case *Server:
			s.config = config
		default:
			return
		}
	}
}

// WithValidator returns an Option function that sets the
// provided validator to instance.
//
// The Option function can only be applied to the following types:
// - *ScanServiceServer
// - *Server
//
// If the type is not one of the above, the Option function does nothing.
func WithValidator(validator sophrosyne.Validator) Option {
	return func(target any) {
		switch s := target.(type) {
		case *ScanServiceServer:
			s.validator = validator
		case *Server:
			s.validator = validator
		default:
			return
		}
	}
}

// WithListener returns an Option function that sets the
// provided listener to instance.
//
// The Option function can only be applied to the following types:
// - *Server
//
// If the type is not one of the above, the Option function does nothing.
func WithListener(listener net.Listener) Option {
	return func(target any) {
		switch s := target.(type) {
		case *Server:
			s.listener = listener
		default:
			return
		}
	}
}

// WithGrpcServer returns an Option function that sets the
// provided grpcServer to instance.
//
// The Option function can only be applied to the following types:
// - *Server
//
// If the type is not one of the above, the Option function does nothing.
func WithGrpcServer(grpcServer *grpc.Server) Option {
	return func(target any) {
		switch s := target.(type) {
		case *Server:
			s.grpcServer = grpcServer
		default:
			return
		}
	}
}

// WithProfileService returns an Option function that sets the
// provided profileService to instance.
//
// The Option function can only be applied to the following types:
// - *ScanServiceServer
//
// If the type is not one of the above, the Option function does nothing.
func WithProfileService(profileService sophrosyne.ProfileService) Option {
	return func(target any) {
		switch s := target.(type) {
		case *ScanServiceServer:
			s.profileService = profileService
		default:
			return
		}
	}
}
