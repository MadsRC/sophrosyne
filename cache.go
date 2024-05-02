// Copyright (c) 2024 Mads R. Havmand
//
// Part of the codebase in this file is lifted from the go-cache project by Patrick Mylund Nielsen. The original project
// can be found at https://github.com/patrickmn/go-cache. The go-cache project is licensed under the MIT License, and
// therefore so is parts of this file.
//
// --- License applicable to the go-cache project ---
//Copyright (c) 2012-2019 Patrick Mylund Nielsen and the go-cache contributors
//
//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in
//all copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
//THE SOFTWARE.
//
// --- End of license applicable to the go-cache project ---
//
// The above license is also applicable to the parts of this file that are lifted from the go-cache project. The rest
// of the file is licensed under the same license as the rest of the sophrosyne project.
//

package sophrosyne

import (
	"runtime"
	"sync"
	"time"
)

type CacheItem struct {
	Value      any
	Expiration int64
}

type Cache struct {
	expiration int64
	items      map[string]CacheItem
	lock       sync.RWMutex
	cleaner    *cacheCleaner
}

func NewCache(expiration int64) *Cache {
	c := &Cache{
		expiration: expiration,
		items:      make(map[string]CacheItem),
	}

	// Doing it this way ensures that the cacheCleaner goroutine does not keep the returned Cache object from being
	// garbage collected. When garbage collection does occur, the finalizer will stop the cacheCleaner goroutine.
	runCacheCleaner(c, time.Duration(expiration)*time.Nanosecond)
	runtime.SetFinalizer(c, stopCacheCleaner)

	return c
}

func (c *Cache) Get(key string) (any, bool) {
	c.lock.RLock()
	item, ok := c.items[key]
	if !ok {
		c.lock.RUnlock()
		return nil, false
	}
	c.lock.RUnlock()
	return item.Value, true
}

func (c *Cache) Set(key string, value any) {
	c.lock.Lock()
	c.items[key] = CacheItem{Value: value, Expiration: c.expiration}
	c.lock.Unlock()
}

func (c *Cache) Delete(key string) {
	c.lock.Lock()
	delete(c.items, key)
	c.lock.Unlock()
}

func (c *Cache) DeleteExpired() {
	now := time.Now().UnixNano()
	c.lock.Lock()
	for key, item := range c.items {
		if item.Expiration > 0 && now > item.Expiration {
			delete(c.items, key)
		}
	}
	c.lock.Unlock()
}

type cacheCleaner struct {
	interval time.Duration
	stop     chan bool
}

func (cleaner *cacheCleaner) Start(c *Cache) {
	ticker := time.NewTicker(cleaner.interval)
	for {
		select {
		case <-ticker.C:
			c.DeleteExpired()
		case <-cleaner.stop:
			ticker.Stop()
			return
		}
	}
}

func runCacheCleaner(c *Cache, interval time.Duration) {
	cleaner := &cacheCleaner{
		interval: interval,
		stop:     make(chan bool),
	}
	go cleaner.Start(c)
}

func stopCacheCleaner(c *Cache) {
	c.cleaner.stop <- true
}
