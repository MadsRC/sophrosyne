package services

import (
	"bytes"
	"context"
	"fmt"
	"github.com/madsrc/sophrosyne"
	"github.com/madsrc/sophrosyne/internal/rpc/jsonrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"strings"
	"testing"
)

func TestNewUserService(t *testing.T) {
	type args struct {
		userService sophrosyne.UserService
		authz       sophrosyne.AuthorizationProvider
		logger      *slog.Logger
		validator   sophrosyne.Validator
	}
	tests := []struct {
		name    string
		args    args
		want    *UserService
		wantErr assert.ErrorAssertionFunc
	}{
		{
			"success",
			args{
				nil,
				nil,
				nil,
				nil,
			},
			&UserService{
				nil,
				nil,
				nil,
				nil,
			},
			assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewUserService(tt.args.userService, tt.args.authz, tt.args.logger, tt.args.validator)
			if !tt.wantErr(t, err, fmt.Sprintf("NewUserService(%v, %v, %v, %v)", tt.args.userService, tt.args.authz, tt.args.logger, tt.args.validator)) {
				return
			}
			assert.Equalf(t, tt.want, got, "NewUserService(%v, %v, %v, %v)", tt.args.userService, tt.args.authz, tt.args.logger, tt.args.validator)
		})
	}
}

func TestUserService_CreateUser(t *testing.T) {
	type fields struct {
		userService sophrosyne.UserService
		authz       sophrosyne.AuthorizationProvider
		logger      *slog.Logger
		validator   sophrosyne.Validator
	}
	type args struct {
		ctx context.Context
		req jsonrpc.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := UserService{
				userService: tt.fields.userService,
				authz:       tt.fields.authz,
				logger:      tt.fields.logger,
				validator:   tt.fields.validator,
			}
			got, err := u.CreateUser(tt.args.ctx, tt.args.req)
			if !tt.wantErr(t, err, fmt.Sprintf("CreateUser(%v, %v)", tt.args.ctx, tt.args.req)) {
				return
			}
			assert.Equalf(t, tt.want, got, "CreateUser(%v, %v)", tt.args.ctx, tt.args.req)
		})
	}
}

func TestUserService_DeleteUser(t *testing.T) {
	type fields struct {
		userService sophrosyne.UserService
		authz       sophrosyne.AuthorizationProvider
		logger      *slog.Logger
		validator   sophrosyne.Validator
	}
	type args struct {
		ctx context.Context
		req jsonrpc.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := UserService{
				userService: tt.fields.userService,
				authz:       tt.fields.authz,
				logger:      tt.fields.logger,
				validator:   tt.fields.validator,
			}
			got, err := u.DeleteUser(tt.args.ctx, tt.args.req)
			if !tt.wantErr(t, err, fmt.Sprintf("DeleteUser(%v, %v)", tt.args.ctx, tt.args.req)) {
				return
			}
			assert.Equalf(t, tt.want, got, "DeleteUser(%v, %v)", tt.args.ctx, tt.args.req)
		})
	}
}

func TestUserService_EntityID(t *testing.T) {
	type fields struct {
		userService sophrosyne.UserService
		authz       sophrosyne.AuthorizationProvider
		logger      *slog.Logger
		validator   sophrosyne.Validator
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"success",
			fields{
				nil,
				nil,
				nil,
				nil,
			},
			"Users",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := UserService{
				userService: tt.fields.userService,
				authz:       tt.fields.authz,
				logger:      tt.fields.logger,
				validator:   tt.fields.validator,
			}
			assert.Equalf(t, tt.want, u.EntityID(), "EntityID()")
		})
	}
}

func TestUserService_EntityType(t *testing.T) {
	type fields struct {
		userService sophrosyne.UserService
		authz       sophrosyne.AuthorizationProvider
		logger      *slog.Logger
		validator   sophrosyne.Validator
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"success",
			fields{
				nil,
				nil,
				nil,
				nil,
			},
			"Service",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := UserService{
				userService: tt.fields.userService,
				authz:       tt.fields.authz,
				logger:      tt.fields.logger,
				validator:   tt.fields.validator,
			}
			assert.Equalf(t, tt.want, u.EntityType(), "EntityType()")
		})
	}
}

func TestUserService_GetUser(t *testing.T) {
	type fields struct {
		userService sophrosyne.UserService
		authz       sophrosyne.AuthorizationProvider
		logger      *slog.Logger
		validator   sophrosyne.Validator
	}
	type args struct {
		ctx context.Context
		req jsonrpc.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := UserService{
				userService: tt.fields.userService,
				authz:       tt.fields.authz,
				logger:      tt.fields.logger,
				validator:   tt.fields.validator,
			}
			got, err := u.GetUser(tt.args.ctx, tt.args.req)
			if !tt.wantErr(t, err, fmt.Sprintf("GetUser(%v, %v)", tt.args.ctx, tt.args.req)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetUser(%v, %v)", tt.args.ctx, tt.args.req)
		})
	}
}

func TestUserService_GetUsers(t *testing.T) {
	type fields struct {
		userService sophrosyne.UserService
		authz       sophrosyne.AuthorizationProvider
		logger      *slog.Logger
		validator   sophrosyne.Validator
	}
	type args struct {
		ctx context.Context
		req jsonrpc.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := UserService{
				userService: tt.fields.userService,
				authz:       tt.fields.authz,
				logger:      tt.fields.logger,
				validator:   tt.fields.validator,
			}
			got, err := u.GetUsers(tt.args.ctx, tt.args.req)
			if !tt.wantErr(t, err, fmt.Sprintf("GetUsers(%v, %v)", tt.args.ctx, tt.args.req)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetUsers(%v, %v)", tt.args.ctx, tt.args.req)
		})
	}
}

