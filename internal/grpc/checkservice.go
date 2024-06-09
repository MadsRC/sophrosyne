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
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/madsrc/sophrosyne"
	v0 "github.com/madsrc/sophrosyne/internal/grpc/sophrosyne/v0"
	"github.com/madsrc/sophrosyne/internal/validator"
)

type CheckServiceServer struct {
	v0.UnimplementedCheckServiceServer
	logger        *slog.Logger                     `validate:"required"`
	config        *sophrosyne.Config               `validate:"required"`
	validator     sophrosyne.Validator             `validate:"required"`
	checkService  sophrosyne.CheckService          `validate:"required"`
	authzProvider sophrosyne.AuthorizationProvider `validate:"required"`
}

func newGetCheckResponseFromCheck(check *sophrosyne.Check) *v0.GetCheckResponse {
	resp := &v0.GetCheckResponse{
		Name:      check.Name,
		CreatedAt: timestamppb.New(check.CreatedAt),
		UpdatedAt: timestamppb.New(check.UpdatedAt),
	}
	for _, profile := range check.Profiles {
		resp.Profiles = append(resp.Profiles, profile.Name)
	}

	for _, svc := range check.UpstreamServices {
		resp.UpstreamServices = append(resp.UpstreamServices, svc.Host)
	}

	if check.DeletedAt != nil {
		resp.DeletedAt = timestamppb.New(*check.DeletedAt)
	}

	return resp
}

func (p CheckServiceServer) GetCheck(ctx context.Context, request *v0.GetCheckRequest) (*v0.GetCheckResponse, error) {
	curUser := sophrosyne.ExtractUser(ctx)
	if curUser == nil {
		return nil, status.Errorf(codes.Unauthenticated, InvalidTokenMsg)
	}

	var check sophrosyne.Check
	var err error

	if request.GetId() != "" {
		check, err = p.checkService.GetCheck(ctx, request.GetId())
	} else {
		check, err = p.checkService.GetCheckByName(ctx, request.GetName())
	}

	if err != nil {
		return nil, status.Errorf(codes.Internal, "error getting check: %v", err)
	}

	ok := p.authzProvider.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curUser,
		Action:    sophrosyne.AuthorizationAction("GetProfile"),
		Resource:  sophrosyne.Profile{ID: check.ID},
	})

	if !ok {
		return nil, status.Errorf(codes.PermissionDenied, "unauthorized")
	}

	return newGetCheckResponseFromCheck(&check), nil
}

func (p CheckServiceServer) GetChecks(ctx context.Context, request *v0.GetChecksRequest) (*v0.GetChecksResponse, error) {
	curUser := sophrosyne.ExtractUser(ctx)
	if curUser == nil {
		return nil, status.Errorf(codes.Unauthenticated, InvalidTokenMsg)
	}

	cursor := &sophrosyne.DatabaseCursor{}
	var err error
	if request.GetCursor() != "" {
		cursor, err = sophrosyne.DecodeDatabaseCursorWithOwner(request.GetCursor(), curUser.ID)
		if err != nil {
			p.logger.ErrorContext(ctx, "unable to decode cursor", "error", err)
			return nil, status.Errorf(codes.InvalidArgument, InvalidCursorMsg)
		}
	}

	checks, err := p.checkService.GetChecks(ctx, cursor)
	if err != nil {
		p.logger.ErrorContext(ctx, "unable to get checks", "error", err)
		return nil, status.Error(codes.Internal, "internal error getting checks")
	}

	var checkResponse []*v0.GetCheckResponse
	for _, check := range checks {
		ok := p.authzProvider.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
			Principal: curUser,
			Action:    sophrosyne.AuthorizationAction("GetCheck"),
			Resource:  sophrosyne.Check{ID: check.ID},
		})
		if ok {
			checkResponse = append(checkResponse, newGetCheckResponseFromCheck(&check)) // #nosec G601
		}
	}

	return &v0.GetChecksResponse{
		Checks: checkResponse,
		Cursor: cursor.String(),
		Total:  int32(len(checks)),
	}, nil
}

