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

package grpc

import (
	"context"
	"encoding/base64"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/madsrc/sophrosyne"
	v0 "github.com/madsrc/sophrosyne/internal/grpc/sophrosyne/v0"
	"github.com/madsrc/sophrosyne/internal/validator"
)

type UserServiceServer struct {
	v0.UnimplementedUserServiceServer
	logger        *slog.Logger                     `validate:"required"`
	config        *sophrosyne.Config               `validate:"required"`
	validator     sophrosyne.Validator             `validate:"required"`
	userService   sophrosyne.UserService           `validate:"required"`
	authzProvider sophrosyne.AuthorizationProvider `validate:"required"`
}

// NewUserServiceServer returns a new UserServiceServer instance.
//
// If the provided options are invalid, an error will be returned.
// Required options are marked with the 'validate:"required"' tag in
// the [UserServiceServer] struct. Every required option has a
// corresponding [Option] function.
//
// If no [sophrosyne.Validator] is provided, a default one will be
// created.
func NewUserServiceServer(ctx context.Context, opts ...Option) (*UserServiceServer, error) {
	s := &UserServiceServer{}
	setOptions(s, defaultUserServiceServerOptions(), opts...)

	if s.logger != nil {
		s.logger.DebugContext(ctx, "validating server options")
	}
	err := s.validator.Validate(s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func defaultUserServiceServerOptions() []Option {
	return []Option{
		WithValidator(validator.NewValidator()),
	}
}

func (s *UserServiceServer) GetUser(ctx context.Context, req *v0.GetUserRequest) (*v0.GetUserResponse, error) {
	curUser := sophrosyne.ExtractUser(ctx)
	if curUser == nil {
		return nil, status.Errorf(codes.Unauthenticated, InvalidTokenMsg)
	}

	var user sophrosyne.User
	var err error

	if req.GetId() != "" {
		user, err = s.userService.GetUser(ctx, req.GetId())
	} else if req.GetEmail() != "" {
		user, err = s.userService.GetUserByEmail(ctx, req.GetEmail())
	} else {
		user, err = s.userService.GetUserByName(ctx, req.GetName())
	}
	if err != nil {
		return nil, err
	}

	if !s.authzProvider.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curUser,
		Action:    sophrosyne.AuthorizationAction("GetUser"),
		Resource:  sophrosyne.User{ID: user.ID},
	}) {
		return nil, status.Errorf(codes.PermissionDenied, "unauthorized")
	}

	resp := &v0.GetUserResponse{
		Name:      user.Name,
		Email:     user.Email,
		IsAdmin:   user.IsAdmin,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}

	if user.DeletedAt != nil {
		resp.DeletedAt = timestamppb.New(*user.DeletedAt)
	}

	return resp, nil
}

func (s *UserServiceServer) GetUsers(ctx context.Context, req *v0.GetUsersRequest) (*v0.GetUsersResponse, error) {
	curUser := sophrosyne.ExtractUser(ctx)
	if curUser == nil {
		return nil, status.Errorf(codes.Unauthenticated, InvalidTokenMsg)
	}

	cursor := sophrosyne.NewDatabaseCursor(curUser.ID, "")
	var err error
	if req.GetCursor() != "" {
		cursor, err = sophrosyne.DecodeDatabaseCursorWithOwner(req.GetCursor(), curUser.ID)
		if err != nil {
			s.logger.ErrorContext(ctx, "unable to decode cursor", "error", err)
			return nil, status.Errorf(codes.InvalidArgument, InvalidCursorMsg)
		}
	}

	users, err := s.userService.GetUsers(ctx, cursor)
	if err != nil {
		s.logger.ErrorContext(ctx, "unable to get users", "error", err)
		return nil, status.Error(codes.Internal, "internal error getting users")
	}

	var usersResponse []*v0.GetUserResponse
	for _, u := range users {
		ok := s.authzProvider.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
			Principal: curUser,
			Action:    sophrosyne.AuthorizationAction("GetUsers"),
			Resource:  sophrosyne.User{ID: u.ID},
		})
		if ok {
			usersResponse = append(usersResponse, &v0.GetUserResponse{
				Name:      u.Name,
				Email:     u.Email,
				IsAdmin:   u.IsAdmin,
				CreatedAt: timestamppb.New(u.CreatedAt),
				UpdatedAt: timestamppb.New(u.UpdatedAt),
			})
			if u.DeletedAt != nil {
				usersResponse[len(usersResponse)-1].DeletedAt = timestamppb.New(*u.DeletedAt)
			}
		}
	}

	s.logger.DebugContext(ctx, "returning users", "total", len(usersResponse), "users", usersResponse)
	return &v0.GetUsersResponse{
		Users:  usersResponse,
		Cursor: cursor.String(),
		Total:  int32(len(usersResponse)),
	}, nil
}

