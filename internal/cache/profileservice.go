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
)

// ProfileServiceCache is a cache for profiles that implements [sophrosyne.ProfileService]. It is designed to sit in
// front of another [sophrosyne.ProfileService] and only cache the result of the [sophrosyne.ProfileService].
type ProfileServiceCache struct {
	cache          *Cache // cache for profiles
	nameToIDCache  *Cache // cache for profile names to IDs.
	profileService sophrosyne.ProfileService
	tracingService sophrosyne.TracingService
}

func NewProfileServiceCache(config *sophrosyne.Config, profileService sophrosyne.ProfileService, tracingService sophrosyne.TracingService) *ProfileServiceCache {
	return &ProfileServiceCache{
		cache:          NewCache(config.Services.Profiles.Cache.TTL, config.Services.Profiles.Cache.CleanupInterval),
		nameToIDCache:  NewCache(config.Services.Profiles.Cache.TTL, config.Services.Profiles.Cache.CleanupInterval),
		profileService: profileService,
		tracingService: tracingService,
	}
}

func (p ProfileServiceCache) GetProfile(ctx context.Context, id string) (sophrosyne.Profile, error) {
	ctx, span := p.tracingService.StartSpan(ctx, "ProfileServiceCache.GetProfile")
	v, ok := p.cache.Get(id)
	if ok {
		span.End()
		return v.(sophrosyne.Profile), nil
	}

	profile, err := p.profileService.GetProfile(ctx, id)
	if err != nil {
		span.End()
		return sophrosyne.Profile{}, err
	}

	p.cache.Set(id, profile)
	span.End()
	return profile, nil
}

func (p ProfileServiceCache) GetProfileByName(ctx context.Context, name string) (sophrosyne.Profile, error) {
	ctx, span := p.tracingService.StartSpan(ctx, "ProfileServiceCache.GetProfileByName")
	id, ok := p.nameToIDCache.Get(name)
	if ok {
		span.End()
		return p.GetProfile(ctx, id.(string))
	}
	profile, err := p.profileService.GetProfileByName(ctx, name)
	if err != nil {
		span.End()
		return sophrosyne.Profile{}, err
	}

	p.nameToIDCache.Set(profile.Name, profile.ID)
	span.End()
	return profile, nil
}

func (p ProfileServiceCache) GetProfiles(ctx context.Context, cursor *sophrosyne.DatabaseCursor) ([]sophrosyne.Profile, error) {
	ctx, span := p.tracingService.StartSpan(ctx, "ProfileServiceCache.GetProfiles")
	profiles, err := p.profileService.GetProfiles(ctx, cursor)
	if err != nil {
		span.End()
		return nil, err
	}

	for _, user := range profiles {
		p.cache.Set(user.ID, user)
	}

	span.End()
	return profiles, nil
}

func (p ProfileServiceCache) CreateProfile(ctx context.Context, profile sophrosyne.CreateProfileRequest) (sophrosyne.Profile, error) {
	ctx, span := p.tracingService.StartSpan(ctx, "ProfileServiceCache.CreateProfile")
	createProfile, err := p.profileService.CreateProfile(ctx, profile)
	if err != nil {
		span.End()
		return sophrosyne.Profile{}, err
	}

	p.cache.Set(createProfile.ID, createProfile)
	span.End()
	return createProfile, nil
}

func (p ProfileServiceCache) UpdateProfile(ctx context.Context, profile sophrosyne.UpdateProfileRequest) (sophrosyne.Profile, error) {
	ctx, span := p.tracingService.StartSpan(ctx, "ProfileServiceCache.UpdateProfile")
	updateProfile, err := p.profileService.UpdateProfile(ctx, profile)
	if err != nil {
		span.End()
		return sophrosyne.Profile{}, err
	}

	p.cache.Set(updateProfile.ID, updateProfile)
	span.End()
	return updateProfile, nil
}

func (p ProfileServiceCache) DeleteProfile(ctx context.Context, name string) error {
	ctx, span := p.tracingService.StartSpan(ctx, "ProfileServiceCache.DeleteProfile")
	err := p.profileService.DeleteProfile(ctx, name)
	if err != nil {
		span.End()
		return err
	}

	p.cache.Delete(name)
	span.End()
	return nil
}
