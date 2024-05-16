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

type UserService struct {
	userService sophrosyne.UserService
	authz       sophrosyne.AuthorizationProvider
	logger      *slog.Logger
	validator   sophrosyne.Validator
}

func NewUserService(userService sophrosyne.UserService, authz sophrosyne.AuthorizationProvider, logger *slog.Logger, validator sophrosyne.Validator) (*UserService, error) {
	u := &UserService{
		userService: userService,
		authz:       authz,
		logger:      logger,
		validator:   validator,
	}

	return u, nil
}

func (u UserService) EntityType() string {
	return "Service"
}

func (u UserService) EntityID() string {
	return "Users"
}

func (u UserService) InvokeMethod(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	m := strings.Split(string(req.Method), "::")
	if len(m) != 2 {
		u.logger.ErrorContext(ctx, "unreachable", "error", sophrosyne.NewUnreachableCodeError())
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}
	switch m[1] {
	case "GetUser":
		return u.GetUser(ctx, req)
	case "GetUsers":
		return u.GetUsers(ctx, req)
	case "CreateUser":
		return u.CreateUser(ctx, req)
	case "UpdateUser":
		return u.UpdateUser(ctx, req)
	case "DeleteUser":
		return u.DeleteUser(ctx, req)
	case "RotateToken":
		return u.RotateToken(ctx, req)
	default:
		u.logger.DebugContext(ctx, "cannot invoke method", "method", req.Method)
		return rpc.ErrorFromRequest(&req, jsonrpc.MethodNotFound, string(jsonrpc.MethodNotFoundMessage))
	}
}

const userNotFoundError = "user not found"

func (u UserService) GetUser(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.GetUserRequest
	err := rpc.ParamsIntoAny(&req, &params, u.validator)
	if err != nil {
		u.logger.ErrorContext(ctx, paramExtractError, "error", err)
		return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
	}

	if params.Email != "" {
		u, _ := u.userService.GetUserByEmail(ctx, params.Email)
		params.ID = u.ID
	}
	if params.Name != "" {
		u, _ := u.userService.GetUserByName(ctx, params.Name)
		params.ID = u.ID
	}

	curUser := sophrosyne.ExtractUser(ctx)
	if curUser == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	if !u.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curUser,
		Action:    sophrosyne.AuthorizationAction("GetUser"),
		Resource:  sophrosyne.User{ID: params.ID},
	}) {
		return rpc.ErrorFromRequest(&req, 12345, "unauthorized")
	}

	user, err := u.userService.GetUser(ctx, params.ID)
	if err != nil {
		u.logger.ErrorContext(ctx, "unable to get user", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, userNotFoundError)
	}

	resp := sophrosyne.GetUserResponse{}

	return rpc.ResponseToRequest(&req, resp.FromUser(user))
}

func (u UserService) GetUsers(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.GetUsersRequest
	err := rpc.ParamsIntoAny(&req, &params, u.validator)
	if err != nil {
		if errors.Is(err, rpc.ErrNoParams) {
			params = sophrosyne.GetUsersRequest{}
		} else {
			u.logger.ErrorContext(ctx, paramExtractError, "error", err)
			return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
		}
	}

	curUser := sophrosyne.ExtractUser(ctx)
	if curUser == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	var cursor *sophrosyne.DatabaseCursor
	if params.Cursor != "" {
		cursor, err = sophrosyne.DecodeDatabaseCursorWithOwner(params.Cursor, curUser.ID)
		if err != nil {
			u.logger.ErrorContext(ctx, "unable to decode cursor", "error", err)
			return rpc.ErrorFromRequest(&req, 12347, "invalid cursor")
		}
	} else {
		cursor = sophrosyne.NewDatabaseCursor(curUser.ID, "")
	}

	users, err := u.userService.GetUsers(ctx, cursor)
	if err != nil {
		u.logger.ErrorContext(ctx, "unable to get users", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, "users not found")
	}

	var usersResponse []sophrosyne.GetUserResponse
	for _, uu := range users {
		ok := u.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
			Principal: curUser,
			Action:    sophrosyne.AuthorizationAction("GetUsers"),
			Resource:  sophrosyne.User{ID: uu.ID},
		})
		if ok {
			ent := &sophrosyne.GetUserResponse{}
			usersResponse = append(usersResponse, *ent.FromUser(uu))
		}
	}

	u.logger.DebugContext(ctx, "returning users", "total", len(usersResponse), "users", usersResponse)
	return rpc.ResponseToRequest(&req, sophrosyne.GetUsersResponse{
		Users:  usersResponse,
		Cursor: cursor.String(),
		Total:  len(usersResponse),
	})
}

