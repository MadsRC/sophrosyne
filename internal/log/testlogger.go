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

package log

import (
	"bytes"
	"log/slog"
)

func NewTestLogger(opts *slog.HandlerOptions) (*slog.Logger, *bytes.Buffer) {
	buf := new(bytes.Buffer)

	if opts == nil {
		opts = &slog.HandlerOptions{Level: slog.LevelDebug}
	}
	return slog.New(slog.NewJSONHandler(buf, opts)), buf
}
