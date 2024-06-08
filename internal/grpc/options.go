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
	"log/slog"
	"net"

	"google.golang.org/grpc"

	"github.com/madsrc/sophrosyne"
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
// provided Logger to instance.
//
// The Option function can only be applied to the following types:
// - *ScanServiceServer
// - *Server
// - *UserServiceServer
// - *CheckServiceServer
// - *ProfileServiceServer
//
// If the type is not one of the above, the Option function does nothing.
func WithLogger(logger *slog.Logger) Option {
	return func(target any) {
		switch s := target.(type) {
		case *ScanServiceServer:
			s.Logger = logger
		case *Server:
			s.logger = logger
		case *UserServiceServer:
			s.logger = logger
		case *CheckServiceServer:
			s.logger = logger
		case *ProfileServiceServer:
			s.logger = logger
		default:
			return
		}
	}
}

// WithConfig returns an Option function that sets the
// provided Config to instance.
//
// The Option function can only be applied to the following types:
// - *ScanServiceServer
// - *Server
// - *UserServiceServer
// - *CheckServiceServer
// - *ProfileServiceServer
//
// If the type is not one of the above, the Option function does nothing.
func WithConfig(config *sophrosyne.Config) Option {
	return func(target any) {
		switch s := target.(type) {
		case *ScanServiceServer:
			s.Config = config
		case *Server:
			s.config = config
		case *UserServiceServer:
			s.config = config
		case *CheckServiceServer:
			s.config = config
		case *ProfileServiceServer:
			s.config = config
		default:
			return
		}
	}
}

// WithValidator returns an Option function that sets the
// provided Validator to instance.
//
// The Option function can only be applied to the following types:
// - *ScanServiceServer
// - *Server
// - *UserServiceServer
// - *CheckServiceServer
// - *ProfileServiceServer
//
// If the type is not one of the above, the Option function does nothing.
func WithValidator(validator sophrosyne.Validator) Option {
	return func(target any) {
		switch s := target.(type) {
		case *ScanServiceServer:
			s.Validator = validator
		case *Server:
			s.validator = validator
		case *UserServiceServer:
			s.validator = validator
		case *CheckServiceServer:
			s.validator = validator
		case *ProfileServiceServer:
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
// provided ProfileService to instance.
//
// The Option function can only be applied to the following types:
// - *ScanServiceServer
// - *ProfileServiceServer
//
// If the type is not one of the above, the Option function does nothing.
func WithProfileService(profileService sophrosyne.ProfileService) Option {
	return func(target any) {
		switch s := target.(type) {
		case *ScanServiceServer:
			s.ProfileService = profileService
		case *ProfileServiceServer:
			s.profileService = profileService
		default:
			return
		}
	}
}

// WithCheckService returns an Option function that sets the
// provided checkService to instance.
//
// The Option function can only be applied to the following types:
// - *CheckServiceServer
//
// If the type is not one of the above, the Option function does nothing.
func WithCheckService(checkService sophrosyne.CheckService) Option {
	return func(target any) {
		switch s := target.(type) {
		case *CheckServiceServer:
			s.checkService = checkService
		default:
			return
		}
	}
}

// WithUserService returns an Option function that sets the
// provided userService to instance.
//
// The Option function can only be applied to the following types:
// - *UserServiceServer
//
// If the type is not one of the above, the Option function does nothing.
func WithUserService(userService sophrosyne.UserService) Option {
	return func(target any) {
		switch s := target.(type) {
		case *UserServiceServer:
			s.userService = userService
		default:
			return
		}
	}
}

// WithAuthorizationProvider returns an Option function that sets the
// provided authorizationProvider to instance.
//
// The Option function can only be applied to the following types:
// - *UserServiceServer
// - *CheckServiceServer
// - *ProfileServiceServer
//
// If the type is not one of the above, the Option function does nothing.
func WithAuthorizationProvider(authorizationProvider sophrosyne.AuthorizationProvider) Option {
	return func(target any) {
		switch s := target.(type) {
		case *UserServiceServer:
			s.authzProvider = authorizationProvider
		case *CheckServiceServer:
			s.authzProvider = authorizationProvider
		case *ProfileServiceServer:
			s.authzProvider = authorizationProvider
		default:
			return
		}
	}
}