func (u UserService) CreateUser(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.CreateUserRequest
	err := rpc.ParamsIntoAny(&req, &params, u.validator)
	if err != nil {
		u.logger.ErrorContext(ctx, paramExtractError, "error", err)
		return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
	}

	curUser := sophrosyne.ExtractUser(ctx)
	if curUser == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	ok := u.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curUser,
		Action:    sophrosyne.AuthorizationAction("CreateUser"),
	})

	if !ok {
		return rpc.ErrorFromRequest(&req, 12345, "unauthorized")
	}

	user, err := u.userService.CreateUser(ctx, params)
	if err != nil {
		u.logger.ErrorContext(ctx, "unable to create user", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, "unable to create user")
	}

	resp := sophrosyne.CreateUserResponse{}
	return rpc.ResponseToRequest(&req, resp.FromUser(user))
}

func (u UserService) UpdateUser(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.UpdateUserRequest
	err := rpc.ParamsIntoAny(&req, &params, u.validator)
	if err != nil {
		u.logger.ErrorContext(ctx, paramExtractError, "error", err)
		return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
	}

	curUser := sophrosyne.ExtractUser(ctx)
	if curUser == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	userToUpdate, err := u.userService.GetUserByName(ctx, params.Name)
	if err != nil {
		return rpc.ErrorFromRequest(&req, 12346, userNotFoundError)
	}

	ok := u.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curUser,
		Action:    sophrosyne.AuthorizationAction("UpdateUser"),
		Resource:  sophrosyne.User{ID: userToUpdate.ID},
	})

	if !ok {
		return rpc.ErrorFromRequest(&req, 12345, "unauthorized")
	}

	user, err := u.userService.UpdateUser(ctx, params)
	if err != nil {
		u.logger.ErrorContext(ctx, "unable to update user", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, "unable to update user")
	}

	resp := &sophrosyne.UpdateUserResponse{}
	return rpc.ResponseToRequest(&req, resp.FromUser(user))
}

func (u UserService) DeleteUser(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.DeleteUserRequest
	err := rpc.ParamsIntoAny(&req, &params, u.validator)
	if err != nil {
		u.logger.ErrorContext(ctx, paramExtractError, "error", err)
		return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
	}

	curUser := sophrosyne.ExtractUser(ctx)
	if curUser == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	userToDelete, err := u.userService.GetUserByName(ctx, params.Name)
	if err != nil {
		return rpc.ErrorFromRequest(&req, 12346, userNotFoundError)
	}

	ok := u.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curUser,
		Action:    sophrosyne.AuthorizationAction("DeleteUser"),
		Resource:  sophrosyne.User{ID: userToDelete.ID},
	})

	if !ok {
		return rpc.ErrorFromRequest(&req, 12345, "unauthorized")
	}

	err = u.userService.DeleteUser(ctx, userToDelete.Name)
	if err != nil {
		u.logger.ErrorContext(ctx, "unable to delete user", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, "unable to delete user")
	}

	return rpc.ResponseToRequest(&req, "ok")
}

func (u UserService) RotateToken(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.RotateTokenRequest
	err := rpc.ParamsIntoAny(&req, &params, u.validator)
	if err != nil {
		u.logger.ErrorContext(ctx, paramExtractError, "error", err)
		return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
	}

	curUser := sophrosyne.ExtractUser(ctx)
	if curUser == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	userToRotate, err := u.userService.GetUserByName(ctx, params.Name)
	if err != nil {
		return rpc.ErrorFromRequest(&req, 12346, userNotFoundError)
	}

	ok := u.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curUser,
		Action:    sophrosyne.AuthorizationAction("RotateToken"),
		Resource:  sophrosyne.User{ID: userToRotate.ID},
	})

	if !ok {
		return rpc.ErrorFromRequest(&req, 12345, "unauthorized")
	}

	token, err := u.userService.RotateToken(ctx, userToRotate.Name)
	if err != nil {
		u.logger.ErrorContext(ctx, "unable to rotate token", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, "unable to rotate token")
	}

	resp := &sophrosyne.RotateTokenResponse{}
	return rpc.ResponseToRequest(&req, resp.FromUser(sophrosyne.User{Token: token}))
}
