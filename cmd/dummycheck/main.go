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
	"fmt"
	"log"
	"net"
	"os"

	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"

	v0 "github.com/madsrc/sophrosyne/internal/grpc/sophrosyne/v0"
)

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "port",
				Usage: "port to listen on",
				Value: 11432,
			},
		},
		Action: func(c *cli.Context) error {
			log.Printf("starting server on port %d\n", c.Int("port"))
			lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", c.Int("port")))
			if err != nil {
				log.Fatalf("failed to listen: %v", err)
			}
			var opts []grpc.ServerOption
			grpcServer := grpc.NewServer(opts...)
			v0.RegisterCheckProviderServiceServer(grpcServer, checkServer{})
			err = grpcServer.Serve(lis)
			if err != nil {
				log.Fatalf("failed to serve: %v", err)
			}

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}

type checkServer struct {
	v0.UnimplementedCheckProviderServiceServer
}

func (c checkServer) Check(ctx context.Context, request *v0.CheckProviderRequest) (*v0.CheckProviderResponse, error) {
	var cnt any
	switch request.GetCheck().(type) {
	case *v0.CheckProviderRequest_Text:
		cnt = request.GetText()
	case *v0.CheckProviderRequest_Image:
		cnt = request.GetImage()
	default:
		cnt = ""
	}
	if cnt == "false" {
		return &v0.CheckProviderResponse{
			Result: false,
		}, nil
	}
	return &v0.CheckProviderResponse{
		Result: true,
	}, nil
}
