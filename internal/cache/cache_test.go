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
	require.Nil(t, tc)
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
