package sophrosyne

import (
	"context"
	"fmt"
	"time"
)

type User struct {
	ID        string
	Name      string
	Email     string
	Token     []byte
	IsAdmin   bool
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
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

type UserContextKey struct{}

type UserServiceCache struct {
	cache          *Cache
	userService    UserService
	tracingService TracingService
}

func NewUserServiceCache(config *Config, userService UserService, tracingService TracingService) *UserServiceCache {
	return &UserServiceCache{
		cache:          NewCache(config.Services.Users.CacheTTL),
		userService:    userService,
		tracingService: tracingService,
	}
}

func (c *UserServiceCache) GetUser(ctx context.Context, id string) (User, error) {
	ctx, span := c.tracingService.StartSpan(ctx, "UserServiceCache.GetUser")
	v, ok := c.cache.Get(id)
	if ok {
		span.End()
		return v.(User), nil
	}

	user, err := c.userService.GetUser(ctx, id)
	if err != nil {
		span.End()
		return User{}, err
	}

	c.cache.Set(id, user)
	span.End()
	return user, nil
}

func (c *UserServiceCache) GetUserByEmail(ctx context.Context, email string) (User, error) {
	ctx, span := c.tracingService.StartSpan(ctx, "UserServiceCache.GetUserByEmail")
	user, err := c.userService.GetUserByEmail(ctx, email)
	if err != nil {
		span.End()
		return User{}, err
	}

	c.cache.Set(user.ID, user)
	span.End()
	return user, nil
}

func (c *UserServiceCache) GetUserByName(ctx context.Context, name string) (User, error) {
	ctx, span := c.tracingService.StartSpan(ctx, "UserServiceCache.GetUserByName")
	user, err := c.userService.GetUserByName(ctx, name)
	if err != nil {
		span.End()
		return User{}, err
	}

	c.cache.Set(user.ID, user)
	span.End()
	return user, nil
}

func (c *UserServiceCache) GetUserByToken(ctx context.Context, token []byte) (User, error) {
	ctx, span := c.tracingService.StartSpan(ctx, "UserServiceCache.GetUserByToken")
	user, err := c.userService.GetUserByToken(ctx, token)
	if err != nil {
		span.End()
		return User{}, err
	}

	c.cache.Set(user.ID, user)
	span.End()
	return user, nil
}

func (c *UserServiceCache) GetUsers(ctx context.Context, cursor *DatabaseCursor) ([]User, error) {
	ctx, span := c.tracingService.StartSpan(ctx, "UserServiceCache.GetUsers")
	users, err := c.userService.GetUsers(ctx, cursor)
	if err != nil {
		span.End()
		return nil, err
	}

	for _, user := range users {
		c.cache.Set(user.ID, user)
	}

	span.End()
	return users, nil
}

func (c *UserServiceCache) CreateUser(ctx context.Context, req CreateUserRequest) (User, error) {
	ctx, span := c.tracingService.StartSpan(ctx, "UserServiceCache.CreateUser")
	user, err := c.userService.CreateUser(ctx, req)
	if err != nil {
		span.End()
		return User{}, err
	}

	c.cache.Set(user.ID, user)
	span.End()
	return user, nil
}

func (c *UserServiceCache) UpdateUser(ctx context.Context, req UpdateUserRequest) (User, error) {
	ctx, span := c.tracingService.StartSpan(ctx, "UserServiceCache.UpdateUser")
	user, err := c.userService.UpdateUser(ctx, req)
	if err != nil {
		span.End()
		return User{}, err
	}

	c.cache.Set(user.ID, user)
	span.End()
	return user, nil
}

func (c *UserServiceCache) DeleteUser(ctx context.Context, id string) error {
	ctx, span := c.tracingService.StartSpan(ctx, "UserServiceCache.DeleteUser")
	err := c.userService.DeleteUser(ctx, id)
	if err != nil {
		span.End()
		return err
	}

	c.cache.Delete(id)
	span.End()
	return nil
}

func (c *UserServiceCache) RotateToken(ctx context.Context, id string) ([]byte, error) {
	return c.userService.RotateToken(ctx, id)
}

type GetUserRequest struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (p GetUserRequest) Validate(interface{}) error {
	if p.ID == "" && p.Name == "" && p.Email == "" {
		return fmt.Errorf("one of ID, Name or Email must be provided")
	}
	if p.ID != "" && (p.Name != "" || p.Email != "") {
		return fmt.Errorf("only one of ID, Name or Email must be provided")
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
