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

	v0 "github.com/madsrc/sophrosyne/internal/grpc/sophrosyne/v0"

	"github.com/madsrc/sophrosyne"
	"github.com/madsrc/sophrosyne/internal/validator"
)

type ProfileServiceServer struct {
	v0.UnimplementedProfileServiceServer
	logger         *slog.Logger                     `validate:"required"`
	config         *sophrosyne.Config               `validate:"required"`
	validator      sophrosyne.Validator             `validate:"required"`
	profileService sophrosyne.ProfileService        `validate:"required"`
	authzProvider  sophrosyne.AuthorizationProvider `validate:"required"`
}

func newGetProfileResponseFromProfile(profile *sophrosyne.Profile) *v0.GetProfileResponse {
	resp := &v0.GetProfileResponse{
		Name:      profile.Name,
		CreatedAt: timestamppb.New(profile.CreatedAt),
		UpdatedAt: timestamppb.New(profile.UpdatedAt),
	}
	for _, check := range profile.Checks {
		resp.Checks = append(resp.Checks, check.Name)
	}

	if profile.DeletedAt != nil {
		resp.DeletedAt = timestamppb.New(*profile.DeletedAt)
	}

	return resp
}

func (p ProfileServiceServer) GetProfile(ctx context.Context, request *v0.GetProfileRequest) (*v0.GetProfileResponse, error) {
	curUser := sophrosyne.ExtractUser(ctx)
	if curUser == nil {
		return nil, status.Errorf(codes.Unauthenticated, InvalidTokenMsg)
	}

	var profile sophrosyne.Profile
	var err error

	if request.GetId() != "" {
		profile, err = p.profileService.GetProfile(ctx, request.GetId())
	} else {
		profile, err = p.profileService.GetProfileByName(ctx, request.GetName())
	}

	if err != nil {
		return nil, status.Errorf(codes.Internal, "error getting profile: %v", err)
	}

	ok := p.authzProvider.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curUser,
		Action:    sophrosyne.AuthorizationAction("GetProfile"),
		Resource:  sophrosyne.Profile{ID: profile.ID},
	})

	if !ok {
		return nil, status.Errorf(codes.PermissionDenied, "unauthorized")
	}

	return newGetProfileResponseFromProfile(&profile), nil
}

func (p ProfileServiceServer) GetProfiles(ctx context.Context, request *v0.GetProfilesRequest) (*v0.GetProfilesResponse, error) {
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

	profiles, err := p.profileService.GetProfiles(ctx, cursor)
	if err != nil {
		p.logger.ErrorContext(ctx, "unable to get profiles", "error", err)
		return nil, status.Error(codes.Internal, "internal error getting profiles")
	}

	var profilesResponse []*v0.GetProfileResponse
	for _, profile := range profiles {
		ok := p.authzProvider.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
			Principal: curUser,
			Action:    sophrosyne.AuthorizationAction("GetProfile"),
			Resource:  sophrosyne.Profile{ID: profile.ID},
		})
		if ok {
			profilesResponse = append(profilesResponse, newGetProfileResponseFromProfile(&profile)) // #nosec G601
		}
	}

	return &v0.GetProfilesResponse{
		Profiles: profilesResponse,
		Cursor:   cursor.String(),
		Total:    int32(len(profiles)),
	}, nil
}

func (p ProfileServiceServer) CreateProfile(ctx context.Context, request *v0.CreateProfileRequest) (*v0.CreateProfileResponse, error) {
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

	profile, err := p.profileService.CreateProfile(ctx, sophrosyne.CreateProfileRequest{
		Name:   request.GetName(),
		Checks: request.Checks,
	})

	if err != nil {
		p.logger.ErrorContext(ctx, "unable to create profile", "error", err)
		return nil, status.Error(codes.Internal, "internal error creating profile")
	}

	p.logger.InfoContext(ctx, "profile created", "profile", profile)
	resp := &v0.CreateProfileResponse{
		Name:      profile.Name,
		Checks:    request.Checks,
		CreatedAt: timestamppb.New(profile.CreatedAt),
		UpdatedAt: timestamppb.New(profile.UpdatedAt),
	}

	return resp, nil
}

func (p ProfileServiceServer) UpdateProfile(ctx context.Context, request *v0.UpdateProfileRequest) (*v0.UpdateProfileResponse, error) {
	target, err := getTargetProfile(ctx, request.GetName(), p.profileService, p.logger, p.authzProvider, "UpdateProfile")
	if err != nil {
		return nil, err
	}

	profile, err := p.profileService.UpdateProfile(ctx, sophrosyne.UpdateProfileRequest{
		Name:   target.Name,
		Checks: request.Checks,
	})
	if err != nil {
		p.logger.ErrorContext(ctx, "unable to update profile", "error", err)
		return nil, status.Error(codes.Internal, "internal error updating profile")
	}

	p.logger.InfoContext(ctx, "profile updated", "profile", profile)
	resp := &v0.UpdateProfileResponse{
		Name:      profile.Name,
		Checks:    request.Checks,
		CreatedAt: timestamppb.New(profile.CreatedAt),
		UpdatedAt: timestamppb.New(profile.UpdatedAt),
	}

	if profile.DeletedAt != nil {
		resp.DeletedAt = timestamppb.New(*profile.DeletedAt)
	}

	return resp, nil
}

func (p ProfileServiceServer) DeleteProfile(ctx context.Context, request *v0.DeleteProfileRequest) (*emptypb.Empty, error) {
	target, err := getTargetProfile(ctx, request.GetName(), p.profileService, p.logger, p.authzProvider, "DeleteProfile")
	if err != nil {
		return nil, err
	}

	err = p.profileService.DeleteProfile(ctx, target.ID)
	if err != nil {
		p.logger.ErrorContext(ctx, "unable to delete profile", "error", err)
		return nil, status.Error(codes.Internal, "internal error deleting profile")
	}

	p.logger.InfoContext(ctx, "profile deleted", "profile", target)
	return &emptypb.Empty{}, nil
}

func getTargetProfile(ctx context.Context, targetName string, profileService sophrosyne.ProfileService, logger *slog.Logger, authzProvider sophrosyne.AuthorizationProvider, action string) (*sophrosyne.Profile, error) {
	curUser := sophrosyne.ExtractUser(ctx)
	if curUser == nil {
		return nil, status.Errorf(codes.Unauthenticated, InvalidTokenMsg)
	}

	target, err := profileService.GetProfileByName(ctx, targetName)
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

// NewProfileServiceServer returns a new ProfileServiceServer instance.
//
// If the provided options are invalid, an error will be returned.
// Required options are marked with the 'validate:"required"' tag in
// the [ProfileServiceServer] struct. Every required option has a
// corresponding [Option] function.
//
// If no [sophrosyne.Validator] is provided, a default one will be
// created.
func NewProfileServiceServer(ctx context.Context, opts ...Option) (*ProfileServiceServer, error) {
	s := &ProfileServiceServer{}
	setOptions(s, defaultProfileServiceServerOptions(), opts...)

	if s.logger != nil {
		s.logger.DebugContext(ctx, "validating server options")
	}
	err := s.validator.Validate(s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func defaultProfileServiceServerOptions() []Option {
	return []Option{
		WithValidator(validator.NewValidator()),
	}
}
