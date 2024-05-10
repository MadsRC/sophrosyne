package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"

	"github.com/madsrc/sophrosyne/internal/grpc/checks"
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
			lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", c.Int("port")))
			if err != nil {
				log.Fatalf("failed to listen: %v", err)
			}
			var opts []grpc.ServerOption
			grpcServer := grpc.NewServer(opts...)
			checks.RegisterCheckServiceServer(grpcServer, checkServer{})
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
	checks.UnimplementedCheckServiceServer
}

func (c checkServer) Check(ctx context.Context, request *checks.CheckRequest) (*checks.CheckResponse, error) {
	var cnt string
	switch request.GetCheck().(type) {
	case *checks.CheckRequest_Text:
		cnt = request.GetText()
	case *checks.CheckRequest_Image:
		cnt = request.GetImage()
	default:
		cnt = ""
	}
	if cnt == "false" {
		return &checks.CheckResponse{
			Result:  false,
			Details: "this was false",
		}, nil
	}
	return &checks.CheckResponse{
		Result:  true,
		Details: "this was true",
	}, nil
}
