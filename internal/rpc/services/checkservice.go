package services

import (
	"context"
	"errors"
	"log/slog"

	"github.com/madsrc/sophrosyne"
	"github.com/madsrc/sophrosyne/internal/rpc"
	"github.com/madsrc/sophrosyne/internal/rpc/internal/jsonrpc"
)

type CheckService struct {
	methods      map[jsonrpc.Method]rpc.Method
	checkService sophrosyne.CheckService
	authz        sophrosyne.AuthorizationProvider
	logger       *slog.Logger
	validator    sophrosyne.Validator
}

func NewCheckService(checkService sophrosyne.CheckService, authz sophrosyne.AuthorizationProvider, logger *slog.Logger, validator sophrosyne.Validator) (*CheckService, error) {
	u := &CheckService{
		methods:      make(map[jsonrpc.Method]rpc.Method),
		checkService: checkService,
		authz:        authz,
		logger:       logger,
		validator:    validator,
	}

	u.methods["Checks::GetCheck"] = getCheck{service: u}
	u.methods["Checks::GetChecks"] = getChecks{service: u}
	u.methods["Checks::CreateCheck"] = createCheck{service: u}
	u.methods["Checks::UpdateCheck"] = updateCheck{service: u}
	u.methods["Checks::DeleteCheck"] = deleteCheck{service: u}

	return u, nil
}

func (u CheckService) EntityType() string {
	return "Service"
}

func (u CheckService) EntityID() string {
	return "Checks"
}

func (u CheckService) InvokeMethod(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	return invokeMethod(ctx, u.logger, u.methods, req)
}

type getCheck struct {
	service *CheckService
}

func (u getCheck) GetService() rpc.Service {
	return u.service
}

func (u getCheck) EntityType() string {
	return "Checks"
}

func (u getCheck) EntityID() string {
	return "GetCheck"
}

func (u getCheck) Invoke(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.GetCheckRequest
	err := rpc.ParamsIntoAny(&req, &params, u.service.validator)
	if err != nil {
		u.service.logger.ErrorContext(ctx, "error extracting params from request", "error", err)
		return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
	}

	if params.Name != "" {
		u, _ := u.service.checkService.GetCheckByName(ctx, params.Name)
		params.ID = u.ID
	}

	curCheck := sophrosyne.ExtractUser(ctx)
	if curCheck == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	if !u.service.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curCheck,
		Action:    u,
		Resource:  sophrosyne.Check{ID: params.ID},
	}) {
		return rpc.ErrorFromRequest(&req, 12345, "unauthorized")
	}

	check, err := u.service.checkService.GetCheck(ctx, params.ID)
	if err != nil {
		u.service.logger.ErrorContext(ctx, "unable to get check", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, "check not found")
	}

	resp := sophrosyne.GetCheckResponse{}

	return rpc.ResponseToRequest(&req, resp.FromCheck(check))
}

type getChecks struct {
	service *CheckService
}

func (u getChecks) GetService() rpc.Service {
	return u.service
}

func (u getChecks) EntityType() string {
	return "Checks"
}

func (u getChecks) EntityID() string {
	return "GetChecks"
}

func (u getChecks) Invoke(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.GetChecksRequest
	err := rpc.ParamsIntoAny(&req, &params, u.service.validator)
	if err != nil {
		if errors.Is(err, rpc.NoParamsError) {
			params = sophrosyne.GetChecksRequest{}
		} else {
			u.service.logger.ErrorContext(ctx, "error extracting params from request", "error", err)
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
			u.service.logger.ErrorContext(ctx, "unable to decode cursor", "error", err)
			return rpc.ErrorFromRequest(&req, 12347, "invalid cursor")
		}
	} else {
		cursor = sophrosyne.NewDatabaseCursor(curCheck.ID, "")
	}

	checks, err := u.service.checkService.GetChecks(ctx, cursor)
	if err != nil {
		u.service.logger.ErrorContext(ctx, "unable to get checks", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, "checks not found")
	}

	var checksResponse []sophrosyne.GetCheckResponse
	for _, uu := range checks {
		ok := u.service.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
			Principal: curCheck,
			Action:    u,
			Resource:  sophrosyne.Check{ID: uu.ID},
		})
		if ok {
			ent := &sophrosyne.GetCheckResponse{}
			checksResponse = append(checksResponse, *ent.FromCheck(uu))
		}
	}

	u.service.logger.DebugContext(ctx, "returning checks", "total", len(checksResponse), "checks", checksResponse)
	return rpc.ResponseToRequest(&req, sophrosyne.GetChecksResponse{
		Checks: checksResponse,
		Cursor: cursor.String(),
		Total:  len(checksResponse),
	})
}

