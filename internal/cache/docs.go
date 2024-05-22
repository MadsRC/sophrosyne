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

// Package cache implements a caching layer for Sophrosyne.
//
// The original caching code is loosely based on Patrick Mylund Nielsen's go-cache project
// (https://github.com/patrickmn/go-cache), but with modifications made to remove unnecessary functionality. The
// code that can be considered a derivative work of go-cache can be found in the cache.go file in this package.
//
// Code in this package carries with it optimizations which to some may seem unnatural (that's a Star Wars quote,
// right?). Examples of these optimizations is the lack of `defer` statements to the extent that they do not hurt the
// readability of the code. See the comments on the BenchmarkDefer function for more details.
package cache
