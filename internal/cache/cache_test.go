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
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCacheTimes(t *testing.T) {
	var found bool

	tc := NewCache(50*time.Millisecond, 1*time.Millisecond)
	require.NotNil(t, tc)
	tc.Set("a", 1)

	_, found = tc.Get("a")
	require.True(t, found, "Did not find a when it should have been found")

	_, found = tc.Get("b")
	require.False(t, found, "Found b when it should not have been found")

	<-time.After(45 * time.Millisecond)
	_, found = tc.Get("a")
	require.True(t, found, "Did not find a when it should have been found")

	<-time.After(55 * time.Millisecond)
	_, found = tc.Get("a")
	require.False(t, found, "Found a when it should have been deleted")
}

func TestNewCacheZeroInterval(t *testing.T) {
	tc := NewCache(0, 0)
	require.NotNil(t, tc)
	require.Nil(t, tc.cleaner)
}

func TestNewCacheDefaultExpiration(t *testing.T) {
	tc := NewCache(0, 100*time.Millisecond)
	require.NotNil(t, tc)
	require.Equal(t, DefaultExpiration, tc.exp)
}

func TestDelete(t *testing.T) {
	tc := NewCache(10*time.Second, 1*time.Second)
	tc.Set("foo", "bar")
	tc.Delete("foo")
	x, found := tc.Get("foo")
	require.False(t, found, "foo was found, but it should have been deleted")
	require.Nilf(t, x, "x was not nil, got %v", x)
}

func BenchmarkCacheGetExpiring(b *testing.B) {
	b.Run("100x100", func(b *testing.B) {
		benchmarkCacheGet(b, 100*time.Millisecond, 100*time.Millisecond)
	})
	b.Run("100x50", func(b *testing.B) {
		benchmarkCacheGet(b, 100*time.Millisecond, 50*time.Millisecond)
	})
	b.Run("100x25", func(b *testing.B) {
		benchmarkCacheGet(b, 100*time.Millisecond, 25*time.Millisecond)
	})
	b.Run("100x150", func(b *testing.B) {
		benchmarkCacheGet(b, 100*time.Millisecond, 150*time.Millisecond)
	})
	b.Run("100x200", func(b *testing.B) {
		benchmarkCacheGet(b, 100*time.Millisecond, 200*time.Millisecond)
	})
}

func benchmarkCacheGet(b *testing.B, exp time.Duration, cleaningInterval time.Duration) {
	b.StopTimer()
	tc := NewCache(exp, cleaningInterval)
	tc.Set("foo", "bar")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.Get("foo")
	}
}

func BenchmarkCacheSetExpiring(b *testing.B) {
	b.Run("100x100", func(b *testing.B) {
		benchmarkCacheSet(b, 100*time.Millisecond, 100*time.Millisecond)
	})
	b.Run("100x50", func(b *testing.B) {
		benchmarkCacheSet(b, 100*time.Millisecond, 50*time.Millisecond)
	})
	b.Run("100x25", func(b *testing.B) {
		benchmarkCacheSet(b, 100*time.Millisecond, 25*time.Millisecond)
	})
	b.Run("100x150", func(b *testing.B) {
		benchmarkCacheSet(b, 100*time.Millisecond, 150*time.Millisecond)
	})
	b.Run("100x200", func(b *testing.B) {
		benchmarkCacheSet(b, 100*time.Millisecond, 200*time.Millisecond)
	})
}

func benchmarkCacheSet(b *testing.B, exp time.Duration, cleaningInterval time.Duration) {
	b.StopTimer()
	tc := NewCache(exp, cleaningInterval)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.Set("foo", "bar")
	}
}

func TestStartMethodReceivesStopSignal(t *testing.T) {
	interval := 1 * time.Second
	c := &cache{
		items: map[string]cacheItem{
			"key1": {Value: "value1", ExpiresAt: time.Now().Add(10 * time.Minute)},
		},
		lock: new(sync.RWMutex),
		exp:  5 * time.Minute,
	}
	j := &cleaner{interval: interval, stop: make(chan struct{})}
	go j.start(c)
	j.stop <- struct{}{}               // Send stop signal immediately
	time.Sleep(100 * time.Millisecond) // Allow some time for the goroutine to receive the stop signal and exit
	c.lock.RLock()
	require.NotEmpty(t, c.items, "Cache should not be empty as cleaner should have stopped before any expiration")
	c.lock.RUnlock()
}

func TestStopCleanerSendsSignal(t *testing.T) {
	c := &cache{
		cleaner: &cleaner{
			stop: make(chan struct{}, 1),
		},
	}
	stopCleaner(&Cache{c})
	require.Len(t, c.cleaner.stop, 1, "Expected stop channel to receive one signal")
}

func (c *cache) getDefer(key string) (any, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	item, ok := c.items[key]
	if !ok {
		return nil, false
	}
	return item.Value, true
}

func (c *cache) setDefer(key string, value any) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.items[key] = cacheItem{ExpiresAt: time.Now().Add(c.exp), Value: value}
}

// BenchmarkDefer tests the performance of the cache with and without defer in key places.
//
// The original go-cache code carries with it a comment which says:
// "Calls to mu.Unlock are currently not deferred because defer adds ~200 ns (as of go1.)".
// As of Go1.22.3, this does not seem to be as bad as it seems, as it only seems to add ~20ns.
// However, since these functions are rather simple, not using a `defer` statement does not hurt
// the readability or maintainability of the code and as such it is decided to not use `defer` statements here.
func BenchmarkDefer(b *testing.B) {
	deferCache := NewCache(10*time.Second, 1*time.Second)
	noDeferCache := NewCache(10*time.Second, 1*time.Second)
	b.Run("set with defer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			deferCache.setDefer(fmt.Sprintf("%d", i), i)
		}
	})
	b.Run("set without defer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			noDeferCache.setDefer(fmt.Sprintf("%d", i), i)
		}
	})
	b.Run("get with defer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = deferCache.getDefer(fmt.Sprintf("%d", i))
		}
	})
	b.Run("get without defer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = noDeferCache.getDefer(fmt.Sprintf("%d", i))
		}
	})
}

func TestCache_Set_NoUpdateExpiry(t *testing.T) {
	cache := NewCache(10*time.Second, 1*time.Second)
	cache.Set("foo", "bar")
	expiresAt := cache.items["foo"].ExpiresAt
	cache.Set("foo", "bar")
	require.Equal(t, expiresAt, cache.items["foo"].ExpiresAt)
}
