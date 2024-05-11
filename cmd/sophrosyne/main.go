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
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	http2 "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	"github.com/madsrc/sophrosyne"
	"github.com/madsrc/sophrosyne/internal/cedar"
	"github.com/madsrc/sophrosyne/internal/configProvider"
	"github.com/madsrc/sophrosyne/internal/healthchecker"
	"github.com/madsrc/sophrosyne/internal/http"
	"github.com/madsrc/sophrosyne/internal/http/middleware"
	"github.com/madsrc/sophrosyne/internal/migrate"
	"github.com/madsrc/sophrosyne/internal/otel"
	"github.com/madsrc/sophrosyne/internal/pgx"
	"github.com/madsrc/sophrosyne/internal/rpc"
	"github.com/madsrc/sophrosyne/internal/rpc/services"
	"github.com/madsrc/sophrosyne/internal/tls"
	"github.com/madsrc/sophrosyne/internal/validator"
)

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		_, _ = fmt.Fprintf(c.App.Writer, "v%s\n", c.App.Version)
	}
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "print the version",
	}
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Usage: "The path to the configuration file",
				Value: "config.yaml",
			},
			&cli.StringSliceFlag{
				Name:  "secretfiles",
				Usage: "Files to read individual configuration values from. Multiple files can be specified by separating them with a comma or supply the option multiple times. The name of the file is used to determine what configuration parameter the content of the file will be read in to. For example, a file called 'database.host' will have its content used as the the value for 'database.host' in the configuration. This option is recommended to be used for secrets.",
				Value: nil,
			},
		},
		Version: "0.0.0",
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "sophrosyne",
				Action: func(c *cli.Context) error {
					return run(c)
				},
			},
			{
				Name:  "version",
				Usage: "print the version",
				Action: func(c *cli.Context) error {
					cli.VersionPrinter(c)
					return nil
				},
			},
			{
				Name:  "migrate",
				Usage: "migrate the database to the latest version",
				Action: func(c *cli.Context) error {
					validate := validator.NewValidator()

					config, err := getConfig(c.String("config"), nil, c.StringSlice("secretfiles"), validate)
					migrationService, err := migrate.NewMigrationService(config)
					if err != nil {
						return err
					}

					err = migrationService.Up()
					if err != nil {
						if !errors.Is(err, migrate.ErrNoChange) {
							return err
						} else {
							_, _ = fmt.Fprintf(c.App.Writer, "No migrations to apply")
							return nil
						}
					}
					v, dirty, err := migrationService.Versions()
					if err != nil {
						return err
					}
					msg := fmt.Sprintf("Migrations applied. Database at version '%d'", v)
					if dirty {
						msg = fmt.Sprintf("%s (dirty)\n", msg)
					} else {
						msg = fmt.Sprintf("%s\n", msg)
					}
					_, _ = fmt.Fprintf(c.App.Writer, msg)
					return nil
				},
			},
			{
				Name:  "config",
				Usage: "show the current configuration",
				Action: func(c *cli.Context) error {
					validate := validator.NewValidator()
					config, err := getConfig(c.String("config"), nil, c.StringSlice("secretfiles"), validate)
					if err != nil {
						return err
					}

					dat, err := yaml.Marshal(config)
					if err != nil {
						return err
					}

					_, _ = fmt.Fprintf(c.App.Writer, "%s\n", dat)
					return nil
				},
			},
			{
				Name:  "healthcheck",
				Usage: "check if the server is running",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "target",
						Usage: "target server address. Must include scheme and port number",
						Value: "https://127.0.0.1:8080/healthz",
					},
					&cli.BoolFlag{
						Name:  "insecure-skip-verify",
						Usage: "Skip TLS certificate verification",
						Value: false,
					},
				},
				Action: func(c *cli.Context) error {
					validate := validator.NewValidator()
					config, err := getConfig(c.String("config"), map[string]interface{}{
						"security.tls.insecureSkipVerify": c.Bool("insecure-skip-verify"),
					}, c.StringSlice("secretfiles"), validate)
					if err != nil {
						return err
					}

					tlsConfig, err := tls.NewTLSClientConfig(config)
					if err != nil {
						return err
					}
					client := http2.Client{
						Timeout: 5 * time.Second,
						Transport: &http2.Transport{
							TLSClientConfig: tlsConfig,
						},
					}
					resp, err := client.Get(c.String("target"))
					if err != nil {
						if errors.Is(err, syscall.ECONNREFUSED) {
							return cli.Exit("unhealthy", 2)
						}
						return cli.Exit(err.Error(), 1)
					}
					if resp.StatusCode == http2.StatusOK {
						return cli.Exit("healthy", 0)
					}
					return cli.Exit("unhealthy", 3)

				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func getConfig(filepath string, overwrites map[string]interface{}, secretfiles []string, validate *validator.Validator) (*sophrosyne.Config, error) {
	cp, err := configProvider.NewConfigProvider(
		filepath,
		overwrites,
		secretfiles,
		validate,
	)
	if err != nil {
		return nil, err
	}

	return cp.Get(), nil
}

func run(c *cli.Context) error {
	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	validate := validator.NewValidator()
	config, err := getConfig(c.String("config"), nil, c.StringSlice("secretfiles"), validate)
	if err != nil {
		return err
	}

	otelService, err := otel.NewOtelService()
	if err != nil {
		return err
	}

	logger := slog.New(sophrosyne.NewLogHandler(config, otelService))

	otelShutdown, err := otel.SetupOTelSDK(ctx, config)
	if err != nil {
		return err
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(ctx))
	}()

	migrationService, err := migrate.NewMigrationService(config)
	if err != nil {
		return err
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
		return sourceErr
	}
	if dbError != nil {
		return dbError
	}

	checkServiceDatabase, err := pgx.NewCheckService(ctx, config, logger)
	if err != nil {
		return err
	}

	checkService := sophrosyne.NewCheckServiceCache(config, checkServiceDatabase, otelService)

	profileServiceDatabase, err := pgx.NewProfileService(ctx, config, logger, checkService)
	if err != nil {
		return err
	}

	userServiceDatabase, err := pgx.NewUserService(ctx, config, logger, rand.Reader, profileServiceDatabase)
	if err != nil {
		return err
	}

	userService := sophrosyne.NewUserServiceCache(config, userServiceDatabase, otelService)
	if err != nil {
		return err
	}

	profileService := sophrosyne.NewProfileServiceCache(config, profileServiceDatabase, otelService)
	if err != nil {
		return err
	}

	authzProvider, err := cedar.NewAuthorizationProvider(ctx, logger, userService, otelService, profileService, checkService)

	rpcServer, err := rpc.NewRPCServer(logger)
	if err != nil {
		return err
	}

	rpcUserService, err := services.NewUserService(userService, authzProvider, logger, validate)
	if err != nil {
		return err
	}

	rpcCheckService, err := services.NewCheckService(checkService, authzProvider, logger, validate)
	if err != nil {
		return err
	}

	rpcProfileService, err := services.NewProfileService(profileService, authzProvider, logger, validate)
	if err != nil {
		return err
	}

	rpcScanService, err := services.NewScanService(authzProvider, logger, validate, profileService, checkService)
	if err != nil {
		return err
	}

	rpcServer.Register(rpcUserService.EntityID(), rpcUserService)
	rpcServer.Register(rpcCheckService.EntityID(), rpcCheckService)
	rpcServer.Register(rpcProfileService.EntityID(), rpcProfileService)
	rpcServer.Register(rpcScanService.EntityID(), rpcScanService)

	tlsConfig, err := tls.NewTLSServerConfig(config, rand.Reader)

	healthcheckService, err := healthchecker.NewHealthcheckService(
		[]sophrosyne.HealthChecker{
			userService,
			userServiceDatabase,
		},
	)

	s, err := http.NewServer(ctx, config, validate, logger, otelService, userService, tlsConfig)
	if err != nil {
		return err
	}

	s.Handle(
		"/v1/rpc",
		middleware.PanicCatcher(
			logger,
			otelService,
			middleware.SetupTracing(
				otelService,
				middleware.RequestLogging(
					logger,
					middleware.Authentication(
						nil,
						config,
						userService,
						logger,
						http.RPCHandler(logger, rpcServer, config),
					),
				),
			),
		),
	)
	s.Handle(
		"/healthz",
		middleware.PanicCatcher(
			logger,
			otelService,
			middleware.SetupTracing(
				otelService,
				middleware.RequestLogging(
					logger,
					http.HealthcheckHandler(logger, healthcheckService),
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

		return err
	case <-ctx.Done():
		// Wait for first CTRL+C.
		// Stop receiving signal notifications as soon as possible.
		stop()
	}

	// When Shutdown is called, ListenAndServe immediately returns ErrServerClosed.
	err = s.Shutdown(context.Background())
	return err
}
