package services

import (
	"context"
	"errors"
	"log/slog"

	"github.com/madsrc/sophrosyne"
	"github.com/madsrc/sophrosyne/internal/rpc"
	"github.com/madsrc/sophrosyne/internal/rpc/internal/jsonrpc"
)

type ProfileService struct {
	methods        map[jsonrpc.Method]rpc.Method
	profileService sophrosyne.ProfileService
	authz          sophrosyne.AuthorizationProvider
	logger         *slog.Logger
	validator      sophrosyne.Validator
}

func NewProfileService(profileService sophrosyne.ProfileService, authz sophrosyne.AuthorizationProvider, logger *slog.Logger, validator sophrosyne.Validator) (*ProfileService, error) {
	u := &ProfileService{
		methods:        make(map[jsonrpc.Method]rpc.Method),
		profileService: profileService,
		authz:          authz,
		logger:         logger,
		validator:      validator,
	}

	u.methods["Profiles::GetProfile"] = getProfile{service: u}
	u.methods["Profiles::GetProfiles"] = getProfiles{service: u}
	u.methods["Profiles::CreateProfile"] = createProfile{service: u}
	u.methods["Profiles::UpdateProfile"] = updateProfile{service: u}
	u.methods["Profiles::DeleteProfile"] = deleteProfile{service: u}

	return u, nil
}

func (u ProfileService) EntityType() string {
	return "Service"
}

func (u ProfileService) EntityID() string {
	return "Profiles"
}

func (u ProfileService) InvokeMethod(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	return invokeMethod(ctx, u.logger, u.methods, req)
}

type getProfile struct {
	service *ProfileService
}

func (u getProfile) GetService() rpc.Service {
	return u.service
}

func (u getProfile) EntityType() string {
	return "Profiles"
}

func (u getProfile) EntityID() string {
	return "GetProfile"
}

func (u getProfile) Invoke(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.GetProfileRequest
	err := rpc.ParamsIntoAny(&req, &params, u.service.validator)
	if err != nil {
		u.service.logger.ErrorContext(ctx, "error extracting params from request", "error", err)
		return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
	}

	if params.Name != "" {
		u, _ := u.service.profileService.GetProfileByName(ctx, params.Name)
		params.ID = u.ID
	}

	curUser := sophrosyne.ExtractUser(ctx)
	if curUser == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	if !u.service.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curUser,
		Action:    u,
		Resource:  sophrosyne.Profile{ID: params.ID},
	}) {
		return rpc.ErrorFromRequest(&req, 12345, "unauthorized")
	}

	Profile, err := u.service.profileService.GetProfile(ctx, params.ID)
	if err != nil {
		u.service.logger.ErrorContext(ctx, "unable to get Profile", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, "Profile not found")
	}

	resp := sophrosyne.GetProfileResponse{}

	return rpc.ResponseToRequest(&req, resp.FromProfile(Profile))
}

type getProfiles struct {
	service *ProfileService
}

func (u getProfiles) GetService() rpc.Service {
	return u.service
}

func (u getProfiles) EntityType() string {
	return "Profiles"
}

func (u getProfiles) EntityID() string {
	return "GetProfiles"
}

func (u getProfiles) Invoke(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.GetProfilesRequest
	err := rpc.ParamsIntoAny(&req, &params, u.service.validator)
	if err != nil {
		if errors.Is(err, rpc.NoParamsError) {
			params = sophrosyne.GetProfilesRequest{}
		} else {
			u.service.logger.ErrorContext(ctx, "error extracting params from request", "error", err)
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
			u.service.logger.ErrorContext(ctx, "unable to decode cursor", "error", err)
			return rpc.ErrorFromRequest(&req, 12347, "invalid cursor")
		}
	} else {
		cursor = sophrosyne.NewDatabaseCursor(curProfile.ID, "")
	}

	Profiles, err := u.service.profileService.GetProfiles(ctx, cursor)
	if err != nil {
		u.service.logger.ErrorContext(ctx, "unable to get Profiles", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, "Profiles not found")
	}

	var ProfilesResponse []sophrosyne.GetProfileResponse
	for _, uu := range Profiles {
		ok := u.service.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
			Principal: curProfile,
			Action:    u,
			Resource:  sophrosyne.Profile{ID: uu.ID},
		})
		if ok {
			ent := &sophrosyne.GetProfileResponse{}
			ProfilesResponse = append(ProfilesResponse, *ent.FromProfile(uu))
		}
	}

	u.service.logger.DebugContext(ctx, "returning Profiles", "total", len(ProfilesResponse), "Profiles", ProfilesResponse)
	return rpc.ResponseToRequest(&req, sophrosyne.GetProfilesResponse{
		Profiles: ProfilesResponse,
		Cursor:   cursor.String(),
		Total:    len(ProfilesResponse),
	})
}

