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

package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"log/slog"
	"net"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	mw "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"

	"github.com/madsrc/sophrosyne/internal/grpc/interceptors"
	v0 "github.com/madsrc/sophrosyne/internal/grpc/sophrosyne/v0"

	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/madsrc/sophrosyne"
	"github.com/madsrc/sophrosyne/internal/grpc"
)

func allButHealthZ(ctx context.Context, callMeta mw.CallMeta) bool {
	return healthpb.Health_ServiceDesc.ServiceName != callMeta.Service
}

// setupGRPCServer creates a new gRPC server and returns it.
// This function is used by the main function and should not be called directly.
func setupGRPCServer(ctx context.Context, config *sophrosyne.Config, logger *slog.Logger, services grpcServices, tlsConfig *tls.Config, validate sophrosyne.Validator, userService sophrosyne.UserService, metricService sophrosyne.MetricService) (*grpc.Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", config.Server.Port))
	if err != nil {
		return nil, err
	}

	grpcPanicRecoveryHandler := func(p any) (err error) {
		metricService.RecordPanic(ctx)
		return status.Errorf(codes.Internal, "%s", p)
	}

	grpcServer, err := grpc.NewServer(ctx, grpc.WithGrpcServer(
		googlegrpc.NewServer(
			googlegrpc.ChainUnaryInterceptor(
				otelgrpc.UnaryServerInterceptor(),
				logging.UnaryServerInterceptor(interceptors.Logger(logger)),
				selector.UnaryServerInterceptor(interceptors.EnsureValidTokenUnary(userService, logger, config), selector.MatchFunc(allButHealthZ)),
				recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
			),
			googlegrpc.ChainStreamInterceptor(
				otelgrpc.StreamServerInterceptor(),
				logging.StreamServerInterceptor(interceptors.Logger(logger)),
				selector.StreamServerInterceptor(interceptors.EnsureValidTokenStream(userService, logger, config), selector.MatchFunc(allButHealthZ)),
				recovery.StreamServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
			),
			googlegrpc.Creds(credentials.NewTLS(tlsConfig)),
		)),
		grpc.WithListener(listener),
		grpc.WithLogger(logger),
		grpc.WithConfig(config),
		grpc.WithValidator(validate),
	)
	if err != nil {
		return nil, err
	}

	v0.RegisterCheckServiceServer(grpcServer, services.grpcCheckServiceServer)
	v0.RegisterUserServiceServer(grpcServer, services.grpcUserServiceServer)
	v0.RegisterProfileServiceServer(grpcServer, services.grpcProfileServiceServer)
	v0.RegisterScanServiceServer(grpcServer, services.grpcScanServiceServer)
	reflection.Register(grpcServer)
	healthcheck := health.NewServer()
	healthgrpc.RegisterHealthServer(grpcServer, healthcheck)

	go func() {
		// TODO: Actually inspect each service and set the health accordingly.
		// asynchronously inspect dependencies and toggle serving status as needed

		for {
			healthcheck.SetServingStatus("", healthpb.HealthCheckResponse_SERVING) // empty string represents the health of the system

			time.Sleep(time.Second * 5)
		}
	}()

	return grpcServer, nil
}

type grpcServices struct {
	grpcUserServiceServer    *grpc.UserServiceServer
	grpcProfileServiceServer *grpc.ProfileServiceServer
	grpcCheckServiceServer   *grpc.CheckServiceServer
	grpcScanServiceServer    *grpc.ScanServiceServer
}

func createGRPCServices(
	ctx context.Context,
	config *sophrosyne.Config,
	logger *slog.Logger,
	validate sophrosyne.Validator,
	authzProvider sophrosyne.AuthorizationProvider,
	checkService sophrosyne.CheckService,
	profileService sophrosyne.ProfileService,
	userService sophrosyne.UserService,
) (*grpcServices, error) {
	ret := grpcServices{}
	var err error
	ret.grpcUserServiceServer, err = grpc.NewUserServiceServer(ctx,
		grpc.WithLogger(logger),
		grpc.WithConfig(config),
		grpc.WithValidator(validate),
		grpc.WithUserService(userService),
		grpc.WithAuthorizationProvider(authzProvider))
	if err != nil {
		return nil, err
	}
	ret.grpcProfileServiceServer, err = grpc.NewProfileServiceServer(ctx,
		grpc.WithLogger(logger),
		grpc.WithConfig(config),
		grpc.WithValidator(validate),
		grpc.WithProfileService(profileService),
		grpc.WithAuthorizationProvider(authzProvider))
	if err != nil {
		return nil, err
	}
	ret.grpcCheckServiceServer, err = grpc.NewCheckServiceServer(ctx,
		grpc.WithLogger(logger),
		grpc.WithConfig(config),
		grpc.WithValidator(validate),
		grpc.WithCheckService(checkService),
		grpc.WithAuthorizationProvider(authzProvider))
	if err != nil {
		return nil, err
	}
	ret.grpcScanServiceServer, err = grpc.NewScanServiceServer(ctx,
		grpc.WithLogger(logger),
		grpc.WithConfig(config),
		grpc.WithValidator(validate),
		grpc.WithProfileService(profileService))
	if err != nil {
		return nil, err
	}

	return &ret, nil
}
