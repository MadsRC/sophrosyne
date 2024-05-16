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

type ProfileService struct {
	profileService sophrosyne.ProfileService
	authz          sophrosyne.AuthorizationProvider
	logger         *slog.Logger
	validator      sophrosyne.Validator
}

func NewProfileService(profileService sophrosyne.ProfileService, authz sophrosyne.AuthorizationProvider, logger *slog.Logger, validator sophrosyne.Validator) (*ProfileService, error) {
	u := &ProfileService{
		profileService: profileService,
		authz:          authz,
		logger:         logger,
		validator:      validator,
	}

	return u, nil
}

func (u ProfileService) EntityType() string {
	return "Service"
}

func (u ProfileService) EntityID() string {
	return "Profiles"
}

func (u ProfileService) InvokeMethod(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	m := strings.Split(string(req.Method), "::")
	if len(m) != 2 {
		u.logger.ErrorContext(ctx, "unreachable", "error", sophrosyne.NewUnreachableCodeError())
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}
	switch m[1] {
	case "GetProfile":
		return u.GetProfile(ctx, req)
	case "GetProfiles":
		return u.GetProfiles(ctx, req)
	case "CreateProfile":
		return u.CreateProfile(ctx, req)
	case "UpdateProfile":
		return u.UpdateProfile(ctx, req)
	case "DeleteProfile":
		return u.DeleteProfile(ctx, req)
	default:
		u.logger.DebugContext(ctx, "cannot invoke method", "method", req.Method)
		return rpc.ErrorFromRequest(&req, jsonrpc.MethodNotFound, string(jsonrpc.MethodNotFoundMessage))
	}
}

func (u ProfileService) GetProfile(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.GetProfileRequest
	err := rpc.ParamsIntoAny(&req, &params, u.validator)
	if err != nil {
		u.logger.ErrorContext(ctx, paramExtractError, "error", err)
		return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
	}

	if params.Name != "" {
		u, _ := u.profileService.GetProfileByName(ctx, params.Name)
		params.ID = u.ID
	}

	curUser := sophrosyne.ExtractUser(ctx)
	if curUser == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	if !u.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curUser,
		Action:    sophrosyne.AuthorizationAction("GetProfile"),
		Resource:  sophrosyne.Profile{ID: params.ID},
	}) {
		return rpc.ErrorFromRequest(&req, 12345, "unauthorized")
	}

	Profile, err := u.profileService.GetProfile(ctx, params.ID)
	if err != nil {
		u.logger.ErrorContext(ctx, "unable to get Profile", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, profileNotFoundError)
	}

	resp := sophrosyne.GetProfileResponse{}

	return rpc.ResponseToRequest(&req, resp.FromProfile(Profile))
}

const profileNotFoundError = "profile not found"

func (u ProfileService) GetProfiles(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.GetProfilesRequest
	err := rpc.ParamsIntoAny(&req, &params, u.validator)
	if err != nil {
		if errors.Is(err, rpc.ErrNoParams) {
			params = sophrosyne.GetProfilesRequest{}
		} else {
			u.logger.ErrorContext(ctx, paramExtractError, "error", err)
			return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
		}
	}

	curProfile := sophrosyne.ExtractUser(ctx)
	if curProfile == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	var cursor *sophrosyne.DatabaseCursor
	if params.Cursor != "" {
		cursor, err = sophrosyne.DecodeDatabaseCursorWithOwner(params.Cursor, curProfile.ID)
		if err != nil {
			u.logger.ErrorContext(ctx, "unable to decode cursor", "error", err)
			return rpc.ErrorFromRequest(&req, 12347, "invalid cursor")
		}
	} else {
		cursor = sophrosyne.NewDatabaseCursor(curProfile.ID, "")
	}

	Profiles, err := u.profileService.GetProfiles(ctx, cursor)
	if err != nil {
		u.logger.ErrorContext(ctx, "unable to get Profiles", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, "Profiles not found")
	}

	var ProfilesResponse []sophrosyne.GetProfileResponse
	for _, uu := range Profiles {
		ok := u.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
			Principal: curProfile,
			Action:    sophrosyne.AuthorizationAction("GetProfiles"),
			Resource:  sophrosyne.Profile{ID: uu.ID},
		})
		if ok {
			ent := &sophrosyne.GetProfileResponse{}
			ProfilesResponse = append(ProfilesResponse, *ent.FromProfile(uu))
		}
	}

	u.logger.DebugContext(ctx, "returning Profiles", "total", len(ProfilesResponse), "Profiles", ProfilesResponse)
	return rpc.ResponseToRequest(&req, sophrosyne.GetProfilesResponse{
		Profiles: ProfilesResponse,
		Cursor:   cursor.String(),
		Total:    len(ProfilesResponse),
	})
}

