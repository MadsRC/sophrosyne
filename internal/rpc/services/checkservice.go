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

package services

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/madsrc/sophrosyne/internal/rpc/jsonrpc"

	"github.com/madsrc/sophrosyne"
	"github.com/madsrc/sophrosyne/internal/rpc"
)

type CheckService struct {
	checkService sophrosyne.CheckService
	authz        sophrosyne.AuthorizationProvider
	logger       *slog.Logger
	validator    sophrosyne.Validator
}

func NewCheckService(checkService sophrosyne.CheckService, authz sophrosyne.AuthorizationProvider, logger *slog.Logger, validator sophrosyne.Validator) (*CheckService, error) {
	u := &CheckService{
		checkService: checkService,
		authz:        authz,
		logger:       logger,
		validator:    validator,
	}

	return u, nil
}

const paramExtractError = "error extracting params from request"
const checkNotFoundError = "check not found"

func (u CheckService) EntityType() string {
	return "Service"
}

func (u CheckService) EntityID() string {
	return "Checks"
}

func (u CheckService) InvokeMethod(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	m := strings.Split(string(req.Method), "::")
	if len(m) != 2 {
		u.logger.ErrorContext(ctx, "unreachable", "error", sophrosyne.NewUnreachableCodeError())
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}
	switch m[1] {
	case "Getcheck":
		return u.GetCheck(ctx, req)
	case "GetChecks":
		return u.GetChecks(ctx, req)
	case "CreateCheck":
		return u.CreateCheck(ctx, req)
	case "UpdateCheck":
		return u.UpdateCheck(ctx, req)
	case "DeleteCheck":
		return u.DeleteCheck(ctx, req)
	default:
		u.logger.DebugContext(ctx, "cannot invoke method", "method", req.Method)
		return rpc.ErrorFromRequest(&req, jsonrpc.MethodNotFound, string(jsonrpc.MethodNotFoundMessage))
	}
}

func (u CheckService) GetCheck(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.GetCheckRequest
	err := rpc.ParamsIntoAny(&req, &params, u.validator)
	if err != nil {
		u.logger.ErrorContext(ctx, paramExtractError, "error", err)
		return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
	}

	if params.Name != "" {
		u, _ := u.checkService.GetCheckByName(ctx, params.Name)
		params.ID = u.ID
	}

	curCheck := sophrosyne.ExtractUser(ctx)
	if curCheck == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	if !u.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curCheck,
		Action:    sophrosyne.AuthorizationAction("GetCheck"),
		Resource:  sophrosyne.Check{ID: params.ID},
	}) {
		return rpc.ErrorFromRequest(&req, 12345, "unauthorized")
	}

	check, err := u.checkService.GetCheck(ctx, params.ID)
	if err != nil {
		u.logger.ErrorContext(ctx, "unable to get check", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, checkNotFoundError)
	}

	resp := sophrosyne.GetCheckResponse{}

	return rpc.ResponseToRequest(&req, resp.FromCheck(check))
}

func (u CheckService) GetChecks(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.GetChecksRequest
	err := rpc.ParamsIntoAny(&req, &params, u.validator)
	if err != nil {
		if errors.Is(err, rpc.ErrNoParams) {
			params = sophrosyne.GetChecksRequest{}
		} else {
			u.logger.ErrorContext(ctx, paramExtractError, "error", err)
			return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
		}
	}

	curCheck := sophrosyne.ExtractUser(ctx)
	if curCheck == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	var cursor *sophrosyne.DatabaseCursor
	if params.Cursor != "" {
		cursor, err = sophrosyne.DecodeDatabaseCursorWithOwner(params.Cursor, curCheck.ID)
		if err != nil {
			u.logger.ErrorContext(ctx, "unable to decode cursor", "error", err)
			return rpc.ErrorFromRequest(&req, 12347, "invalid cursor")
		}
	} else {
		cursor = sophrosyne.NewDatabaseCursor(curCheck.ID, "")
	}

	checks, err := u.checkService.GetChecks(ctx, cursor)
	if err != nil {
		u.logger.ErrorContext(ctx, "unable to get checks", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, "checks not found")
	}

	var checksResponse []sophrosyne.GetCheckResponse
	for _, uu := range checks {
		ok := u.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
			Principal: curCheck,
			Action:    sophrosyne.AuthorizationAction("GetChecks"),
			Resource:  sophrosyne.Check{ID: uu.ID},
		})
		if ok {
			ent := &sophrosyne.GetCheckResponse{}
			checksResponse = append(checksResponse, *ent.FromCheck(uu))
		}
	}

	u.logger.DebugContext(ctx, "returning checks", "total", len(checksResponse), "checks", checksResponse)
	return rpc.ResponseToRequest(&req, sophrosyne.GetChecksResponse{
		Checks: checksResponse,
		Cursor: cursor.String(),
		Total:  len(checksResponse),
	})
}

