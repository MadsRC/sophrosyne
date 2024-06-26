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
	"fmt"
	"log/slog"
	"strings"

	"github.com/madsrc/sophrosyne/internal/rpc/jsonrpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/madsrc/sophrosyne"
	"github.com/madsrc/sophrosyne/internal/grpc/checks"
	"github.com/madsrc/sophrosyne/internal/rpc"
)

type ScanService struct {
	authz          sophrosyne.AuthorizationProvider
	logger         *slog.Logger
	validator      sophrosyne.Validator
	profileService sophrosyne.ProfileService
	checkService   sophrosyne.CheckService
}

func NewScanService(authz sophrosyne.AuthorizationProvider, logger *slog.Logger, validator sophrosyne.Validator, profileService sophrosyne.ProfileService, checkService sophrosyne.CheckService) (*ScanService, error) {
	s := &ScanService{
		authz:          authz,
		logger:         logger,
		validator:      validator,
		profileService: profileService,
		checkService:   checkService,
	}

	return s, nil
}

func (s ScanService) EntityType() string { return "Service" }

func (s ScanService) EntityID() string { return "Scans" }

func (s ScanService) InvokeMethod(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	m := strings.Split(string(req.Method), "::")
	if len(m) != 2 {
		s.logger.ErrorContext(ctx, "unreachable", "error", sophrosyne.NewUnreachableCodeError())
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}
	switch m[1] {
	case "PerformScan":
		return s.PerformScan(ctx, req)
	default:
		s.logger.DebugContext(ctx, "cannot invoke method", "method", req.Method)
		return rpc.ErrorFromRequest(&req, jsonrpc.MethodNotFound, string(jsonrpc.MethodNotFoundMessage))
	}
}

func (p ScanService) PerformScan(ctx context.Context, req jsonrpc.Request) ([]byte, error) {
	curUser := sophrosyne.ExtractUser(ctx)
	if curUser == nil {
		return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
	}

	var params sophrosyne.PerformScanRequest
	err := rpc.ParamsIntoAny(&req, &params, p.validator)
	if err != nil {
		p.logger.ErrorContext(ctx, "error extracting params from request", "error", err)
		return rpc.ErrorFromRequest(&req, jsonrpc.InvalidParams, string(jsonrpc.InvalidParamsMessage))
	}

	var profile *sophrosyne.Profile
	if params.Profile != "" {
		dbp, err := p.profileService.GetProfileByName(ctx, params.Profile)
		if err != nil {
			p.logger.ErrorContext(ctx, "error getting profile by name", "profile", params.Profile, "error", err)
			return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
		}
		p.logger.DebugContext(ctx, "using profile from params for scan", "profile", params.Profile)
		profile = &dbp
	} else {
		if curUser.DefaultProfile.Name == "" {
			dbp, err := p.profileService.GetProfileByName(ctx, "default")
			if err != nil {
				p.logger.ErrorContext(ctx, "error getting default profile", "error", err)
				return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
			}
			p.logger.DebugContext(ctx, "using service-wide default profile for scan", "profile", dbp.Name)
			profile = &dbp
		} else {
			p.logger.DebugContext(ctx, "using default profile for scan", "profile", curUser.DefaultProfile.Name)
			profile = &curUser.DefaultProfile
		}
	}

	checkResults := make(map[string]checkResult)
	var success bool

	for _, check := range profile.Checks {
		p.logger.DebugContext(ctx, "running check from profile", "profile", profile.Name, "check", check.Name)
		res, err := doCheck(ctx, p.logger, check)
		if err != nil {
			p.logger.ErrorContext(ctx, "error running check", "check", check.Name, "error", err)
			return rpc.ErrorFromRequest(&req, jsonrpc.InternalError, string(jsonrpc.InternalErrorMessage))
		}
		checkResults[check.Name] = res
		if res.Status {
			success = true
		} else {
			success = false
		}
	}

	resp := struct {
		Result bool                   `json:"result"`
		Checks map[string]checkResult `json:"checks"`
	}{
		Result: success,
		Checks: checkResults,
	}

	return rpc.ResponseToRequest(&req, resp)
}

type checkResult struct {
	Status bool   `json:"status"`
	Detail string `json:"detail"`
}

func doCheck(ctx context.Context, logger *slog.Logger, check sophrosyne.Check) (checkResult, error) {
	if len(check.UpstreamServices) == 0 {
		logger.ErrorContext(ctx, "no upstream services for check", "check", check.Name)
		return checkResult{}, fmt.Errorf("missing upstream services")
	}
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient(check.UpstreamServices[0].Host, opts...)
	if err != nil {
		logger.ErrorContext(ctx, "error connecting to check", "check", check.Name, "error", err)
		return checkResult{}, err
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			logger.ErrorContext(ctx, "error closing grpc connection", "check", check.Name, "error", err)
		}
	}()
	client := checks.NewCheckServiceClient(conn)
	resp, err := client.Check(ctx, &checks.CheckRequest{Check: &checks.CheckRequest_Text{Text: "something"}})
	if err != nil {
		logger.ErrorContext(ctx, "error calling check", "check", check.Name, "error", err)
		return checkResult{}, err
	}
	return checkResult{
		Status: resp.Result,
		Detail: resp.Details,
	}, nil
}