type createCheck struct {
	service *CheckService
}

func (u createCheck) GetService() rpc.Service {
	return u.service
}

func (u createCheck) EntityType() string {
	return "Checks"
}

func (u createCheck) EntityID() string {
	return "CreateCheck"
}

func (u createCheck) Invoke(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.CreateCheckRequest
	err := rpc.ParamsIntoAny(&req, &params, u.service.validator)
	if err != nil {
		u.service.logger.ErrorContext(ctx, "error extracting params from request", "error", err)
		return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
	}

	curCheck := sophrosyne.ExtractUser(ctx)
	if curCheck == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	ok := u.service.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curCheck,
		Action:    u,
	})

	if !ok {
		return rpc.ErrorFromRequest(&req, 12345, "unauthorized")
	}

	check, err := u.service.checkService.CreateCheck(ctx, params)
	if err != nil {
		u.service.logger.ErrorContext(ctx, "unable to create check", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, "unable to create check")
	}

	resp := sophrosyne.CreateCheckResponse{}
	return rpc.ResponseToRequest(&req, resp.FromCheck(check))
}

type updateCheck struct {
	service *CheckService
}

func (u updateCheck) GetService() rpc.Service {
	return u.service
}

func (u updateCheck) EntityType() string {
	return "Checks"
}

func (u updateCheck) EntityID() string {
	return "CreateCheck"
}

func (u updateCheck) Invoke(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.UpdateCheckRequest
	err := rpc.ParamsIntoAny(&req, &params, u.service.validator)
	if err != nil {
		u.service.logger.ErrorContext(ctx, "error extracting params from request", "error", err)
		return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
	}

	curCheck := sophrosyne.ExtractUser(ctx)
	if curCheck == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	checkToUpdate, err := u.service.checkService.GetCheckByName(ctx, params.Name)
	if err != nil {
		return rpc.ErrorFromRequest(&req, 12346, "check not found")
	}

	ok := u.service.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curCheck,
		Action:    u,
		Resource:  sophrosyne.Check{ID: checkToUpdate.ID},
	})

	if !ok {
		return rpc.ErrorFromRequest(&req, 12345, "unauthorized")
	}

	check, err := u.service.checkService.UpdateCheck(ctx, params)
	if err != nil {
		u.service.logger.ErrorContext(ctx, "unable to update check", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, "unable to update check")
	}

	resp := &sophrosyne.UpdateCheckResponse{}
	return rpc.ResponseToRequest(&req, resp.FromCheck(check))
}

type deleteCheck struct {
	service *CheckService
}

func (u deleteCheck) GetService() rpc.Service {
	return u.service
}

func (u deleteCheck) EntityType() string {
	return "Checks"
}

func (u deleteCheck) EntityID() string {
	return "CreateCheck"
}

func (u deleteCheck) Invoke(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	var params sophrosyne.DeleteCheckRequest
	err := rpc.ParamsIntoAny(&req, &params, u.service.validator)
	if err != nil {
		u.service.logger.ErrorContext(ctx, "error extracting params from request", "error", err)
		return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
	}

	curCheck := sophrosyne.ExtractUser(ctx)
	if curCheck == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	checkToDelete, err := u.service.checkService.GetCheckByName(ctx, params.Name)
	if err != nil {
		return rpc.ErrorFromRequest(&req, 12346, "check not found")
	}

	ok := u.service.authz.IsAuthorized(ctx, sophrosyne.AuthorizationRequest{
		Principal: curCheck,
		Action:    u,
		Resource:  sophrosyne.Check{ID: checkToDelete.ID},
	})

	if !ok {
		return rpc.ErrorFromRequest(&req, 12345, "unauthorized")
	}

	err = u.service.checkService.DeleteCheck(ctx, checkToDelete.Name)
	if err != nil {
		u.service.logger.ErrorContext(ctx, "unable to delete check", "error", err)
		return rpc.ErrorFromRequest(&req, 12346, "unable to delete check")
	}

	return rpc.ResponseToRequest(&req, "ok")
}
