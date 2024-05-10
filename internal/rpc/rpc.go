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
package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/madsrc/sophrosyne"
	"github.com/madsrc/sophrosyne/internal/rpc/internal/jsonrpc"
)

type Server struct {
	services map[string]Service
	logger   *slog.Logger
}

func NewRPCServer(logger *slog.Logger) (*Server, error) {
	return &Server{
		services: make(map[string]Service),
		logger:   logger,
	}, nil
}

func (s *Server) HandleRPCRequest(ctx context.Context, req []byte) ([]byte, error) {
	pReq := jsonrpc.Request{}
	err := pReq.UnmarshalJSON(req)
	if err != nil {
		return nil, err
	}

	svcName := strings.Split(string(pReq.Method), "::")[0]

	service, ok := s.services[svcName]
	if !ok {
		s.logger.InfoContext(ctx, "rpc service not found", "service", svcName, "method", pReq.Method)
		return ErrorFromRequest(&pReq, jsonrpc.MethodNotFound, string(jsonrpc.MethodNotFoundMessage))
	}
	data, err := service.InvokeMethod(ctx, pReq)
	if err != nil {
		s.logger.ErrorContext(ctx, "error handling rpc request", "error", err)
		return nil, err
	}

	return data, nil
}

func (s *Server) Register(name string, service Service) {
	s.services[name] = service
}

type Service interface {
	sophrosyne.AuthorizationEntity
	InvokeMethod(ctx context.Context, req jsonrpc.Request) ([]byte, error)
}

type Method interface {
	sophrosyne.AuthorizationEntity
	GetService() Service
	Invoke(ctx context.Context, req jsonrpc.Request) ([]byte, error)
}

func ErrorFromRequest(req *jsonrpc.Request, code jsonrpc.RPCErrorCode, message string) ([]byte, error) {
	return jsonrpc.Response{
		ID: req.ID,
		Error: &jsonrpc.Error{
			Code:    code,
			Message: message,
		},
	}.MarshalJSON()
}

func ResponseToRequest(req *jsonrpc.Request, result interface{}) ([]byte, error) {
	if req.IsNotification() {
		return nil, nil
	}
	return jsonrpc.Response{
		ID:     req.ID,
		Result: result,
	}.MarshalJSON()
}

func GetParams(req *jsonrpc.Request) (*jsonrpc.ParamsObject, *jsonrpc.ParamsArray, bool) {
	var ook bool
	po, ok := req.Params.(*jsonrpc.ParamsObject)
	if ok {
		ook = ok
	}
	pa, ok := req.Params.(*jsonrpc.ParamsArray)
	if ok {
		ook = ok
	}

	return po, pa, ook
}

var NoParamsError = fmt.Errorf("no params found")

func ParamsIntoAny(req *jsonrpc.Request, target any, validate sophrosyne.Validator) error {
	pa, po, ok := GetParams(req)
	if !ok {
		return NoParamsError
	}

	var b []byte
	var err error
	if pa != nil {
		b, err = json.Marshal(pa)
	}
	if po != nil {
		b, err = json.Marshal(po)
	}
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, &target)
	if err != nil {
		return err
	}

	if validate != nil {
		err = validate.Validate(target)
		if err != nil {
			return err
		}
	}

	vd, ok := target.(sophrosyne.Validator)
	if ok {
		return vd.Validate(nil)
	}

	return nil
}
