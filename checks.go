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

package sophrosyne

import (
	"context"
	"net/url"
	"time"
)

type Check struct {
	ID               string
	Name             string
	Profiles         []Profile
	UpstreamServices []url.URL
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

func (c Check) EntityType() string { return "Check" }

func (c Check) EntityID() string { return c.ID }

type CheckService interface {
	GetCheck(ctx context.Context, id string) (Check, error)
	GetCheckByName(ctx context.Context, name string) (Check, error)
	GetChecks(ctx context.Context, cursor *DatabaseCursor) ([]Check, error)
	CreateCheck(ctx context.Context, check CreateCheckRequest) (Check, error)
	UpdateCheck(ctx context.Context, check UpdateCheckRequest) (Check, error)
	DeleteCheck(ctx context.Context, id string) error
}

type GetCheckRequest struct {
	ID   string `json:"id"`
	Name string `json:"name" validate:"required_without=ID,excluded_with=ID"`
}

type GetCheckResponse struct {
	Name             string   `json:"name"`
	Profiles         []string `json:"profiles"`
	UpstreamServices []string `json:"upstream_services"`
	CreatedAt        string   `json:"createdAt"`
	UpdatedAt        string   `json:"updatedAt"`
	DeletedAt        string   `json:"deletedAt,omitempty"`
}

func (r *GetCheckResponse) FromCheck(c Check) *GetCheckResponse {
	var p []string
	for _, entry := range c.Profiles {
		p = append(p, entry.Name)
	}
	var u []string
	for _, entry := range c.UpstreamServices {
		u = append(u, entry.String())
	}
	r.Name = c.Name
	r.Profiles = p
	r.UpstreamServices = u
	r.CreatedAt = c.CreatedAt.Format(TimeFormatInResponse)
	r.UpdatedAt = c.UpdatedAt.Format(TimeFormatInResponse)
	if c.DeletedAt != nil {
		r.DeletedAt = c.DeletedAt.Format(TimeFormatInResponse)
	}
	return r
}

type GetChecksRequest struct {
	Cursor string `json:"cursor"`
}

type GetChecksResponse struct {
	Checks []GetCheckResponse `json:"checks"`
	Cursor string             `json:"cursor"`
	Total  int                `json:"total"`
}

type CreateCheckRequest struct {
	Name             string   `json:"name" validate:"required"`
	Profiles         []string `json:"profiles"`
	UpstreamServices []string `json:"upstream_services" validate:"dive,url"`
}

type CreateCheckResponse struct {
	GetCheckResponse
}

type UpdateCheckRequest struct {
	Name             string   `json:"name" validate:"required"`
	Profiles         []string `json:"profiles"`
	UpstreamServices []string `json:"upstream_services" validate:"url"`
}

type UpdateCheckResponse struct {
	GetCheckResponse
}

type DeleteCheckRequest struct {
	Name string `json:"name" validate:"required"`
}

type CheckServiceCache struct {
	cache          *Cache
	checkService   CheckService
	tracingService TracingService
}

func NewCheckServiceCache(config *Config, checkService CheckService, tracingService TracingService) *CheckServiceCache {
	return &CheckServiceCache{
		cache:          NewCache(config.Services.Profiles.CacheTTL),
		checkService:   checkService,
		tracingService: tracingService,
	}
}

func (c CheckServiceCache) GetCheck(ctx context.Context, id string) (Check, error) {
	ctx, span := c.tracingService.StartSpan(ctx, "CheckServiceCache.GetCheck")
	v, ok := c.cache.Get(id)
	if ok {
		span.End()
		return v.(Check), nil
	}

	profile, err := c.checkService.GetCheck(ctx, id)
	if err != nil {
		span.End()
		return Check{}, err
	}

	c.cache.Set(id, profile)
	span.End()
	return profile, nil
}

func (c CheckServiceCache) GetCheckByName(ctx context.Context, name string) (Check, error) {
	ctx, span := c.tracingService.StartSpan(ctx, "CheckServiceCache.GetCheckByName")
	profile, err := c.checkService.GetCheckByName(ctx, name)
	if err != nil {
		span.End()
		return Check{}, err
	}

	c.cache.Set(profile.ID, profile)
	span.End()
	return profile, nil
}

func (c CheckServiceCache) GetChecks(ctx context.Context, cursor *DatabaseCursor) ([]Check, error) {
	ctx, span := c.tracingService.StartSpan(ctx, "CheckServiceCache.GetChecks")
	profiles, err := c.checkService.GetChecks(ctx, cursor)
	if err != nil {
		span.End()
		return nil, err
	}

	for _, user := range profiles {
		c.cache.Set(user.ID, user)
	}

	span.End()
	return profiles, nil
}

func (c CheckServiceCache) CreateCheck(ctx context.Context, check CreateCheckRequest) (Check, error) {
	ctx, span := c.tracingService.StartSpan(ctx, "CheckServiceCache.CreateCheck")
	createProfile, err := c.checkService.CreateCheck(ctx, check)
	if err != nil {
		span.End()
		return Check{}, err
	}

	c.cache.Set(createProfile.ID, createProfile)
	span.End()
	return createProfile, nil
}

func (c CheckServiceCache) UpdateCheck(ctx context.Context, check UpdateCheckRequest) (Check, error) {
	ctx, span := c.tracingService.StartSpan(ctx, "CheckServiceCache.UpdateCheck")
	updateProfile, err := c.checkService.UpdateCheck(ctx, check)
	if err != nil {
		span.End()
		return Check{}, err
	}

	c.cache.Set(updateProfile.ID, updateProfile)
	span.End()
	return updateProfile, nil
}

func (c CheckServiceCache) DeleteCheck(ctx context.Context, id string) error {
	ctx, span := c.tracingService.StartSpan(ctx, "CheckServiceCache.DeleteCheck")
	err := c.checkService.DeleteCheck(ctx, id)
	if err != nil {
		span.End()
		return err
	}

	c.cache.Delete(id)
	span.End()
	return nil
}
