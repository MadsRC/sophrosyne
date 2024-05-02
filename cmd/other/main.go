package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/madsrc/sophrosyne/internal/cedar"
	"github.com/madsrc/sophrosyne/internal/configProvider"
	"github.com/madsrc/sophrosyne/internal/http"
	"github.com/madsrc/sophrosyne/internal/http/middleware"
	"github.com/madsrc/sophrosyne/internal/migrate"
	"github.com/madsrc/sophrosyne/internal/otel"
	"github.com/madsrc/sophrosyne/internal/pgx"
	"github.com/madsrc/sophrosyne/internal/rpc"
	"github.com/madsrc/sophrosyne/internal/rpc/services"
	"github.com/madsrc/sophrosyne/internal/tls"
	"github.com/madsrc/sophrosyne/internal/validator"
	"log/slog"
	"os"
	"os/signal"

	"github.com/madsrc/sophrosyne"
)

func main() {
	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	validate := validator.NewValidator()
	cp, err := configProvider.NewConfigProvider(
		"configurations/dev.yaml",
		nil,
		nil,
		validate,
	)
	if err != nil {
		panic(err)
	}

	config := cp.Get()

	otelService, err := otel.NewOtelService()
	if err != nil {
		panic(err)
	}

	logger := slog.New(sophrosyne.NewLogHandler(config, otelService))

	otelShutdown, err := otel.SetupOTelSDK(ctx, config)
	if err != nil {
		panic(err)
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(ctx))
	}()

	migrationService, err := migrate.NewMigrationService(config)
	if err != nil {
		panic(err)
	}

	logger.DebugContext(ctx, "Applying migrations")
	err = migrationService.Up()
	if err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			panic(err)
		} else {
			logger.DebugContext(ctx, "No migrations to apply")
		}
	}
	sourceErr, dbError := migrationService.Close()
	if sourceErr != nil {
		panic(sourceErr)
	}
	if dbError != nil {
		panic(dbError)
	}

	userServiceDatabase, err := pgx.NewUserService(ctx, config, logger, rand.Reader)
	if err != nil {
		panic(err)
	}

	userService := sophrosyne.NewUserServiceCache(config, userServiceDatabase, otelService)
	if err != nil {
		panic(err)
	}

	authzProvider, err := cedar.NewAuthorizationProvider(ctx, logger, userService, otelService)

	rpcServer, err := rpc.NewRPCServer(logger)
	if err != nil {
		panic(err)
	}

	rpcUserService, err := services.NewUserService(userService, authzProvider, logger, validate)
	if err != nil {
		panic(err)
	}

	rpcServer.Register(rpcUserService.EntityID(), rpcUserService)

	tlsConfig, err := tls.NewTLSConfig(config, rand.Reader)

	s, err := http.NewServer(ctx, config, validate, logger, otelService, userService, tlsConfig)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", config)

	s.Handle(
		"/v1/rpc",
		middleware.PanicCatcher(
			logger,
			otelService,
			middleware.SetupTracing(
				otelService,
				middleware.Authentication(
					nil,
					config,
					userService,
					logger,
					http.RPCHandler(logger, rpcServer),
				),
			),
		),
	)

	srvErr := make(chan error, 1)
	go func() {
		srvErr <- s.Start()
	}()

	// Wait for interruption.
	select {
	case err = <-srvErr:
		// Error when starting HTTP server.
		return
	case <-ctx.Done():
		// Wait for first CTRL+C.
		// Stop receiving signal notifications as soon as possible.
		stop()
	}

	// When Shutdown is called, ListenAndServe immediately returns ErrServerClosed.
	err = s.Shutdown(context.Background())
}