func (p CheckServiceServer) CreateCheck(ctx context.Context, request *v0.CreateCheckRequest) (*v0.CreateCheckResponse, error) {
	curUser := sophrosyne.ExtractUser(ctx)
	if curUser == nil {
		return nil, status.Errorf(codes.Unauthenticated, InvalidTokenMsg)
	}

	ok := p.authzProvider.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curUser,
		Action:    sophrosyne.AuthorizationAction("CreateProfile"),
	})

	if !ok {
		return nil, status.Errorf(codes.PermissionDenied, "unauthorized")
	}

	check, err := p.checkService.CreateCheck(ctx, sophrosyne.CreateCheckRequest{
		Name:             request.GetName(),
		Profiles:         request.Profiles,
		UpstreamServices: request.UpstreamServices,
	})

	if err != nil {
		p.logger.ErrorContext(ctx, "unable to create check", "error", err)
		return nil, status.Error(codes.Internal, "internal error creating check")
	}

	p.logger.InfoContext(ctx, "check created", "check", check)
	resp := &v0.CreateCheckResponse{
		Name:             check.Name,
		Profiles:         request.Profiles,
		UpstreamServices: request.UpstreamServices,
		CreatedAt:        timestamppb.New(check.CreatedAt),
		UpdatedAt:        timestamppb.New(check.UpdatedAt),
	}

	return resp, nil
}

func (p CheckServiceServer) UpdateCheck(ctx context.Context, request *v0.UpdateCheckRequest) (*v0.UpdateCheckResponse, error) {
	target, err := getTargetCheck(ctx, request.GetName(), p.checkService, p.logger, p.authzProvider, "UpdateCheck")
	if err != nil {
		return nil, err
	}

	check, err := p.checkService.UpdateCheck(ctx, sophrosyne.UpdateCheckRequest{
		Name:             target.Name,
		Profiles:         request.Profiles,
		UpstreamServices: request.UpstreamServices,
	})
	if err != nil {
		p.logger.ErrorContext(ctx, "unable to update check", "error", err)
		return nil, status.Error(codes.Internal, "internal error updating check")
	}

	p.logger.InfoContext(ctx, "check updated", "check", check)
	resp := &v0.UpdateCheckResponse{
		Name:             check.Name,
		Profiles:         request.Profiles,
		UpstreamServices: request.UpstreamServices,
		CreatedAt:        timestamppb.New(check.CreatedAt),
		UpdatedAt:        timestamppb.New(check.UpdatedAt),
	}

	if check.DeletedAt != nil {
		resp.DeletedAt = timestamppb.New(*check.DeletedAt)
	}

	return resp, nil
}

func (p CheckServiceServer) DeleteCheck(ctx context.Context, request *v0.DeleteCheckRequest) (*emptypb.Empty, error) {
	target, err := getTargetCheck(ctx, request.GetName(), p.checkService, p.logger, p.authzProvider, "DeleteCheck")
	if err != nil {
		return nil, err
	}

	err = p.checkService.DeleteCheck(ctx, target.ID)
	if err != nil {
		p.logger.ErrorContext(ctx, "unable to delete check", "error", err)
		return nil, status.Error(codes.Internal, "internal error deleting check")
	}

	p.logger.InfoContext(ctx, "check deleted", "check", target)
	return &emptypb.Empty{}, nil
}

func getTargetCheck(ctx context.Context, targetName string, checkService sophrosyne.CheckService, logger *slog.Logger, authzProvider sophrosyne.AuthorizationProvider, action string) (*sophrosyne.Check, error) {
	curUser := sophrosyne.ExtractUser(ctx)
	if curUser == nil {
		return nil, status.Errorf(codes.Unauthenticated, InvalidTokenMsg)
	}

	target, err := checkService.GetCheckByName(ctx, targetName)
	if err != nil {
		logger.ErrorContext(ctx, "unable to get profile", "error", err)
		return nil, status.Errorf(codes.Internal, "unable to get profile")
	}

	ok := authzProvider.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curUser,
		Action:    sophrosyne.AuthorizationAction(action),
		Resource:  sophrosyne.User{ID: target.ID},
	})
	if !ok {
		return nil, status.Errorf(codes.PermissionDenied, "unauthorized")
	}

	return &target, nil
}

// NewCheckServiceServer returns a new CheckServiceServer instance.
//
// If the provided options are invalid, an error will be returned.
// Required options are marked with the 'validate:"required"' tag in
// the [CheckServiceServer] struct. Every required option has a
// corresponding [Option] function.
//
// If no [sophrosyne.Validator] is provided, a default one will be
// created.
func NewCheckServiceServer(ctx context.Context, opts ...Option) (*CheckServiceServer, error) {
	s := &CheckServiceServer{}
	setOptions(s, defaultCheckServiceServerOptions(), opts...)

	if s.logger != nil {
		s.logger.DebugContext(ctx, "validating server options")
	}
	err := s.validator.Validate(s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func defaultCheckServiceServerOptions() []Option {
	return []Option{
		WithValidator(validator.NewValidator()),
	}
}
