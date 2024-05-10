// Sophrosyne
//
//	Copyright (C) 2024  Mads R. Havmand
//
// This program is free software: you can redistribute it and/or modify
//
//	it under the terms of the GNU Affero General Public License as published by
//	the Free Software Foundation, either version 3 of the License, or
//	(at your option) any later version.
//
//	This program is distributed in the hope that it will be useful,
//	but WITHOUT ANY WARRANTY; without even the implied warranty of
//	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//	GNU Affero General Public License for more details.
//
//	You should have received a copy of the GNU Affero General Public License
//	along with this program.  If not, see <http://www.gnu.org/licenses/>.
package services

import (
	"context"
	"log/slog"

	"github.com/madsrc/sophrosyne/internal/rpc"
	"github.com/madsrc/sophrosyne/internal/rpc/internal/jsonrpc"
)

func invokeMethod(ctx context.Context, logger *slog.Logger, methods map[jsonrpc.Method]rpc.Method, req jsonrpc.Request) ([]byte, error) {
	if methods[req.Method] == nil {
		logger.DebugContext(ctx, "cannot invoke method", "method", req.Method)
		return rpc.ErrorFromRequest(&req, jsonrpc.MethodNotFound, string(jsonrpc.MethodNotFoundMessage))
	}

	return methods[req.Method].Invoke(ctx, req)
}