func logAssertion(t *testing.T, expected, got []string) {
	t.Helper()
	require.Lenf(t, got, len(expected), "logAssertion(%v, %v)", expected, got)
	if got[0] == "" && expected[0] == "" && len(got) == 1 && len(expected) == 1 {
		// We're fine, we expected no logs - Because we splitting lines by \n has the funky quirk of returning an
		// array of length 1 with an empty string if the string being split is empty...
		return
	}
	for i, log := range got {
		require.JSONEqf(t, expected[i], log, "logAssertion(%v, %v)", expected[i], got)
	}
}

func TestUserService_InvokeMethod(t *testing.T) {
	type fields struct {
		userService sophrosyne.UserService
		authz       sophrosyne.AuthorizationProvider
		logger      *slog.Logger
		validator   sophrosyne.Validator
	}
	type args struct {
		ctx context.Context
		req jsonrpc.Request
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		want       []byte
		wantErr    assert.ErrorAssertionFunc
		assertLogs []string
	}{
		{
			name: "method without separator",
			fields: fields{
				nil,
				nil,
				slog.Default(),
				nil,
			},
			args: args{
				context.Background(),
				jsonrpc.Request{
					Method: "test",
				},
			},
			want:    []byte(`{"jsonrpc":"2.0","error":{"code":-32603,"message":"Internal error"},"id":""}`),
			wantErr: assert.NoError,
			assertLogs: []string{
				`{"error":{"stack":"stacktrace"}, "level":"ERROR", "msg":"unreachable"}`,
			},
		},
		{
			name: "Successfull call to Users::GetUser",
			fields: fields{
				nil,
				nil,
				slog.Default(),
				nil,
			},
			args: args{
				context.Background(),
				jsonrpc.Request{
					ID:     "123",
					Method: "Users::GetUser",
				},
			},
			want:    []byte(`{"jsonrpc":"2.0","error":{"code":-32602,"message":"Invalid Params"},"id":"123"}`),
			wantErr: assert.NoError,
			assertLogs: []string{
				"{\"level\":\"ERROR\",\"msg\":\"error extracting params from request\",\"error\":\"no params found\"}",
			},
		},
		{
			name: "Successfull call to Users::GetUsers",
			fields: fields{
				nil,
				nil,
				slog.Default(),
				nil,
			},
			args: args{
				context.Background(),
				jsonrpc.Request{
					ID:     "sadlk;fghj",
					Method: "Users::GetUsers",
				},
			},
			want:       []byte(`{"jsonrpc":"2.0","error":{"code":-32603,"message":"Internal error"},"id":"sadlk;fghj"}`),
			wantErr:    assert.NoError,
			assertLogs: []string{""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := UserService{
				userService: tt.fields.userService,
				authz:       tt.fields.authz,
				logger:      tt.fields.logger,
				validator:   tt.fields.validator,
			}
			var logBuf bytes.Buffer
			u.logger = slog.New(slog.NewJSONHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug, ReplaceAttr: func(group []string, a slog.Attr) slog.Attr {
				if a.Key == "time" {
					return slog.Attr{}
				}
				if a.Key == "stack" {
					return slog.Attr{Key: "stack", Value: slog.StringValue("stacktrace")}
				}
				return slog.Attr{Key: a.Key, Value: a.Value}
			}}))
			got, err := u.InvokeMethod(tt.args.ctx, tt.args.req)
			logAssertion(t, tt.assertLogs, strings.Split(strings.TrimRight(logBuf.String(), "\n"), "\n"))
			if !tt.wantErr(t, err, fmt.Sprintf("InvokeMethod(%v, %v)", tt.args.ctx, tt.args.req)) {
				return
			}
			assert.Equalf(t, tt.want, got, "InvokeMethod(%v, %v)", tt.args.ctx, tt.args.req)
		})
	}
}

func TestUserService_RotateToken(t *testing.T) {
	type fields struct {
		userService sophrosyne.UserService
		authz       sophrosyne.AuthorizationProvider
		logger      *slog.Logger
		validator   sophrosyne.Validator
	}
	type args struct {
		ctx context.Context
		req jsonrpc.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := UserService{
				userService: tt.fields.userService,
				authz:       tt.fields.authz,
				logger:      tt.fields.logger,
				validator:   tt.fields.validator,
			}
			got, err := u.RotateToken(tt.args.ctx, tt.args.req)
			if !tt.wantErr(t, err, fmt.Sprintf("RotateToken(%v, %v)", tt.args.ctx, tt.args.req)) {
				return
			}
			assert.Equalf(t, tt.want, got, "RotateToken(%v, %v)", tt.args.ctx, tt.args.req)
		})
	}
}

func TestUserService_UpdateUser(t *testing.T) {
	type fields struct {
		userService sophrosyne.UserService
		authz       sophrosyne.AuthorizationProvider
		logger      *slog.Logger
		validator   sophrosyne.Validator
	}
	type args struct {
		ctx context.Context
		req jsonrpc.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := UserService{
				userService: tt.fields.userService,
				authz:       tt.fields.authz,
				logger:      tt.fields.logger,
				validator:   tt.fields.validator,
			}
			got, err := u.UpdateUser(tt.args.ctx, tt.args.req)
			if !tt.wantErr(t, err, fmt.Sprintf("UpdateUser(%v, %v)", tt.args.ctx, tt.args.req)) {
				return
			}
			assert.Equalf(t, tt.want, got, "UpdateUser(%v, %v)", tt.args.ctx, tt.args.req)
		})
	}
}
