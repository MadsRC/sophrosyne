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

package interceptors

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"log/slog"
)

func Logger(log *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		largs := append([]any{}, fields...)
		switch lvl {
		case logging.LevelDebug:
			log.DebugContext(ctx, msg, largs...)
		case logging.LevelInfo:
			log.InfoContext(ctx, msg, largs...)
		case logging.LevelWarn:
			log.WarnContext(ctx, msg, largs...)
		case logging.LevelError:
			log.ErrorContext(ctx, msg, largs...)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}