func (u ProfileService) CreateProfile(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.CreateProfileRequest
	err := rpc.ParamsIntoAny(&req, &params, u.validator)
	if err != nil {
		u.logger.ErrorContext(ctx, paramExtractError, "error", err)
		return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
	}

	curProfile := sophrosyne.ExtractUser(ctx)
	if curProfile == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	ok := u.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curProfile,
		Action:    sophrosyne.AuthorizationAction("CreateProfile"),
	})

	if !ok {
		return rpc.ErrorFromRequest(&req, 12345, "unauthorized")
	}

	Profile, err := u.profileService.CreateProfile(ctx, params)
	if err != nil {
		u.logger.ErrorContext(ctx, "unable to create Profile", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, "unable to create Profile")
	}

	resp := sophrosyne.CreateProfileResponse{}
	return rpc.ResponseToRequest(&req, resp.FromProfile(Profile))
}

func (u ProfileService) UpdateProfile(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.UpdateProfileRequest
	err := rpc.ParamsIntoAny(&req, &params, u.validator)
	if err != nil {
		u.logger.ErrorContext(ctx, paramExtractError, "error", err)
		return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
	}

	curProfile := sophrosyne.ExtractUser(ctx)
	if curProfile == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	ProfileToUpdate, err := u.profileService.GetProfileByName(ctx, params.Name)
	if err != nil {
		return rpc.ErrorFromRequest(&req, 12346, profileNotFoundError)
	}

	ok := u.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curProfile,
		Action:    sophrosyne.AuthorizationAction("UpdateProfile"),
		Resource:  sophrosyne.Profile{ID: ProfileToUpdate.ID},
	})

	if !ok {
		return rpc.ErrorFromRequest(&req, 12345, "unauthorized")
	}

	Profile, err := u.profileService.UpdateProfile(ctx, params)
	if err != nil {
		u.logger.ErrorContext(ctx, "unable to update Profile", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, "unable to update Profile")
	}

	resp := &sophrosyne.UpdateProfileResponse{}
	return rpc.ResponseToRequest(&req, resp.FromProfile(Profile))
}

func (u ProfileService) DeleteProfile(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.DeleteProfileRequest
	err := rpc.ParamsIntoAny(&req, &params, u.validator)
	if err != nil {
		u.logger.ErrorContext(ctx, paramExtractError, "error", err)
		return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
	}

	curProfile := sophrosyne.ExtractUser(ctx)
	if curProfile == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	ProfileToDelete, err := u.profileService.GetProfileByName(ctx, params.Name)
	if err != nil {
		return rpc.ErrorFromRequest(&req, 12346, profileNotFoundError)
	}

	ok := u.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curProfile,
		Action:    sophrosyne.AuthorizationAction("DeleteProfile"),
		Resource:  sophrosyne.Profile{ID: ProfileToDelete.ID},
	})

	if !ok {
		return rpc.ErrorFromRequest(&req, 12345, "unauthorized")
	}

	err = u.profileService.DeleteProfile(ctx, ProfileToDelete.Name)
	if err != nil {
		u.logger.ErrorContext(ctx, "unable to delete Profile", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, "unable to delete Profile")
	}

	return rpc.ResponseToRequest(&req, "ok")
}
