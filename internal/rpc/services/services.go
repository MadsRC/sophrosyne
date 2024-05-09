package services

import (
	"context"
	"github.com/madsrc/sophrosyne/internal/rpc"
	"github.com/madsrc/sophrosyne/internal/rpc/internal/jsonrpc"
	"log/slog"
)

func invokeMethod(ctx context.Context, logger *slog.Logger, methods map[jsonrpc.Method]rpc.Method, req jsonrpc.Request) ([]byte, error) {
	if methods[req.Method] == nil {
		logger.DebugContext(ctx, "cannot invoke method", "method", req.Method)
		return rpc.ErrorFromRequest(&req, jsonrpc.MethodNotFound, string(jsonrpc.MethodNotFoundMessage))
	}

	return methods[req.Method].Invoke(ctx, req)
}
