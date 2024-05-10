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
package middleware

import (
	"context"
	"encoding/base64"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/madsrc/sophrosyne"

	ownHttp "github.com/madsrc/sophrosyne/internal/http"
)

// Middleware to catch panics.
//
// When a panic is encountered, a response is returned to the client using
// [sophrosyne.RespondWithHTTPError] with a [sophrosyne.PanicError].
//
// This middleware should be the first middleware in the chain.
//
// This middleware does not attempt to log the panic, but relies on the fact
// that the creation of a [sophrosyne.PanicError] will capture the necessary
// information, and the [sophrosyne.RespondWithHTTPError] function will ensure the
// error is handled appropriately.
func PanicCatcher(logger *slog.Logger, metricService sophrosyne.MetricService, next http.Handler) http.Handler {
	logger.Debug("Creating PanicCatcher middleware")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.DebugContext(r.Context(), "Entering PanicCatcher middleware")
		defer func() {
			logger.DebugContext(r.Context(), "Executing deferred function in PanicCatcher middleware")
			if err := recover(); err != nil {
				metricService.RecordPanic(r.Context())
				logger.ErrorContext(r.Context(), "Panic encountered", "error", err)
				ownHttp.WriteInternalServerError(r.Context(), w, logger)
			}
		}()
		next.ServeHTTP(w, r)
		logger.DebugContext(r.Context(), "Exiting PanicCatcher middleware")
	})

}

func SetupTracing(tracingService sophrosyne.TracingService, next http.Handler) http.Handler {
	return tracingService.NewHTTPHandler("incoming HTTP request", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	}))
}

func Authentication(exceptions []string, config *sophrosyne.Config, userService sophrosyne.UserService, logger *slog.Logger, next http.Handler) http.Handler {
	logger.Debug("Creating Authentication middleware")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.DebugContext(r.Context(), "Entering Authentication middleware")
		defer logger.DebugContext(r.Context(), "Exiting Authentication middleware")

		// Check if the request path is in the exceptions list
		for _, path := range exceptions {
			if strings.HasPrefix(r.URL.Path, path) {
				logger.DebugContext(r.Context(), "request path is in authentication exceptions list", "matched_exception_entry", path, "request_path", r.URL.Path)
				next.ServeHTTP(w, r)
				return
			}
		}

		// Extract authentication header
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			logger.DebugContext(r.Context(), "unable to extract token from Authorization header", "header", authHeader)
			logger.InfoContext(r.Context(), "authentication", "result", "failed")
			ownHttp.WriteResponse(r.Context(), w, http.StatusUnauthorized, "text/plain", nil, logger)
			return
		}

		// Extract token
		token, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(authHeader, "Bearer "))
		if err != nil {
			logger.DebugContext(r.Context(), "unable to decode token", "token", token, "error", err)
			logger.InfoContext(r.Context(), "authentication", "result", "failed")
			ownHttp.WriteResponse(r.Context(), w, http.StatusUnauthorized, "text/plain", nil, logger)
			return
		}

		// Hash the token using ProtectToken
		hashedToken := sophrosyne.ProtectToken(token, config)

		// Validate token
		user, err := userService.GetUserByToken(r.Context(), hashedToken)
		if err != nil {
			logger.DebugContext(r.Context(), "unable to validate token", "error", err)
			logger.InfoContext(r.Context(), "authentication", "result", "failed")
			ownHttp.WriteResponse(r.Context(), w, http.StatusUnauthorized, "text/plain", nil, logger)
			return
		}
		user.Token = []byte{} // Overwrite the token, so we don't leak it into the context
		ctx := r.Context()
		ctx = context.WithValue(ctx, sophrosyne.UserContextKey{}, &user)
		r = r.WithContext(ctx)
		logger.InfoContext(r.Context(), "authenticated", "result", "success")

		next.ServeHTTP(w, r)
	})
}

type responseWrapper struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWrapper {
	return &responseWrapper{ResponseWriter: w}
}

func (w *responseWrapper) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
	w.wroteHeader = true

	return
}

func (w *responseWrapper) Status() int {
	return w.status
}

func RequestLogging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		begin := time.Now()
		defer func() {
			logger.InfoContext(r.Context(), "request served", "remote", r.RemoteAddr, "method", r.Method, "path", r.URL.Path, "user_agent", r.UserAgent(), "duration_ms", time.Since(begin)+time.Millisecond)
		}()
		wrapped := wrapResponseWriter(w)
		next.ServeHTTP(wrapped, r)
	})
}
