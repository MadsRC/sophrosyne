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

package validator

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
)

func TestNewValidator(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "NewValidator",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewValidator()
			require.NotNil(t, got)
			require.NotNil(t, got.v)
		})
	}
}

type stupid struct{}

func (s stupid) Error() string {
	return "stupid"
}

func TestValidator_Validate(t *testing.T) {

	type fields struct {
		v *validator.Validate
	}
	type args struct {
		i interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		panics  bool
	}{
		{
			name: "Validate_success",
			fields: fields{
				v: validator.New(validator.WithRequiredStructEnabled()),
			},
			args: args{
				i: struct {
					ID   string `validate:"required"`
					Name string `validate:"required"`
				}{
					ID:   "1",
					Name: "name",
				},
			},
		},
		{
			name: "Validate_bad_tag_cause_panic",
			fields: fields{
				v: validator.New(validator.WithRequiredStructEnabled()),
			},
			args: args{
				i: struct {
					ID string `validate:"somethingVeryCustom"`
				}{
					ID: "1",
				},
			},
			wantErr: true,
			panics:  true,
		},
		{
			name: "Validate_error",
			fields: fields{
				v: validator.New(validator.WithRequiredStructEnabled()),
			},
			args: args{
				i: struct {
					ID   string `validate:"required"`
					Name string `validate:"required"`
				}{
					ID: "1",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Validator{
				v: tt.fields.v,
			}
			if tt.panics {
				require.Panics(t, func() {
					_ = v.Validate(tt.args.i)
				})
				return
			}
			err := v.Validate(tt.args.i)
			if tt.wantErr {
				require.Error(t, err)
				require.ErrorAs(t, err, &validator.ValidationErrors{})
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMutualExclusivity_Two_Fields(t *testing.T) {
	type obj struct {
		A string
		B string `validate:"required_without=A,excluded_with=A"`
	}
	type args struct {
		i interface{}
	}
	tests := []struct {
		name      string
		args      args
		wantErr   bool
		failedTag string
	}{
		{
			name: "only A set",
			args: args{
				i: obj{
					A: "a",
				},
			},
		},
		{
			name: "A and B set",
			args: args{
				i: obj{
					A: "a",
					B: "b",
				},
			},
			wantErr:   true,
			failedTag: "excluded_with",
		},
		{
			name: "only B set",
			args: args{
				i: obj{
					B: "b",
				},
			},
		},
		{
			name: "none is set",
			args: args{
				i: obj{},
			},
			wantErr:   true,
			failedTag: "required_without",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			err := v.Validate(tt.args.i)
			if tt.wantErr {
				require.Error(t, err)
				var ve validator.ValidationErrors
				require.ErrorAs(t, err, &ve)
				require.Len(t, ve, 1)
				require.Equal(t, "B", ve[0].Field())
				require.Equal(t, tt.failedTag, ve[0].Tag())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
