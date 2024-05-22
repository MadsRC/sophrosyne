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
	"sync"
	"testing"

	"github.com/stretchr/testify/mock"

	sophrosyne2 "github.com/madsrc/sophrosyne/internal/mocks"
)

type commonTestStuff struct {
	t              *testing.T
	ctx            context.Context
	tracingService *sophrosyne2.MockTracingService
	profileService *sophrosyne2.MockProfileService
	checkService   *sophrosyne2.MockCheckService
	userService    *sophrosyne2.MockUserService
	span           *sophrosyne2.MockSpan
}

func setupTestStuff(t *testing.T, cts *commonTestStuff) *commonTestStuff {
	t.Helper()

	if cts == nil {
		cts = &commonTestStuff{}
	}

	cts.t = t

	if cts.ctx == nil {
		cts.ctx = context.Background()
	}

	if cts.span == nil {
		cts.span = sophrosyne2.NewMockSpan(t)
		cts.span.On("End").Once().Return(nil)
	}

	if cts.tracingService == nil {
		cts.tracingService = sophrosyne2.NewMockTracingService(t)
		cts.tracingService.On("StartSpan", cts.ctx, mock.Anything).Once().Return(cts.ctx, cts.span)
	}

	if cts.profileService == nil {
		cts.profileService = sophrosyne2.NewMockProfileService(t)
	}

	if cts.checkService == nil {
		cts.checkService = sophrosyne2.NewMockCheckService(t)
	}

	if cts.userService == nil {
		cts.userService = sophrosyne2.NewMockUserService(t)
	}

	t.Cleanup(cts.tearDown)

	return cts
}

func (cts *commonTestStuff) tearDown() {
	cts.span.AssertExpectations(cts.t)
	cts.tracingService.AssertExpectations(cts.t)
	cts.profileService.AssertExpectations(cts.t)
	cts.checkService.AssertExpectations(cts.t)
	cts.userService.AssertExpectations(cts.t)
}

func getProfileServiceCache(t *testing.T, cts *commonTestStuff) *ProfileServiceCache {
	t.Helper()
	profileServiceCache := ProfileServiceCache{
		cache:          &Cache{&cache{items: make(map[string]cacheItem), lock: new(sync.RWMutex)}},
		nameToIDCache:  &Cache{&cache{items: make(map[string]cacheItem), lock: new(sync.RWMutex)}},
		profileService: cts.profileService,
		tracingService: cts.tracingService,
	}
	return &profileServiceCache
}

func getUserServiceCache(t *testing.T, cts *commonTestStuff) *UserServiceCache {
	t.Helper()
	userServiceCache := UserServiceCache{
		cache:          &Cache{&cache{items: make(map[string]cacheItem), lock: new(sync.RWMutex)}},
		nameToIDCache:  &Cache{&cache{items: make(map[string]cacheItem), lock: new(sync.RWMutex)}},
		userService:    cts.userService,
		tracingService: cts.tracingService,
	}
	return &userServiceCache
}

func getCheckServiceCache(t *testing.T, cts *commonTestStuff) *CheckServiceCache {
	t.Helper()
	checkServiceCache := CheckServiceCache{
		cache:          &Cache{&cache{items: make(map[string]cacheItem), lock: new(sync.RWMutex)}},
		nameToIDCache:  &Cache{&cache{items: make(map[string]cacheItem), lock: new(sync.RWMutex)}},
		checkService:   cts.checkService,
		tracingService: cts.tracingService,
	}
	return &checkServiceCache
}
