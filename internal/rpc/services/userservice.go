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

	"github.com/madsrc/sophrosyne"
	"github.com/madsrc/sophrosyne/internal/rpc"
	"github.com/madsrc/sophrosyne/internal/rpc/internal/jsonrpc"
)

type UserService struct {
	methods     map[jsonrpc.Method]rpc.Method
	userService sophrosyne.UserService
	authz       sophrosyne.AuthorizationProvider
	logger      *slog.Logger
	validator   sophrosyne.Validator
}

func NewUserService(userService sophrosyne.UserService, authz sophrosyne.AuthorizationProvider, logger *slog.Logger, validator sophrosyne.Validator) (*UserService, error) {
	u := &UserService{
		methods:     make(map[jsonrpc.Method]rpc.Method),
		userService: userService,
		authz:       authz,
		logger:      logger,
		validator:   validator,
	}

	u.methods["Users::GetUser"] = getUser{service: u}
	u.methods["Users::GetUsers"] = getUsers{service: u}
	u.methods["Users::CreateUser"] = createUser{service: u}
	u.methods["Users::UpdateUser"] = updateUser{service: u}
	u.methods["Users::DeleteUser"] = deleteUser{service: u}
	u.methods["Users::RotateToken"] = rotateToken{service: u}

	return u, nil
}

func (u UserService) EntityType() string {
	return "Service"
}

func (u UserService) EntityID() string {
	return "Users"
}

func (u UserService) InvokeMethod(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	return invokeMethod(ctx, u.logger, u.methods, req)
}

const userNotFoundError = "user not found"

type getUser struct {
	service *UserService
}

func (u getUser) GetService() rpc.Service {
	return u.service
}

func (u getUser) EntityType() string {
	return "Users"
}

func (u getUser) EntityID() string {
	return "GetUser"
}

func (u getUser) Invoke(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.GetUserRequest
	err := rpc.ParamsIntoAny(&req, &params, u.service.validator)
	if err != nil {
		u.service.logger.ErrorContext(ctx, paramExtractError, "error", err)
		return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
	}

	if params.Email != "" {
		u, _ := u.service.userService.GetUserByEmail(ctx, params.Email)
		params.ID = u.ID
	}
	if params.Name != "" {
		u, _ := u.service.userService.GetUserByName(ctx, params.Name)
		params.ID = u.ID
	}

	curUser := sophrosyne.ExtractUser(ctx)
	if curUser == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	if !u.service.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curUser,
		Action:    u,
		Resource:  sophrosyne.User{ID: params.ID},
	}) {
		return rpc.ErrorFromRequest(&req, 12345, "unauthorized")
	}

	user, err := u.service.userService.GetUser(ctx, params.ID)
	if err != nil {
		u.service.logger.ErrorContext(ctx, "unable to get user", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, userNotFoundError)
	}

	resp := sophrosyne.GetUserResponse{}

	return rpc.ResponseToRequest(&req, resp.FromUser(user))
}

type getUsers struct {
	service *UserService
}

func (u getUsers) GetService() rpc.Service {
	return u.service
}

func (u getUsers) EntityType() string {
	return "Users"
}

func (u getUsers) EntityID() string {
	return "GetUsers"
}

func (u getUsers) Invoke(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.GetUsersRequest
	err := rpc.ParamsIntoAny(&req, &params, u.service.validator)
	if err != nil {
		if errors.Is(err, rpc.NoParamsError) {
			params = sophrosyne.GetUsersRequest{}
		} else {
			u.service.logger.ErrorContext(ctx, paramExtractError, "error", err)
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
			u.service.logger.ErrorContext(ctx, "unable to decode cursor", "error", err)
			return rpc.ErrorFromRequest(&req, 12347, "invalid cursor")
		}
	} else {
		cursor = sophrosyne.NewDatabaseCursor(curUser.ID, "")
	}

	users, err := u.service.userService.GetUsers(ctx, cursor)
	if err != nil {
		u.service.logger.ErrorContext(ctx, "unable to get users", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, "users not found")
	}

	var usersResponse []sophrosyne.GetUserResponse
	for _, uu := range users {
		ok := u.service.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
			Principal: curUser,
			Action:    u,
			Resource:  sophrosyne.User{ID: uu.ID},
		})
		if ok {
			ent := &sophrosyne.GetUserResponse{}
			usersResponse = append(usersResponse, *ent.FromUser(uu))
		}
	}

	u.service.logger.DebugContext(ctx, "returning users", "total", len(usersResponse), "users", usersResponse)
	return rpc.ResponseToRequest(&req, sophrosyne.GetUsersResponse{
		Users:  usersResponse,
		Cursor: cursor.String(),
		Total:  len(usersResponse),
	})
}

type createUser struct {
	service *UserService
}

func (u createUser) GetService() rpc.Service {
	return u.service
}

func (u createUser) EntityType() string {
	return "Users"
}

