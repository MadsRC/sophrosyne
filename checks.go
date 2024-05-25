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
