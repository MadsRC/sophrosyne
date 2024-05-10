package rpc

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/madsrc/sophrosyne"
	"github.com/madsrc/sophrosyne/internal/rpc/internal/jsonrpc"
	"github.com/madsrc/sophrosyne/internal/validator"
)

func TestParamsIntoAny(t *testing.T) {
	type testTarget struct {
		ID        string `json:"id" validate:"required_without=Name Something"`
		Name      string `json:"name" validate:"required_without=ID Something"`
		Something string `json:"something" validate:"required_without=ID Name"`
	}
	type args struct {
		req      *jsonrpc.Request
		target   any
		validate sophrosyne.Validator
	}
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{
		{
			name: "ParamsIntoAny_success",
			args: args{
				req: &jsonrpc.Request{
					Params: &jsonrpc.ParamsObject{
						"id": "1",
					},
				},
				target:   &testTarget{},
				validate: validator.NewValidator(),
			},
			want: &testTarget{
				ID: "1",
			},
		},
		{
			name: "ParamsIntoAny_validate_error",
			args: args{
				req: &jsonrpc.Request{
					Params: &jsonrpc.ParamsObject{
						"id":   "1",
						"Name": "name",
					},
				},
				target:   &testTarget{},
				validate: validator.NewValidator(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ParamsIntoAny(tt.args.req, tt.args.target, tt.args.validate)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.Equal(t, tt.want, tt.args.target)
			}

		})
	}
}

func TestSomething(t *testing.T) {
	b := []byte(`{"jsonrpc":"2.0","method":"Users::GetUser","id":"1234","params":{"id":"coo1tog2e0g00gf27t70"}}`)
	req := &jsonrpc.Request{}
	err := req.UnmarshalJSON(b)
	require.NoError(t, err)

	require.NotNil(t, req)
	require.NotNil(t, req.Params)
}
