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

type User struct {
	ID             string
	Name           string
	Email          string
	Token          []byte
	IsAdmin        bool
	DefaultProfile Profile
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}

func (u User) EntityType() string {
	return "User"
}

func (u User) EntityID() string {
	return u.ID
}

type UserService interface {
	GetUser(ctx context.Context, id string) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUserByName(ctx context.Context, name string) (User, error)
	GetUserByToken(ctx context.Context, token []byte) (User, error)
	// Returns a list of users less than, or equal to, the configured page size.
	// Configuration of the page size is an implementation detail, but should be
	// derived from [Config.Services.Users.PageSize].
	//
	// The cursor is used to paginate the results. [DatabaseCursor.Position] is
	// treated as the last read ID. If the cursor is nil, the first page of
	// results should be returned.
	//
	// When the users have been read, but before returning the list of users,
	// the cursor must be advanced to the ID of the last user returned. If no
	// users are returned, or if it is known that a subsequent call would return
	// zero users, the cursors Reset method must be called.
	//
	// The returned list of users should be ordered by ID in ascending order.
	GetUsers(ctx context.Context, cursor *DatabaseCursor) ([]User, error)
	CreateUser(ctx context.Context, user CreateUserRequest) (User, error)
	UpdateUser(ctx context.Context, user UpdateUserRequest) (User, error)
	DeleteUser(ctx context.Context, name string) error
	RotateToken(ctx context.Context, name string) ([]byte, error)
}

type GetUserRequest struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (p GetUserRequest) Validate(interface{}) error {
	if p.ID == "" && p.Name == "" && p.Email == "" {
		return NewValidationError("one of ID, Name or Email must be provided")
	}
	if p.ID != "" && (p.Name != "" || p.Email != "") {
		return NewValidationError("only one of ID, Name or Email must be provided")
	}
	return nil
}

type GetUserResponse struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	IsAdmin   bool   `json:"is_admin"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	DeletedAt string `json:"deleted_at,omitempty"`
}

func (r *GetUserResponse) FromUser(u User) *GetUserResponse {
	r.Name = u.Name
	r.Email = u.Email
	r.IsAdmin = u.IsAdmin
	r.CreatedAt = u.CreatedAt.Format(TimeFormatInResponse)
	r.UpdatedAt = u.UpdatedAt.Format(TimeFormatInResponse)
	if u.DeletedAt != nil {
		r.DeletedAt = u.DeletedAt.Format(TimeFormatInResponse)
	}

	return r
}

type GetUsersRequest struct {
	Cursor string `json:"cursor"`
}

type GetUsersResponse struct {
	Users  []GetUserResponse `json:"users"`
	Cursor string            `json:"cursor"`
	Total  int               `json:"total"`
}

type CreateUserRequest struct {
	Name    string `json:"name" validate:"required"`
	Email   string `json:"email" validate:"required"`
	IsAdmin bool   `json:"is_admin"`
}

type CreateUserResponse struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Token     []byte `json:"token"`
	IsAdmin   bool   `json:"is_admin"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	DeletedAt string `json:"deleted_at,omitempty"`
}

func (r *CreateUserResponse) FromUser(u User) *CreateUserResponse {
	r.Name = u.Name
	r.Email = u.Email
	r.Token = u.Token
	r.IsAdmin = u.IsAdmin
	r.CreatedAt = u.CreatedAt.Format(TimeFormatInResponse)
	r.UpdatedAt = u.UpdatedAt.Format(TimeFormatInResponse)
	if u.DeletedAt != nil {
		r.DeletedAt = u.DeletedAt.Format(TimeFormatInResponse)
	}

	return r
}

type UpdateUserRequest struct {
	Name    string `json:"name" validate:"required"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"is_admin"`
}

type UpdateUserResponse struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	IsAdmin   bool   `json:"is_admin"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	DeletedAt string `json:"deleted_at,omitempty"`
}

func (r *UpdateUserResponse) FromUser(u User) *UpdateUserResponse {
	r.Name = u.Name
	r.Email = u.Email
	r.IsAdmin = u.IsAdmin
	r.CreatedAt = u.CreatedAt.Format(TimeFormatInResponse)
	r.UpdatedAt = u.UpdatedAt.Format(TimeFormatInResponse)
	if u.DeletedAt != nil {
		r.DeletedAt = u.DeletedAt.Format(TimeFormatInResponse)
	}

	return r
}

type DeleteUserRequest struct {
	Name string `json:"name" validate:"required"`
}

type RotateTokenRequest struct {
	Name string `json:"name" validate:"required"`
}

type RotateTokenResponse struct {
	Token []byte `json:"token"`
}

func (r *RotateTokenResponse) FromUser(u User) *RotateTokenResponse {
	r.Token = u.Token

	return r
}

type UserContextKey struct{}