func (u CheckService) CreateCheck(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.CreateCheckRequest
	err := rpc.ParamsIntoAny(&req, &params, u.validator)
	if err != nil {
		u.logger.ErrorContext(ctx, paramExtractError, "error", err)
		return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
	}

	curCheck := sophrosyne.ExtractUser(ctx)
	if curCheck == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	ok := u.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curCheck,
		Action:    sophrosyne.AuthorizationAction("CreateCheck"),
	})

	if !ok {
		return rpc.ErrorFromRequest(&req, 12345, "unauthorized")
	}

	check, err := u.checkService.CreateCheck(ctx, params)
	if err != nil {
		u.logger.ErrorContext(ctx, "unable to create check", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, "unable to create check")
	}

	resp := sophrosyne.CreateCheckResponse{}
	return rpc.ResponseToRequest(&req, resp.FromCheck(check))
}

func (u CheckService) UpdateCheck(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.UpdateCheckRequest
	err := rpc.ParamsIntoAny(&req, &params, u.validator)
	if err != nil {
		u.logger.ErrorContext(ctx, paramExtractError, "error", err)
		return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
	}

	curCheck := sophrosyne.ExtractUser(ctx)
	if curCheck == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	checkToUpdate, err := u.checkService.GetCheckByName(ctx, params.Name)
	if err != nil {
		return rpc.ErrorFromRequest(&req, 12346, checkNotFoundError)
	}

	ok := u.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curCheck,
		Action:    sophrosyne.AuthorizationAction("UpdateCheck"),
		Resource:  sophrosyne.Check{ID: checkToUpdate.ID},
	})

	if !ok {
		return rpc.ErrorFromRequest(&req, 12345, "unauthorized")
	}

	check, err := u.checkService.UpdateCheck(ctx, params)
	if err != nil {
		u.logger.ErrorContext(ctx, "unable to update check", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, "unable to update check")
	}

	resp := &sophrosyne.UpdateCheckResponse{}
	return rpc.ResponseToRequest(&req, resp.FromCheck(check))
}

func (u CheckService) DeleteCheck(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.DeleteCheckRequest
	err := rpc.ParamsIntoAny(&req, &params, u.validator)
	if err != nil {
		u.logger.ErrorContext(ctx, paramExtractError, "error", err)
		return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
	}

	curCheck := sophrosyne.ExtractUser(ctx)
	if curCheck == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	checkToDelete, err := u.checkService.GetCheckByName(ctx, params.Name)
	if err != nil {
		return rpc.ErrorFromRequest(&req, 12346, checkNotFoundError)
	}

	ok := u.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curCheck,
		Action:    sophrosyne.AuthorizationAction("DeleteCheck"),
		Resource:  sophrosyne.Check{ID: checkToDelete.ID},
	})

	if !ok {
		return rpc.ErrorFromRequest(&req, 12345, "unauthorized")
	}

	err = u.checkService.DeleteCheck(ctx, checkToDelete.Name)
	if err != nil {
		u.logger.ErrorContext(ctx, "unable to delete check", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, "unable to delete check")
	}

	return rpc.ResponseToRequest(&req, "ok")
}
