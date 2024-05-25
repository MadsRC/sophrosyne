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

type UserServiceCache struct {
	cache          *Cache
	nameToIDCache  *Cache
	emailToIDCache *Cache
	userService    sophrosyne.UserService
	tracingService sophrosyne.TracingService
}

func NewUserServiceCache(config *sophrosyne.Config, userService sophrosyne.UserService, tracingService sophrosyne.TracingService) *UserServiceCache {
	return &UserServiceCache{
		cache:          NewCache(config.Services.Users.Cache.TTL, config.Services.Users.Cache.CleanupInterval),
		nameToIDCache:  NewCache(config.Services.Users.Cache.TTL, config.Services.Users.Cache.CleanupInterval),
		emailToIDCache: NewCache(config.Services.Users.Cache.TTL, config.Services.Users.Cache.CleanupInterval),
		userService:    userService,
		tracingService: tracingService,
	}
}

func (c *UserServiceCache) GetUser(ctx context.Context, id string) (sophrosyne.User, error) {
	ctx, span := c.tracingService.StartSpan(ctx, "UserServiceCache.GetUser")
	v, ok := c.cache.Get(id)
	if ok {
		span.End()
		return v.(sophrosyne.User), nil
	}

	user, err := c.userService.GetUser(ctx, id)
	if err != nil {
		span.End()
		return sophrosyne.User{}, err
	}

	c.cache.Set(id, user)
	span.End()
	return user, nil
}

func (c *UserServiceCache) GetUserByEmail(ctx context.Context, email string) (sophrosyne.User, error) {
	ctx, span := c.tracingService.StartSpan(ctx, "UserServiceCache.GetUserByEmail")
	v, ok := c.emailToIDCache.Get(email)
	if ok {
		span.End()
		return c.GetUser(ctx, v.(string))
	}
	user, err := c.userService.GetUserByEmail(ctx, email)
	if err != nil {
		span.End()
		return sophrosyne.User{}, err
	}

	c.emailToIDCache.Set(user.Email, user.ID)
	span.End()
	return user, nil
}

func (c *UserServiceCache) GetUserByName(ctx context.Context, name string) (sophrosyne.User, error) {
	ctx, span := c.tracingService.StartSpan(ctx, "UserServiceCache.GetUserByName")
	v, ok := c.nameToIDCache.Get(name)
	if ok {
		span.End()
		return c.GetUser(ctx, v.(string))
	}
	user, err := c.userService.GetUserByName(ctx, name)
	if err != nil {
		span.End()
		return sophrosyne.User{}, err
	}

	c.nameToIDCache.Set(user.Name, user.ID)
	span.End()
	return user, nil
}

// Get the user associated with the given token.
//
// This method bypasses the cache and retrieves the user directly from the underlying service in order to ensure that
// the token is still valid.
//
// The returned user is written to the cache before being returned.
func (c *UserServiceCache) GetUserByToken(ctx context.Context, token []byte) (sophrosyne.User, error) {
	ctx, span := c.tracingService.StartSpan(ctx, "UserServiceCache.GetUserByToken")
	user, err := c.userService.GetUserByToken(ctx, token)
	if err != nil {
		span.End()
		return sophrosyne.User{}, err
	}

	c.cache.Set(user.ID, user)
	span.End()
	return user, nil
}

func (c *UserServiceCache) GetUsers(ctx context.Context, cursor *sophrosyne.DatabaseCursor) ([]sophrosyne.User, error) {
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

func (c *UserServiceCache) CreateUser(ctx context.Context, req sophrosyne.CreateUserRequest) (sophrosyne.User, error) {
	ctx, span := c.tracingService.StartSpan(ctx, "UserServiceCache.CreateUser")
	user, err := c.userService.CreateUser(ctx, req)
	if err != nil {
		span.End()
		return sophrosyne.User{}, err
	}

	c.cache.Set(user.ID, user)
	span.End()
	return user, nil
}

func (c *UserServiceCache) UpdateUser(ctx context.Context, req sophrosyne.UpdateUserRequest) (sophrosyne.User, error) {
	ctx, span := c.tracingService.StartSpan(ctx, "UserServiceCache.UpdateUser")
	user, err := c.userService.UpdateUser(ctx, req)
	if err != nil {
		span.End()
		return sophrosyne.User{}, err
	}

	c.cache.Set(user.ID, user)
	span.End()
	return user, nil
}

func (c *UserServiceCache) DeleteUser(ctx context.Context, id string) error {
	ctx, span := c.tracingService.StartSpan(ctx, "UserServiceCache.DeleteUser")
	user, err := c.userService.GetUser(ctx, id)
	if err != nil {
		span.End()
		return err
	}
	err = c.userService.DeleteUser(ctx, id)
	if err != nil {
		span.End()
		return err
	}

	c.nameToIDCache.Delete(user.Name)
	c.cache.Delete(id)
	span.End()
	return nil
}

func (c *UserServiceCache) RotateToken(ctx context.Context, id string) ([]byte, error) {
	ctx, span := c.tracingService.StartSpan(ctx, "UserServiceCache.RotateToken")
	result, err := c.userService.RotateToken(ctx, id)
	span.End()

	return result, err
}

func (c *UserServiceCache) Health(ctx context.Context) (bool, []byte) {
	_, span := c.tracingService.StartSpan(ctx, "UserServiceCache.Health")
	span.End()
	return true, []byte(`{"ok"}`)
}
