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
	"os"
	"os/signal"

	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/madsrc/sophrosyne/internal/cache"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	"github.com/madsrc/sophrosyne"
	"github.com/madsrc/sophrosyne/internal/cedar"
	"github.com/madsrc/sophrosyne/internal/configProvider"
	"github.com/madsrc/sophrosyne/internal/migrate"
	"github.com/madsrc/sophrosyne/internal/otel"
	"github.com/madsrc/sophrosyne/internal/pgx"
	"github.com/madsrc/sophrosyne/internal/tls"
	"github.com/madsrc/sophrosyne/internal/validator"
)

var (
	version = "0.0.0-dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		_, _ = fmt.Fprintf(c.App.Writer, "v%s\n", c.App.Version)
		if !c.Bool("verbose") {
			return
		}
		_, _ = fmt.Fprintf(c.App.Writer, "commit: %s\n", commit)
		_, _ = fmt.Fprintf(c.App.Writer, "date: %s\n", date)
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
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "If set, application will provide verbose outputs for commands that doesn't use the log.",
				Value: false,
			},
		},
		Version: version,
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
					if err != nil {
						return err
					}
					migrationService, err := migrate.NewMigrationService(config)
					if err != nil {
						return err
					}

					err = migrationService.Up()
					if err != nil {
						if !errors.Is(err, migrate.ErrNoChange) {
							return err
						} else {
							_, _ = fmt.Fprint(c.App.Writer, "No migrations to apply")
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
					_, _ = fmt.Fprint(c.App.Writer, msg)
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
						Value: "127.0.0.1:8080",
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

					options := []googlegrpc.DialOption{
						googlegrpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
					}
					if config.Security.TLS.InsecureSkipVerify {
						options = append(options, googlegrpc.WithTransportCredentials(insecure.NewCredentials()))
					}

					conn, err := googlegrpc.NewClient(c.String("target"), options...)
					if err != nil {
						return err
					}

					defer conn.Close()

					healthClient := healthpb.NewHealthClient(conn)

					resp, err := healthClient.Check(context.Background(), &healthpb.HealthCheckRequest{
						Service: "",
					})

					// TODO: Why does this trigger "panic: rpc error: code = Unavailable desc = connection error: desc = "error reading server preface: EOF" ?
					if err != nil {
						return err
					}

					if resp.GetStatus() == healthpb.HealthCheckResponse_SERVING {
						return cli.Exit("healthy", 0)
					}
					return cli.Exit("unhealthy", 1)

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

	checkService := cache.NewCheckServiceCache(config, checkServiceDatabase, otelService)

	profileServiceDatabase, err := pgx.NewProfileService(ctx, config, logger, checkService)
	if err != nil {
		return err
	}

	userServiceDatabase, err := pgx.NewUserService(ctx, config, logger, rand.Reader, profileServiceDatabase, nil)
	if err != nil {
		return err
	}

	userService := cache.NewUserServiceCache(config, userServiceDatabase, otelService)

	profileService := cache.NewProfileServiceCache(config, profileServiceDatabase, otelService)

	authzProvider, err := cedar.NewAuthorizationProvider(ctx, logger, userService, otelService, profileService, checkService)

	tlsConfig, err := tls.NewTLSServerConfig(config, rand.Reader)

	GRPCServices, err := createGRPCServices(ctx, config, logger, validate, authzProvider, checkService, profileService, userService)
	if err != nil {
		return err
	}

	grpcServer, err := setupGRPCServer(ctx, config, logger, *GRPCServices, tlsConfig, validate, userService, otelService)
	if err != nil {
		return err
	}

	srvErr := make(chan error, 1)
	go func() {
		srvErr <- grpcServer.Serve()
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

	grpcServer.GracefulStop()
	return nil
}
