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

// Part of the codebase in this file is lifted from the go-cache project (commit
// 46f407853014144407b6c2ec7ccc76bf67958d93) by Patrick Mylund Nielsen. The original project
// can be found at https://github.com/patrickmn/go-cache. The go-cache project is licensed under the MIT License, and
// therefore so is parts of this file.
//
// --- License applicable to the go-cache project ---
// Copyright (c) 2012-2019 Patrick Mylund Nielsen and the go-cache contributors
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
//
// --- End of license applicable to the go-cache project ---

package cache

import (
	"runtime"
	"sync"
	"time"
)

const DefaultExpiration = 100 * time.Millisecond

type cacheItem struct {
	ExpiresAt time.Time
	Value     any
}

type Cache struct {
	*cache
}

type cache struct {
	items   map[string]cacheItem
	lock    *sync.RWMutex
	exp     time.Duration
	cleaner *cleaner
}

// NewCache creates a new cache with the given expiration time and cleaning interval.
//
// If the cleaning interval is 0, a nil cache is returned.
//
// If the expiration time is 0 or less, [DefaultExpiration] will be used.
func NewCache(exp time.Duration, cleanerInterval time.Duration) *Cache {
	if cleanerInterval <= 0 {
		return nil
	}

	if exp <= 0 {
		exp = DefaultExpiration
	}

	c := &cache{
		items: make(map[string]cacheItem),
		lock:  &sync.RWMutex{},
		exp:   exp,
	}

	// Doing it this way ensures that the cleaner goroutine does not keep the returned Cache object from being
	// garbage collected. When garbage collection does occur, the finalizer will stop the cleaner goroutine.
	C := &Cache{c}
	runCleaner(c, cleanerInterval)
	runtime.SetFinalizer(C, stopCleaner)

	return C
}

// Set sets the value of the item in the cache with the given key.
func (c *cache) Set(key string, value any) {
	c.lock.Lock()
	c.items[key] = cacheItem{ExpiresAt: time.Now().Add(c.exp), Value: value}
	c.lock.Unlock()
}

// Get retrieves the value associated with the given key.
func (c *cache) Get(key string) (any, bool) {
	c.lock.RLock()
	item, ok := c.items[key]
	if !ok {
		c.lock.RUnlock()
		return nil, false
	}
	c.lock.RUnlock()
	return item.Value, true
}

// Delete removes the item with the specified key from the cache.
func (c *cache) Delete(key string) {
	c.lock.Lock()
	delete(c.items, key)
	c.lock.Unlock()
}

// Expire removes expired items from the cache.
//
// It iterates over the items in the cache and deletes any item whose expiration time is before the current time.
// The function does not take any parameters.
// It does not return any values.
func (c *cache) Expire() {
	now := time.Now()
	c.lock.Lock()
	for key, item := range c.items {
		if item.ExpiresAt.Before(now) {
			delete(c.items, key)
		}
	}
	c.lock.Unlock()
}

type cleaner struct {
	interval time.Duration
	stop     chan struct{}
}

// start starts the cleaner goroutine for the given cache.
//
// It takes a pointer to a cache object as a parameter.
// The function does not return any value.
func (j *cleaner) start(c *cache) {
	ticker := time.NewTicker(j.interval)
	for {
		select {
		case <-ticker.C:
			c.Expire()
		case <-j.stop:
			ticker.Stop()
			return
		}
	}
}

// runCleaner will create and start a new cleaner goroutine for the given cache using the given interval.
func runCleaner(c *cache, interval time.Duration) {
	j := &cleaner{
		interval: interval,
		stop:     make(chan struct{}),
	}
	c.cleaner = j
	go j.start(c)
}

// stopCleaner will send a signal to stop the cleaner goroutine for the given cache.
func stopCleaner(c *Cache) {
	c.cleaner.stop <- struct{}{}
}
