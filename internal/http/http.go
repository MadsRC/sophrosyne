package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/madsrc/sophrosyne"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"
)

type Server struct {
	appConfig      *sophrosyne.Config   `validate:"required"`
	mux            *http.ServeMux       `validate:"required"`
	validator      sophrosyne.Validator `validate:"required"`
	middleware     []func(http.Handler) http.Handler
	logger         *slog.Logger              `validate:"required"`
	http           *http.Server              `validate:"required"`
	tracingService sophrosyne.TracingService `validate:"required"`
	userService    sophrosyne.UserService    `validate:"required"`
}

func NewServer(ctx context.Context, appConfig *sophrosyne.Config, validator sophrosyne.Validator, logger *slog.Logger, tracingService sophrosyne.TracingService, userService sophrosyne.UserService, tlsConfig *tls.Config) (*Server, error) {
	mux := http.NewServeMux()
	s := Server{appConfig: appConfig,
		validator: validator,
		logger:    logger,
		http: &http.Server{
			Addr:         fmt.Sprintf(":%d", appConfig.Server.Port),
			Handler:      mux,
			BaseContext:  func(_ net.Listener) context.Context { return ctx },
			ReadTimeout:  time.Second,
			WriteTimeout: 10 * time.Second,
			TLSConfig:    tlsConfig,
		},
		mux:            mux,
		tracingService: tracingService,
		userService:    userService,
	}

	if err := s.validator.Validate(s); err != nil {
		return nil, err
	}

	return &s, nil
}

func (s *Server) Start() error {
	s.logger.Info("Starting server", "port", s.appConfig.Server.Port)
	return s.http.ListenAndServeTLS("", "")
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.InfoContext(ctx, "Shutting down server")
	return s.http.Shutdown(ctx)
}

func (s *Server) Handle(path string, handler http.Handler) {
	s.mux.Handle(path, handler)
}

func RPCHandler(logger *slog.Logger, rpcService sophrosyne.RPCServer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		body, err := io.ReadAll(r.Body) // Find a way to implement a limit on the body size
		b, err := rpcService.HandleRPCRequest(r.Context(), body)
		if err != nil {
			logger.ErrorContext(r.Context(), "error handling rpc request", "error", err)
			WriteInternalServerError(r.Context(), w, logger)
			return
		}
		WriteResponse(r.Context(), w, http.StatusOK, "application/json", b, logger)
	})
}

func WriteResponse(ctx context.Context, w http.ResponseWriter, status int, contentType string, data []byte, logger *slog.Logger) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	_, err := w.Write(data)
	if err != nil {
		logger.ErrorContext(ctx, "unable to write response", "error", err)
	}
}

func WriteInternalServerError(ctx context.Context, w http.ResponseWriter, logger *slog.Logger) {
	logger.ErrorContext(ctx, "returning internal server error")
	WriteResponse(ctx, w, http.StatusInternalServerError, "text/plain", []byte("Internal Server Error"), logger)
}
