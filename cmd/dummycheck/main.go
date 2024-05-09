package main

import (
	"context"
	"fmt"
	"github.com/madsrc/sophrosyne/internal/grpc/checks"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 11432))
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