type createProfile struct {
	service *ProfileService
}

func (u createProfile) GetService() rpc.Service {
	return u.service
}

func (u createProfile) EntityType() string {
	return "Profiles"
}

func (u createProfile) EntityID() string {
	return "CreateProfile"
}

func (u createProfile) Invoke(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.CreateProfileRequest
	err := rpc.ParamsIntoAny(&req, &params, u.service.validator)
	if err != nil {
		u.service.logger.ErrorContext(ctx, "error extracting params from request", "error", err)
		return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
	}

	curProfile := sophrosyne.ExtractUser(ctx)
	if curProfile == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	ok := u.service.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curProfile,
		Action:    u,
	})

	if !ok {
		return rpc.ErrorFromRequest(&req, 12345, "unauthorized")
	}

	Profile, err := u.service.profileService.CreateProfile(ctx, params)
	if err != nil {
		u.service.logger.ErrorContext(ctx, "unable to create Profile", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, "unable to create Profile")
	}

	resp := sophrosyne.CreateProfileResponse{}
	return rpc.ResponseToRequest(&req, resp.FromProfile(Profile))
}

type updateProfile struct {
	service *ProfileService
}

func (u updateProfile) GetService() rpc.Service {
	return u.service
}

func (u updateProfile) EntityType() string {
	return "Profiles"
}

func (u updateProfile) EntityID() string {
	return "CreateProfile"
}

func (u updateProfile) Invoke(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.UpdateProfileRequest
	err := rpc.ParamsIntoAny(&req, &params, u.service.validator)
	if err != nil {
		u.service.logger.ErrorContext(ctx, "error extracting params from request", "error", err)
		return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
	}

	curProfile := sophrosyne.ExtractUser(ctx)
	if curProfile == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	ProfileToUpdate, err := u.service.profileService.GetProfileByName(ctx, params.Name)
	if err != nil {
		return rpc.ErrorFromRequest(&req, 12346, "Profile not found")
	}

	ok := u.service.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curProfile,
		Action:    u,
		Resource:  sophrosyne.Profile{ID: ProfileToUpdate.ID},
	})

	if !ok {
		return rpc.ErrorFromRequest(&req, 12345, "unauthorized")
	}

	Profile, err := u.service.profileService.UpdateProfile(ctx, params)
	if err != nil {
		u.service.logger.ErrorContext(ctx, "unable to update Profile", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, "unable to update Profile")
	}

	resp := &sophrosyne.UpdateProfileResponse{}
	return rpc.ResponseToRequest(&req, resp.FromProfile(Profile))
}

type deleteProfile struct {
	service *ProfileService
}

func (u deleteProfile) GetService() rpc.Service {
	return u.service
}

func (u deleteProfile) EntityType() string {
	return "Profiles"
}

func (u deleteProfile) EntityID() string {
	return "CreateProfile"
}

func (u deleteProfile) Invoke(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.DeleteProfileRequest
	err := rpc.ParamsIntoAny(&req, &params, u.service.validator)
	if err != nil {
		u.service.logger.ErrorContext(ctx, "error extracting params from request", "error", err)
		return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
	}

	curProfile := sophrosyne.ExtractUser(ctx)
	if curProfile == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	ProfileToDelete, err := u.service.profileService.GetProfileByName(ctx, params.Name)
	if err != nil {
		return rpc.ErrorFromRequest(&req, 12346, "Profile not found")
	}

	ok := u.service.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curProfile,
		Action:    u,
		Resource:  sophrosyne.Profile{ID: ProfileToDelete.ID},
	})

	if !ok {
		return rpc.ErrorFromRequest(&req, 12345, "unauthorized")
	}

	err = u.service.profileService.DeleteProfile(ctx, ProfileToDelete.Name)
	if err != nil {
		u.service.logger.ErrorContext(ctx, "unable to delete Profile", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, "unable to delete Profile")
	}

	return rpc.ResponseToRequest(&req, "ok")
}
