// Sophrosyne
//
//	Copyright (C) 2024  Mads R. Havmand
//
// This program is free software: you can redistribute it and/or modify
//
//	it under the terms of the GNU Affero General Public License as published by
//	the Free Software Foundation, either version 3 of the License, or
//	(at your option) any later version.
//
//	This program is distributed in the hope that it will be useful,
//	but WITHOUT ANY WARRANTY; without even the implied warranty of
//	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//	GNU Affero General Public License for more details.
//
//	You should have received a copy of the GNU Affero General Public License
//	along with this program.  If not, see <http://www.gnu.org/licenses/>.
package sophrosyne

import (
	"context"
	"log/slog"
	"os"
)

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
)

type LogFormat string

const (
	LogFormatText LogFormat = "text"
	LogFormatJSON LogFormat = "json"
)

func LogLevelToSlogLevel(level LogLevel) slog.Level {
	switch level {
	case LogLevelDebug:
		return slog.LevelDebug
	case LogLevelInfo:
		return slog.LevelInfo
	default:
		return slog.LevelInfo
	}
}

type LogHandler struct {
	Handler        slog.Handler   `validate:"required"`
	config         *Config        `validate:"required"`
	tracingService TracingService `validate:"required"`
}

func NewLogHandler(config *Config, tracingService TracingService) *LogHandler {
	h := LogHandler{
		config:         config,
		tracingService: tracingService,
	}
	handlerOpts := slog.HandlerOptions{
		Level: LogLevelToSlogLevel(config.Logging.Level),
	}

	if config.Logging.Format == LogFormatJSON {
		h.Handler = slog.NewJSONHandler(os.Stdout, &handlerOpts)
	} else {
		h.Handler = slog.NewTextHandler(os.Stdout, &handlerOpts)
	}

	return &h
}

// Enabled returns true if the log level is enabled for the handler and false
// otherwise.
//
// The log level is enabled if the level of the record is greater than or equal
// to the level defined in [config.Log.Level].
//
// This is called early in the logging process to determine if the handler
// should be called. Because the handler has access to the configuration, it
// allows us to not have to restart the application to change the log level,
// provided that the part of the configuraiton we change allows for hot
// reloading.
func (h LogHandler) Enabled(ctx context.Context, Level slog.Level) bool {
	return Level >= LogLevelToSlogLevel(h.config.Logging.Level)
}

// Handle adds contextual attributes to the Record before calling the underlying
// handler.
func (h LogHandler) Handle(ctx context.Context, r slog.Record) error {
	if h.tracingService.GetTraceID(ctx) != "" {
		r.AddAttrs(slog.String("trace_id", h.tracingService.GetTraceID(ctx)))
	}
	if ExtractUser(ctx) != nil {
		r.AddAttrs(slog.String("user_id", ExtractUser(ctx).ID))
	}

	return h.Handler.Handle(ctx, r)
}
func (h LogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h.Handler.WithAttrs(attrs)
}
func (h LogHandler) WithGroup(name string) slog.Handler {
	return h.Handler.WithGroup(name)
}
