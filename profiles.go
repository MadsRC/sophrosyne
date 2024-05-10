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
	"time"
)

type Profile struct {
	ID        string
	Name      string
	Checks    []Check
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func (p Profile) EntityType() string { return "Profile" }

func (p Profile) EntityID() string { return p.ID }

type ProfileService interface {
	GetProfile(ctx context.Context, id string) (Profile, error)
	GetProfileByName(ctx context.Context, name string) (Profile, error)
	GetProfiles(ctx context.Context, cursor *DatabaseCursor) ([]Profile, error)
	CreateProfile(ctx context.Context, profile CreateProfileRequest) (Profile, error)
	UpdateProfile(ctx context.Context, profile UpdateProfileRequest) (Profile, error)
	DeleteProfile(ctx context.Context, name string) error
}

type GetProfileRequest struct {
	ID   string `json:"id"`
	Name string `json:"name" validate:"required_without=ID,excluded_with=ID"`
}

type GetProfileResponse struct {
	Name      string   `json:"name"`
	Checks    []string `json:"checks"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`
	DeletedAt string   `json:"deletedAt,omitempty"`
}

func (r *GetProfileResponse) FromProfile(p Profile) *GetProfileResponse {
	var c []string
	for _, entry := range p.Checks {
		c = append(c, entry.Name)
	}
	r.Name = p.Name
	r.Checks = c
	r.CreatedAt = p.CreatedAt.Format(TimeFormatInResponse)
	r.UpdatedAt = p.UpdatedAt.Format(TimeFormatInResponse)
	if p.DeletedAt != nil {
		r.DeletedAt = p.DeletedAt.Format(TimeFormatInResponse)
	}
	return r
}

type GetProfilesRequest struct {
	Cursor string `json:"cursor"`
}

type GetProfilesResponse struct {
	Profiles []GetProfileResponse `json:"profiles"`
	Cursor   string               `json:"cursor"`
	Total    int                  `json:"total"`
}

type CreateProfileRequest struct {
	Name   string   `json:"name" validate:"required"`
	Checks []string `json:"checks"`
}

type CreateProfileResponse struct {
	GetProfileResponse
}

type UpdateProfileRequest struct {
	Name   string   `json:"name" validate:"required"`
	Checks []string `json:"checks"`
}

type UpdateProfileResponse struct {
	GetProfileResponse
}

type DeleteProfileRequest struct {
	Name string `json:"name" validate:"required"`
}

type ProfileServiceCache struct {
	cache          *Cache
	profileService ProfileService
	tracingService TracingService
}

func NewProfileServiceCache(config *Config, profileService ProfileService, tracingService TracingService) *ProfileServiceCache {
	return &ProfileServiceCache{
		cache:          NewCache(config.Services.Profiles.CacheTTL),
		profileService: profileService,
		tracingService: tracingService,
	}
}

func (p ProfileServiceCache) GetProfile(ctx context.Context, id string) (Profile, error) {
	ctx, span := p.tracingService.StartSpan(ctx, "ProfileServiceCache.GetProfile")
	v, ok := p.cache.Get(id)
	if ok {
		span.End()
		return v.(Profile), nil
	}

	profile, err := p.profileService.GetProfile(ctx, id)
	if err != nil {
		span.End()
		return Profile{}, err
	}

	p.cache.Set(id, profile)
	span.End()
	return profile, nil
}

func (p ProfileServiceCache) GetProfileByName(ctx context.Context, name string) (Profile, error) {
	ctx, span := p.tracingService.StartSpan(ctx, "ProfileServiceCache.GetProfileByName")
	profile, err := p.profileService.GetProfileByName(ctx, name)
	if err != nil {
		span.End()
		return Profile{}, err
	}

	p.cache.Set(profile.ID, profile)
	span.End()
	return profile, nil
}

func (p ProfileServiceCache) GetProfiles(ctx context.Context, cursor *DatabaseCursor) ([]Profile, error) {
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

func (p ProfileServiceCache) CreateProfile(ctx context.Context, profile CreateProfileRequest) (Profile, error) {
	ctx, span := p.tracingService.StartSpan(ctx, "ProfileServiceCache.CreateProfile")
	createProfile, err := p.profileService.CreateProfile(ctx, profile)
	if err != nil {
		span.End()
		return Profile{}, err
	}

	p.cache.Set(createProfile.ID, createProfile)
	span.End()
	return createProfile, nil
}

func (p ProfileServiceCache) UpdateProfile(ctx context.Context, profile UpdateProfileRequest) (Profile, error) {
	ctx, span := p.tracingService.StartSpan(ctx, "ProfileServiceCache.UpdateProfile")
	updateProfile, err := p.profileService.UpdateProfile(ctx, profile)
	if err != nil {
		span.End()
		return Profile{}, err
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
