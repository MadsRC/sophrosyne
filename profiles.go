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
