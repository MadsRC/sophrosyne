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
	"fmt"
	"log/slog"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/madsrc/sophrosyne"
	v0 "github.com/madsrc/sophrosyne/internal/grpc/sophrosyne/v0"
	"github.com/madsrc/sophrosyne/internal/validator"
)

// ScanServiceServer is a gGRPC server that handles scans.
type ScanServiceServer struct {
	v0.UnimplementedScanServiceServer
	logger         *slog.Logger              `validate:"required"`
	config         *sophrosyne.Config        `validate:"required"`
	validator      sophrosyne.Validator      `validate:"required"`
	profileService sophrosyne.ProfileService `validate:"required"`
}

// NewScanServiceServer returns a new ScanServiceServer instance.
//
// If the provided options are invalid, an error will be returned.
// Required options are marked with the 'validate:"required"' tag in
// the [ScanServiceServer] struct. Every required option has a
// corresponding [Option] function.
//
// If no [sophrosyne.Validator] is provided, a default one will be
// created.
func NewScanServiceServer(ctx context.Context, opts ...Option) (*ScanServiceServer, error) {
	s := &ScanServiceServer{}
	setOptions(s, defaultScanServiceServerOptions(), opts...)

	if s.logger != nil {
		s.logger.DebugContext(ctx, "validating server options")
	}
	err := s.validator.Validate(s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func defaultScanServiceServerOptions() []Option {
	return []Option{
		WithValidator(validator.NewValidator()),
	}
}

func (s ScanServiceServer) Scan(ctx context.Context, request *v0.ScanRequest) (*v0.ScanResponse, error) {
	curUser := sophrosyne.ExtractUser(ctx)
	if curUser == nil {
		return nil, status.Errorf(codes.Unauthenticated, InvalidTokenMsg)
	}

	profile, err := s.lookupProfile(ctx, request, curUser)
	if err != nil {
		return nil, fmt.Errorf("error looking up profile: %w", err)
	}
	return s.performScan(ctx, request, profile)
}

// lookupProfile takes a request and a user and returns a profile.
//
// If the request contains the name of a profile, that profile will be fetched
// from the profile service and returned.
// If the request does not contain a profile name, the user's default profile
// will be fetched from the profile service and returned.
// If the user does not have a default profile, the server-wide default profile
// will be fetched from the profile service and returned.
// If, for any reason, the profile cannot be retrieved, an error will be returned.
func (s ScanServiceServer) lookupProfile(ctx context.Context, req *v0.ScanRequest, curUser *sophrosyne.User) (*sophrosyne.Profile, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if curUser == nil {
		return nil, fmt.Errorf("curUser cannot be nil")
	}
	var profile *sophrosyne.Profile
	if req.GetProfile() != "" {
		dbp, err := s.profileService.GetProfileByName(ctx, req.GetProfile())
		if err != nil {
			return nil, fmt.Errorf("error getting profile by name: %v", err)
		}
		s.logger.DebugContext(ctx, "using profile from request for scan", "profile", req.Profile)
		profile = &dbp
	} else {
		if curUser.DefaultProfile.Name == "" {
			dbp, err := s.profileService.GetProfileByName(ctx, "default")
			if err != nil {
				return nil, fmt.Errorf("error getting default profile: %v", err)
			}
			s.logger.DebugContext(ctx, "using service-wide default profile for scan", "profile", dbp.Name)
			profile = &dbp
		} else {
			s.logger.DebugContext(ctx, "using default profile for scan", "profile", curUser.DefaultProfile.Name)
			profile = &curUser.DefaultProfile
		}
	}

	return profile, nil
}

func (s ScanServiceServer) performScan(ctx context.Context, req *v0.ScanRequest, profile *sophrosyne.Profile) (*v0.ScanResponse, error) {
	messages := make(chan *v0.CheckResult, len(profile.Checks))
	var wg sync.WaitGroup
	wg.Add(len(profile.Checks))

	for _, check := range profile.Checks {
		s.logger.DebugContext(ctx, "running check from profile", "profile", profile.Name, "check", check.Name)
		go func(check sophrosyne.Check) {
			defer wg.Done()
			res, err := doCheck(ctx, s.logger, check, nil, req)
			if err != nil {
				s.logger.ErrorContext(ctx, "error running check", "check", check.Name, "error", err)
			}

			messages <- res
		}(check)
	}

	wg.Wait()
	close(messages)

	results := processCheckResults(ctx, messages, s.logger)

	return &v0.ScanResponse{
		Result: results.Result,
		Checks: results.Checks,
	}, nil
}

func doCheck(ctx context.Context, log *slog.Logger, check sophrosyne.Check, client v0.CheckProviderServiceClient, req *v0.ScanRequest) (*v0.CheckResult, error) {
	if len(check.UpstreamServices) == 0 {
		log.DebugContext(ctx, "no upstream services for check", "check", check.Name)
		return nil, fmt.Errorf("missing upstream services")
	}

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient(check.UpstreamServices[0].Host, opts...)
	if err != nil {
		log.DebugContext(ctx, "error connecting to check", "check", check.Name, "error", err)
		return nil, err
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			log.ErrorContext(ctx, "error closing grpc connection", "check", check.Name, "error", err)
		}
	}()

	if client == nil {
		client = v0.NewCheckProviderServiceClient(conn)
	}

	outReq, err := checkProviderRequestFromScanRequest(ctx, log, req)
	if err != nil {
		log.DebugContext(ctx, "error creating check request", "check", check.Name, "error", err)
		return nil, err
	}

	resp, err := client.Check(ctx, outReq)
	if err != nil {
		log.DebugContext(ctx, "error calling check", "check", check.Name, "error", err)
		return nil, err
	}
	log.DebugContext(ctx, "finished calling upstream service", "check", check.Name, "result", resp)
	return &v0.CheckResult{
		Name:    check.Name,
		Result:  resp.Result,
		Details: resp.Details,
	}, nil
}

type processedCheckResults struct {
	Result bool
	Checks []*v0.CheckResult
}

func processCheckResults(ctx context.Context, messages chan *v0.CheckResult, logger *slog.Logger) processedCheckResults {
	logger.DebugContext(ctx, "processing results from checks")
	checkResults := make([]*v0.CheckResult, 0)
	var success bool
	for msg := range messages {
		logger.DebugContext(ctx, "receiving check result", "check", msg.Name, "check_result", msg)
		if msg.Name == "" {
			logger.DebugContext(ctx, "ignoring check result")
			continue
		}
		if msg.Result {
			success = true
		} else {
			success = false
		}

		checkResults = append(checkResults, msg)
	}

	resp := processedCheckResults{
		Result: success,
		Checks: checkResults,
	}
	logger.DebugContext(ctx, "finished processing results from checks", "result", resp)

	return resp
}

func checkProviderRequestFromScanRequest(ctx context.Context, logger *slog.Logger, req *v0.ScanRequest) (*v0.CheckProviderRequest, error) {
	if len(req.GetImage()) != 0 {
		logger.DebugContext(ctx, "creating check request for image", "image", req.GetImage())
		return &v0.CheckProviderRequest{Check: &v0.CheckProviderRequest_Image{Image: req.GetImage()}}, nil
	}
	if req.GetText() != "" {
		logger.DebugContext(ctx, "creating check request for text", "text", req.GetText())
		return &v0.CheckProviderRequest{Check: &v0.CheckProviderRequest_Text{Text: req.GetText()}}, nil
	}
	return nil, nil
}
