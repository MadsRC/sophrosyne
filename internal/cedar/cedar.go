package cedar

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/cedar-policy/cedar-go"
	"github.com/madsrc/sophrosyne"
	"log/slog"
	"sync"
)

//go:embed policies.cedar
var Policies []byte

func UserToEntity(u sophrosyne.User) cedar.Entity {
	out := cedar.Entity{
		UID: cedar.EntityUID{Type: u.EntityType(), ID: u.EntityID()},
		Attributes: cedar.Record{
			"id":         cedar.String(u.ID),
			"name":       cedar.String(u.Name),
			"email":      cedar.String(u.Email),
			"is_admin":   cedar.Boolean(u.IsAdmin),
			"created_at": cedar.Long(u.CreatedAt.Unix()),
			"updated_at": cedar.Long(u.UpdatedAt.Unix()),
		},
	}
	if u.DeletedAt != nil {
		out.Attributes["deleted_at"] = cedar.Long(u.DeletedAt.Unix())
	}
	return out
}

func ProfileToEntity(u sophrosyne.Profile) cedar.Entity {
	out := cedar.Entity{
		UID: cedar.EntityUID{Type: u.EntityType(), ID: u.EntityID()},
		Attributes: cedar.Record{
			"id":         cedar.String(u.ID),
			"name":       cedar.String(u.Name),
			"created_at": cedar.Long(u.CreatedAt.Unix()),
			"updated_at": cedar.Long(u.UpdatedAt.Unix()),
		},
	}
	if u.DeletedAt != nil {
		out.Attributes["deleted_at"] = cedar.Long(u.DeletedAt.Unix())
	}
	return out
}

func CheckToEntity(u sophrosyne.Check) cedar.Entity {
	out := cedar.Entity{
		UID: cedar.EntityUID{Type: u.EntityType(), ID: u.EntityID()},
		Attributes: cedar.Record{
			"id":         cedar.String(u.ID),
			"name":       cedar.String(u.Name),
			"created_at": cedar.Long(u.CreatedAt.Unix()),
			"updated_at": cedar.Long(u.UpdatedAt.Unix()),
		},
	}
	if u.DeletedAt != nil {
		out.Attributes["deleted_at"] = cedar.Long(u.DeletedAt.Unix())
	}
	return out
}

type AuthorizationProvider struct {
	policySet      cedar.PolicySet
	psMutex        *sync.RWMutex
	logger         *slog.Logger
	userService    sophrosyne.UserService
	profileService sophrosyne.ProfileService
	checkService   sophrosyne.CheckService
	tracingService sophrosyne.TracingService
}

func NewAuthorizationProvider(ctx context.Context, logger *slog.Logger, userService sophrosyne.UserService, tracingService sophrosyne.TracingService, profileService sophrosyne.ProfileService, checkService sophrosyne.CheckService) (*AuthorizationProvider, error) {
	ap := AuthorizationProvider{
		logger:         logger,
		userService:    userService,
		profileService: profileService,
		checkService:   checkService,
		tracingService: tracingService,
	}
	ap.psMutex = &sync.RWMutex{}
	err := ap.RefreshPolicies(ctx, Policies)
	if err != nil {
		return nil, err
	}
	return &ap, nil
}

func (a *AuthorizationProvider) RefreshPolicies(ctx context.Context, b []byte) error {
	ps, err := cedar.NewPolicySet("policies.cedar", b)
	if err != nil {
		a.logger.DebugContext(ctx, "error refreshing policies", "error", err.Error())
		return err
	}
	a.psMutex.Lock()
	defer a.psMutex.Unlock()
	a.policySet = ps
	return nil
}

func (a *AuthorizationProvider) fetchEntities(ctx context.Context, req cedar.Request) (cedar.Entities, error) {
	var principal cedar.Entity
	var resource cedar.Entity

	pri, err := a.userService.GetUser(ctx, req.Principal.ID)
	if err != nil {
		return nil, err
	}

	if !req.Resource.IsZero() {
		switch req.Resource.Type {
		case "User":
			res, err := a.userService.GetUser(ctx, req.Resource.ID)
			if err != nil {
				return nil, err
			}
			resource = UserToEntity(res)
		case "Profile":
			res, err := a.profileService.GetProfile(ctx, req.Resource.ID)
			if err != nil {
				return nil, err
			}
			resource = ProfileToEntity(res)
		case "Check":
			res, err := a.checkService.GetCheck(ctx, req.Resource.ID)
			if err != nil {
				return nil, err
			}
			resource = CheckToEntity(res)
		default:
			return nil, fmt.Errorf("unknown resource type: %s", req.Resource.Type)
		}

	}

	principal = UserToEntity(pri)

	entities := cedar.Entities{
		principal.UID: principal,
	}
	if !resource.UID.IsZero() {
		entities[resource.UID] = resource
	}

	a.logger.DebugContext(ctx, "fetched entities", "entities", entities)

	return entities, nil
}

func (a *AuthorizationProvider) IsAuthorized(ctx context.Context, req sophrosyne.AuthorizationRequest) bool {
	ctx, span := a.tracingService.StartSpan(ctx, "AuthorizationProvider.IsAuthorized")
	defer span.End()
	reqCtx, err := contextToRecord(req.Context)
	if err != nil {
		a.logger.InfoContext(ctx, "error converting context to record", "error", err.Error())
		return false
	}

	cReq := cedar.Request{
		Principal: cedar.NewEntityUID(req.Principal.EntityType(), req.Principal.EntityID()),
		Action:    cedar.NewEntityUID(req.Action.EntityType(), req.Action.EntityID()),
		Context:   *reqCtx,
	}
	if req.Resource != nil {
		cReq.Resource = cedar.NewEntityUID(req.Resource.EntityType(), req.Resource.EntityID())
	}
	entities, err := a.fetchEntities(ctx, cReq)
	if err != nil {
		a.logger.InfoContext(ctx, "error fetching entities", "error", err.Error())
		return false
	}

	a.psMutex.RLock()
	defer a.psMutex.RUnlock()
	a.logger.DebugContext(ctx, "checking authorization", "request", cReq)
	decision, diag := a.policySet.IsAuthorized(entities, cReq)
	a.logger.InfoContext(ctx, "authorization decision", "decision", decision, "diag", diag)
	return decision == cedar.Allow
}

func contextToRecord(in map[string]interface{}) (*cedar.Record, error) {
	b, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	var out cedar.Record
	err = json.Unmarshal(b, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
