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

package interceptors

import (
	"context"
	"encoding/base64"
	"log/slog"
	"strings"

	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/madsrc/sophrosyne"
	"github.com/madsrc/sophrosyne/internal/grpc"
)

func tokenFromMetadata(md metadata.MD, logger *slog.Logger) []byte {
	logger.DebugContext(context.Background(), "extracting token from metadata", "metadata", md)
	if len(md.Get("authorization")) == 0 {
		logger.DebugContext(context.Background(), "no token provided")
		return []byte{}
	}
	token, err := base64.StdEncoding.DecodeString(
		strings.TrimPrefix(md.Get("authorization")[0], "Bearer "),
	)
	if err != nil {
		logger.InfoContext(context.Background(), "unable to decode token", "error", err)
		return []byte{}
	}

	logger.DebugContext(context.Background(), "token extracted", "token", token)

	return token
}

var MissingMetadata = status.Errorf(codes.InvalidArgument, "no metadata provided")

func EnsureValidTokenUnary(userService sophrosyne.UserService, logger *slog.Logger, config *sophrosyne.Config) googlegrpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *googlegrpc.UnaryServerInfo, handler googlegrpc.UnaryHandler) (any, error) {
		logger.InfoContext(ctx, "ensuring valid token - unary")
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, MissingMetadata
		}

		user, err := userService.GetUserByToken(ctx, sophrosyne.ProtectToken(tokenFromMetadata(md, logger), config))
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, grpc.InvalidTokenMsg)
		}

		logger.InfoContext(ctx, "valid token", "user_id", user.ID)
		ctx = context.WithValue(ctx, sophrosyne.UserContextKey{}, &user)
		// Continue execution of handler after ensuring a valid token.
		return handler(ctx, req)
	}
}

func EnsureValidTokenStream(userService sophrosyne.UserService, logger *slog.Logger, config *sophrosyne.Config) googlegrpc.StreamServerInterceptor {
	return func(srv any, ss googlegrpc.ServerStream, info *googlegrpc.StreamServerInfo, handler googlegrpc.StreamHandler) error {
		logger.InfoContext(ss.Context(), "ensuring valid token - stream")
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			logger.InfoContext(ss.Context(), "no metadata provided")
			return MissingMetadata
		}

		user, err := userService.GetUserByToken(ss.Context(), sophrosyne.ProtectToken(tokenFromMetadata(md, logger), config))
		if err != nil {
			logger.InfoContext(ss.Context(), "unable to get user by token", "error", err)
			return status.Errorf(codes.Unauthenticated, grpc.InvalidTokenMsg)
		}

		logger.InfoContext(ss.Context(), "valid token", "user_id", user.ID)
		ctx := context.WithValue(ss.Context(), sophrosyne.UserContextKey{}, &user)

		type serverStream struct {
			googlegrpc.ServerStream
			ctx context.Context
		}
		// Continue execution of handler after ensuring a valid token.
		return handler(srv, &serverStream{
			ServerStream: ss,
			ctx:          ctx,
		})
	}
}