func (u createUser) EntityID() string {
	return "CreateUser"
}

func (u createUser) Invoke(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.CreateUserRequest
	err := rpc.ParamsIntoAny(&req, &params, u.service.validator)
	if err != nil {
		u.service.logger.ErrorContext(ctx, paramExtractError, "error", err)
		return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
	}

	curUser := sophrosyne.ExtractUser(ctx)
	if curUser == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	ok := u.service.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curUser,
		Action:    u,
	})

	if !ok {
		return rpc.ErrorFromRequest(&req, 12345, "unauthorized")
	}

	user, err := u.service.userService.CreateUser(ctx, params)
	if err != nil {
		u.service.logger.ErrorContext(ctx, "unable to create user", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, "unable to create user")
	}

	resp := sophrosyne.CreateUserResponse{}
	return rpc.ResponseToRequest(&req, resp.FromUser(user))
}

type updateUser struct {
	service *UserService
}

func (u updateUser) GetService() rpc.Service {
	return u.service
}

func (u updateUser) EntityType() string {
	return "Users"
}

func (u updateUser) EntityID() string {
	return "CreateUser"
}

func (u updateUser) Invoke(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.UpdateUserRequest
	err := rpc.ParamsIntoAny(&req, &params, u.service.validator)
	if err != nil {
		u.service.logger.ErrorContext(ctx, paramExtractError, "error", err)
		return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
	}

	curUser := sophrosyne.ExtractUser(ctx)
	if curUser == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	userToUpdate, err := u.service.userService.GetUserByName(ctx, params.Name)
	if err != nil {
		return rpc.ErrorFromRequest(&req, 12346, userNotFoundError)
	}

	ok := u.service.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curUser,
		Action:    u,
		Resource:  sophrosyne.User{ID: userToUpdate.ID},
	})

	if !ok {
		return rpc.ErrorFromRequest(&req, 12345, "unauthorized")
	}

	user, err := u.service.userService.UpdateUser(ctx, params)
	if err != nil {
		u.service.logger.ErrorContext(ctx, "unable to update user", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, "unable to update user")
	}

	resp := &sophrosyne.UpdateUserResponse{}
	return rpc.ResponseToRequest(&req, resp.FromUser(user))
}

type deleteUser struct {
	service *UserService
}

func (u deleteUser) GetService() rpc.Service {
	return u.service
}

func (u deleteUser) EntityType() string {
	return "Users"
}

func (u deleteUser) EntityID() string {
	return "CreateUser"
}

func (u deleteUser) Invoke(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.DeleteUserRequest
	err := rpc.ParamsIntoAny(&req, &params, u.service.validator)
	if err != nil {
		u.service.logger.ErrorContext(ctx, paramExtractError, "error", err)
		return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
	}

	curUser := sophrosyne.ExtractUser(ctx)
	if curUser == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	userToDelete, err := u.service.userService.GetUserByName(ctx, params.Name)
	if err != nil {
		return rpc.ErrorFromRequest(&req, 12346, userNotFoundError)
	}

	ok := u.service.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curUser,
		Action:    u,
		Resource:  sophrosyne.User{ID: userToDelete.ID},
	})

	if !ok {
		return rpc.ErrorFromRequest(&req, 12345, "unauthorized")
	}

	err = u.service.userService.DeleteUser(ctx, userToDelete.Name)
	if err != nil {
		u.service.logger.ErrorContext(ctx, "unable to delete user", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, "unable to delete user")
	}

	return rpc.ResponseToRequest(&req, "ok")
}

type rotateToken struct {
	service *UserService
}

func (u rotateToken) GetService() rpc.Service {
	return u.service
}

func (u rotateToken) EntityType() string {
	return "Users"
}

func (u rotateToken) EntityID() string {
	return "CreateUser"
}

func (u rotateToken) Invoke(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.RotateTokenRequest
	err := rpc.ParamsIntoAny(&req, &params, u.service.validator)
	if err != nil {
		u.service.logger.ErrorContext(ctx, paramExtractError, "error", err)
		return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
	}

	curUser := sophrosyne.ExtractUser(ctx)
	if curUser == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	userToRotate, err := u.service.userService.GetUserByName(ctx, params.Name)
	if err != nil {
		return rpc.ErrorFromRequest(&req, 12346, userNotFoundError)
	}

	ok := u.service.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curUser,
		Action:    u,
		Resource:  sophrosyne.User{ID: userToRotate.ID},
	})

	if !ok {
		return rpc.ErrorFromRequest(&req, 12345, "unauthorized")
	}

	token, err := u.service.userService.RotateToken(ctx, userToRotate.Name)
	if err != nil {
		u.service.logger.ErrorContext(ctx, "unable to rotate token", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, "unable to rotate token")
	}

	resp := &sophrosyne.RotateTokenResponse{}
	return rpc.ResponseToRequest(&req, resp.FromUser(sophrosyne.User{Token: token}))
}
