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

package cache

import (
	"context"
	"github.com/madsrc/sophrosyne"
	"time"
)

type CheckServiceCache struct {
	cache          *Cache
	checkService   sophrosyne.CheckService
	tracingService sophrosyne.TracingService
}

func NewCheckServiceCache(config *sophrosyne.Config, checkService sophrosyne.CheckService, tracingService sophrosyne.TracingService) *CheckServiceCache {
	return &CheckServiceCache{
		cache:          NewCache(time.Duration(config.Services.Checks.CacheTTL)*time.Millisecond, 100*time.Millisecond),
		checkService:   checkService,
		tracingService: tracingService,
	}
}

func (c CheckServiceCache) GetCheck(ctx context.Context, id string) (sophrosyne.Check, error) {
	ctx, span := c.tracingService.StartSpan(ctx, "CheckServiceCache.GetCheck")
	v, ok := c.cache.Get(id)
	if ok {
		span.End()
		return v.(sophrosyne.Check), nil
	}

	profile, err := c.checkService.GetCheck(ctx, id)
	if err != nil {
		span.End()
		return sophrosyne.Check{}, err
	}

	c.cache.Set(id, profile)
	span.End()
	return profile, nil
}

func (c CheckServiceCache) GetCheckByName(ctx context.Context, name string) (sophrosyne.Check, error) {
	ctx, span := c.tracingService.StartSpan(ctx, "CheckServiceCache.GetCheckByName")
	profile, err := c.checkService.GetCheckByName(ctx, name)
	if err != nil {
		span.End()
		return sophrosyne.Check{}, err
	}

	c.cache.Set(profile.ID, profile)
	span.End()
	return profile, nil
}

func (c CheckServiceCache) GetChecks(ctx context.Context, cursor *sophrosyne.DatabaseCursor) ([]sophrosyne.Check, error) {
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

func (c CheckServiceCache) CreateCheck(ctx context.Context, check sophrosyne.CreateCheckRequest) (sophrosyne.Check, error) {
	ctx, span := c.tracingService.StartSpan(ctx, "CheckServiceCache.CreateCheck")
	createProfile, err := c.checkService.CreateCheck(ctx, check)
	if err != nil {
		span.End()
		return sophrosyne.Check{}, err
	}

	c.cache.Set(createProfile.ID, createProfile)
	span.End()
	return createProfile, nil
}

func (c CheckServiceCache) UpdateCheck(ctx context.Context, check sophrosyne.UpdateCheckRequest) (sophrosyne.Check, error) {
	ctx, span := c.tracingService.StartSpan(ctx, "CheckServiceCache.UpdateCheck")
	updateProfile, err := c.checkService.UpdateCheck(ctx, check)
	if err != nil {
		span.End()
		return sophrosyne.Check{}, err
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