func (s *UserServiceServer) CreateUser(ctx context.Context, req *v0.CreateUserRequest) (*v0.CreateUserResponse, error) {
	curUser := sophrosyne.ExtractUser(ctx)
	if curUser == nil {
		return nil, status.Errorf(codes.Unauthenticated, InvalidTokenMsg)
	}

	ok := s.authzProvider.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curUser,
		Action:    sophrosyne.AuthorizationAction("CreateUser"),
	})
	if !ok {
		return nil, status.Errorf(codes.PermissionDenied, "unauthorized")
	}

	user, err := s.userService.CreateUser(ctx, sophrosyne.CreateUserRequest{
		Name:    req.GetName(),
		Email:   req.GetEmail(),
		IsAdmin: req.GetIsAdmin(),
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "unable to create user", "error", err)
		return nil, err
	}

	s.logger.InfoContext(ctx, "created user", "user_id", user.ID)
	return &v0.CreateUserResponse{
		Name:      user.Name,
		Email:     user.Email,
		IsAdmin:   user.IsAdmin,
		Token:     base64.StdEncoding.EncodeToString(user.Token),
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}, nil
}

func (s *UserServiceServer) UpdateUser(ctx context.Context, req *v0.UpdateUserRequest) (*v0.UpdateUserResponse, error) {
	targetUser, err := getTargetUser(ctx, req.GetName(), s.userService, s.logger, s.authzProvider, "UpdateUser")
	if err != nil {
		return nil, err
	}

	user, err := s.userService.UpdateUser(ctx, sophrosyne.UpdateUserRequest{
		Name:    req.GetName(),
		Email:   req.GetEmail(),
		IsAdmin: req.GetIsAdmin(),
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "unable to update user", "error", err)
		return nil, err
	}

	s.logger.InfoContext(ctx, "updated user", "user_id", targetUser.ID)
	resp := &v0.UpdateUserResponse{
		Name:      user.Name,
		Email:     user.Email,
		IsAdmin:   user.IsAdmin,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}

	if user.DeletedAt != nil {
		resp.DeletedAt = timestamppb.New(*user.DeletedAt)
	}

	return resp, nil
}

func (s *UserServiceServer) DeleteUser(ctx context.Context, req *v0.DeleteUserRequest) (*emptypb.Empty, error) {
	targetUser, err := getTargetUser(ctx, req.GetName(), s.userService, s.logger, s.authzProvider, "DeleteUser")
	if err != nil {
		return nil, err
	}

	err = s.userService.DeleteUser(ctx, req.GetName())
	if err != nil {
		s.logger.ErrorContext(ctx, "unable to delete user", "error", err)
		return nil, err
	}

	s.logger.InfoContext(ctx, "deleted user", "user_id", targetUser.ID)
	return &emptypb.Empty{}, nil
}

func (s *UserServiceServer) RotateToken(ctx context.Context, req *v0.RotateTokenRequest) (*v0.RotateTokenResponse, error) {
	targetUser, err := getTargetUser(ctx, req.GetName(), s.userService, s.logger, s.authzProvider, "RotateToken")
	if err != nil {
		return nil, err
	}

	token, err := s.userService.RotateToken(ctx, req.GetName())
	if err != nil {
		s.logger.ErrorContext(ctx, "unable to rotate token", "error", err)
		return nil, err
	}

	s.logger.InfoContext(ctx, "rotated token", "user_id", targetUser.ID)
	return &v0.RotateTokenResponse{
		Token: base64.StdEncoding.EncodeToString(token),
	}, nil
}

func getTargetUser(ctx context.Context, targetUserName string, userService sophrosyne.UserService, logger *slog.Logger, authzProvider sophrosyne.AuthorizationProvider, action string) (*sophrosyne.User, error) {
	curUser := sophrosyne.ExtractUser(ctx)
	if curUser == nil {
		return nil, status.Errorf(codes.Unauthenticated, InvalidTokenMsg)
	}

	targetUser, err := userService.GetUserByName(ctx, targetUserName)
	if err != nil {
		logger.ErrorContext(ctx, "unable to get user", "error", err)
		return nil, status.Errorf(codes.Internal, "unable to get user")
	}

	ok := authzProvider.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curUser,
		Action:    sophrosyne.AuthorizationAction(action),
		Resource:  sophrosyne.User{ID: targetUser.ID},
	})
	if !ok {
		return nil, status.Errorf(codes.PermissionDenied, "unauthorized")
	}

	return &targetUser, nil
}
