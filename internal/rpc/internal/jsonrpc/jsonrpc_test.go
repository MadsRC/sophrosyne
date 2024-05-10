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
package jsonrpc

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestID_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		id      ID
		want    []byte
		wantErr bool
	}{
		{
			name: "id is string",
			id:   ID("test"),
			want: []byte(`"test"`),
		},
		{
			name: "id is int",
			id:   ID("1"),
			want: []byte(`"1"`),
		},
		{
			name: "id is float",
			id:   ID("1.1"),
			want: []byte(`"1.1"`),
		},
		{
			name: "id is empty",
			id:   ID(""),
			want: []byte(`""`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalJSON() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestID_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		id      ID
		args    args
		wantErr bool
	}{
		{
			name: "id is string",
			id:   ID("test"),
			args: args{data: []byte(`"test"`)},
		},
		{
			name: "id is int",
			id:   ID("1"),
			args: args{data: []byte(`1`)},
		},
		{
			name: "id is float",
			id:   ID("1.1"),
			args: args{data: []byte(`"1.1"`)},
		},
		{
			name: "id is null",
			id:   ID(""),
			args: args{data: []byte(`null`)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.id.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParamsArray_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		pa      ParamsArray
		args    args
		wantErr bool
	}{
		{
			name: "Params is array",
			pa:   ParamsArray{},
			args: args{data: []byte(`[1,2,3]`)},
		},
		{
			name:    "Params is object",
			pa:      ParamsArray{},
			args:    args{data: []byte(`{"test":1}`)},
			wantErr: true,
		},
		{
			name:    "Params is string",
			pa:      ParamsArray{},
			args:    args{data: []byte(`"test"`)},
			wantErr: true,
		},
		{
			name: "Params are floats",
			pa:   ParamsArray{},
			args: args{data: []byte(`[1.1,2.2,3.3]`)},
		},
		{
			name: "Params are mixed",
			pa:   ParamsArray{},
			args: args{data: []byte(`[1,"test",3.3]`)},
		},
		{
			name: "Params are empty",
			pa:   ParamsArray{},
			args: args{data: []byte(`[]`)},
		},
		{
			name: "Params are null",
			pa:   ParamsArray{},
			args: args{data: []byte(`null`)},
		},
		{
			name:    "empty bytes",
			pa:      ParamsArray{},
			args:    args{data: []byte(``)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.pa.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParamsArray_isParams(t *testing.T) {
	tests := []struct {
		name string
		pa   ParamsArray
	}{
		{
			name: "Params is array",
			pa:   ParamsArray{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.pa.isParams()
		})
	}
}

func TestParamsObject_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		pa      ParamsObject
		args    args
		wantErr bool
	}{
		{
			name: "Params is object",
			pa:   ParamsObject{},
			args: args{data: []byte(`{"test":1}`)},
		},
		{
			name:    "Params is array",
			pa:      ParamsObject{},
			args:    args{data: []byte(`[1,2,3]`)},
			wantErr: true,
		},
		{
			name:    "Params is string",
			pa:      ParamsObject{},
			args:    args{data: []byte(`"test"`)},
			wantErr: true,
		},
		{
			name: "Params are floats",
			pa:   ParamsObject{},
			args: args{data: []byte(`{"test":1.1}`)},
		},
		{
			name: "Params are mixed",
			pa:   ParamsObject{},
			args: args{data: []byte(`{"test":1,"test2":"test","test3":3.3}`)},
		},
		{
			name: "Params are empty",
			pa:   ParamsObject{},
			args: args{data: []byte(`{}`)},
		},
		{
			name: "Params are null",
			pa:   ParamsObject{},
			args: args{data: []byte(`null`)},
		},
		{
			name:    "empty bytes",
			pa:      ParamsObject{},
			args:    args{data: []byte(``)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.pa.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParamsObject_isParams(t *testing.T) {
	tests := []struct {
		name string
		pa   ParamsObject
	}{
		{
			name: "Params is object",
			pa:   ParamsObject{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.pa.isParams()
		})
	}
}

func TestRequest_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name           string
		r              Request
		args           args
		wantErr        bool
		isNotification bool
	}{
		{
			name: "request is object",
			r:    Request{},
			args: args{data: []byte(`{"jsonrpc":"2.0","method":"test","params":[1,2,3],"id":1}`)},
		},
		{
			name:    "request is array",
			r:       Request{},
			args:    args{data: []byte(`[1,2,3]`)},
			wantErr: true,
		},
		{
			name:    "request is string",
			r:       Request{},
			args:    args{data: []byte(`"test"`)},
			wantErr: true,
		},
		{
			name:    "request is null",
			r:       Request{},
			args:    args{data: []byte(`null`)},
			wantErr: true,
		},
		{
			name:    "empty bytes",
			r:       Request{},
			args:    args{data: []byte(``)},
			wantErr: true,
		},
		{
			name:    "request is missing jsonrpc",
			r:       Request{},
			args:    args{data: []byte(`{"method":"test","params":[1,2,3],"id":1}`)},
			wantErr: true,
		},
		{
			name:    "request has bad jsonrpc version",
			r:       Request{},
			args:    args{data: []byte(`{"jsonrpc":"1.0","method":"test","params":[1,2,3],"id":1}`)},
			wantErr: true,
		},
		{
			name:    "request is missing method",
			r:       Request{},
			args:    args{data: []byte(`{"jsonrpc":"2.0","params":[1,2,3],"id":1}`)},
			wantErr: true,
		},
		{
			name:    "request has empty method",
			r:       Request{},
			args:    args{data: []byte(`{"jsonrpc":"2.0","method":"","params":[1,2,3],"id":1}`)},
			wantErr: true,
		},
		{
			name:    "request version is a number",
			r:       Request{},
			args:    args{data: []byte(`{"jsonrpc":2.0,"method":"test","params":[1,2,3],"id":1}`)},
			wantErr: true,
		},
		{
			name: "request ID is zero",
			r:    Request{},
			args: args{data: []byte(`{"jsonrpc":"2.0","method":"test","params":[1,2,3],"id":0}`)},
		},
		{
			name: "request ID is zero string",
			r:    Request{},
			args: args{data: []byte(`{"jsonrpc":"2.0","method":"test","params":{"one":"two"},"id":"0"}`)},
		},
		{
			name:    "request ID is an empty object",
			r:       Request{},
			args:    args{data: []byte(`{"jsonrpc":"2.0","method":"test","params":[1,2,3],"id":{}}`)},
			wantErr: true,
		},
		{
			name:    "request ID is an object",
			r:       Request{},
			args:    args{data: []byte(`{"jsonrpc":"2.0","method":"test","params":[1,2,3],"id":{"test":1}}`)},
			wantErr: true,
		},
		{
			name:    "request ID is an array",
			r:       Request{},
			args:    args{data: []byte(`{"jsonrpc":"2.0","method":"test","params":[1,2,3],"id":[1,2,3]}`)},
			wantErr: true,
		},
		{
			name:    "request ID is an empty array",
			r:       Request{},
			args:    args{data: []byte(`{"jsonrpc":"2.0","method":"test","params":[1,2,3],"id":[]}`)},
			wantErr: true,
		},
		{
			name: "request ID is an empty string",
			r:    Request{},
			args: args{data: []byte(`{"jsonrpc":"2.0","method":"test","params":[1,2,3],"id":""}`)},
		},
		{
			name:    "method is rpc.",
			r:       Request{},
			args:    args{data: []byte(`{"jsonrpc":"2.0","method":"rpc.test","params":[1,2,3],"id":1}`)},
			wantErr: true,
		},
		{
			name:    "method is empty",
			r:       Request{},
			args:    args{data: []byte(`{"jsonrpc":"2.0","method":"","params":[1,2,3],"id":1}`)},
			wantErr: true,
		},
		{
			name: "request has null Params",
			r:    Request{},
			args: args{data: []byte(`{"jsonrpc":"2.0","method":"test","params":null,"id":1}`)},
		},
		{
			name:           "request is notification",
			r:              Request{},
			args:           args{data: []byte(`{"jsonrpc":"2.0","method":"test","params":[1,2,3]}`)},
			isNotification: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.r.UnmarshalJSON(tt.args.data)
			if tt.wantErr {
				require.Error(t, err, "unmarshalled object: %+v", tt.r)
			} else {
				require.NoError(t, err, "unmarshalled object: %+v", tt.r)
				if tt.isNotification {
					require.True(t, tt.r.IsNotification(), "unmarshalled object: %+v", tt.r)
				}
			}
		})
	}
}

func TestRequest_isNotification(t *testing.T) {
	tests := []struct {
		name string
		r    Request
		want bool
	}{
		{
			name: "request is notification",
			r: Request{
				isNotification: true,
			},
			want: true,
		},
		{
			name: "request is not notification",
			r: Request{
				isNotification: false,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.IsNotification(); got != tt.want {
				t.Errorf("IsNotification() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequest_AsNotification(t *testing.T) {
	tests := []struct {
		name string
		r    Request
		want Request
	}{
		{
			name: "request is notification",
			r: Request{
				isNotification: true,
			},
			want: Request{
				isNotification: true,
			},
		},
		{
			name: "request is not notification",
			r: Request{
				isNotification: false,
			},
			want: Request{
				isNotification: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.AsNotification()
			require.Equal(t, tt.want, tt.r)
		})
	}
}

func TestRequest_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		r       Request
		want    []byte
		wantErr bool
	}{
		{
			name: "request is object",
			r: Request{
				Method: "test",
				Params: &ParamsArray{1, 2, 3},
				ID:     ID("1"),
			},
			want: []byte(`{"jsonrpc":"2.0","method":"test","params":[1,2,3],"id":"1"}`),
		},
		{
			name: "request is object with null params",
			r: Request{
				Method: "test",
				Params: nil,
				ID:     ID("1"),
			},
			want: []byte(`{"jsonrpc":"2.0","method":"test","id":"1"}`),
		},
		{
			name: "request is object with empty params",
			r: Request{
				Method: "test",
				Params: &ParamsObject{},
				ID:     ID("1"),
			},
			want: []byte(`{"jsonrpc":"2.0","method":"test","params":{},"id":"1"}`),
		},
		{
			name: "request is object with empty id",
			r: Request{
				Method: "test",
				Params: &ParamsArray{1, 2, 3},
				ID:     ID(""),
			},
			want: []byte(`{"jsonrpc":"2.0","method":"test","params":[1,2,3],"id":""}`),
		},
		{
			name: "request is object with zero id",
			r: Request{
				Method: "test",
				Params: &ParamsArray{1, 2, 3},
				ID:     ID("0"),
			},
			want: []byte(`{"jsonrpc":"2.0","method":"test","params":[1,2,3],"id":"0"}`),
		},
		{
			name: "request is a notification",
			r: Request{
				Method:         "test",
				Params:         &ParamsArray{1, 2, 3},
				isNotification: true,
			},
			want: []byte(`{"jsonrpc":"2.0","method":"test","params":[1,2,3]}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.MarshalJSON()
			if tt.wantErr {
				require.Error(t, err, "marshalled object: %+v", tt.r)
			} else {
				require.NoError(t, err, "marshalled object: %+v", tt.r)
				require.JSONEq(t, string(tt.want), string(got), "marshalled object: %+v", tt.r)
			}
		})
	}
}

func TestValidateErrorCode(t *testing.T) {
	type args struct {
		field reflect.Value
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "ParseError passes validation",
			args: args{field: reflect.ValueOf(ParseError)},
			want: nil,
		},
		{
			name: "InvalidRequest passes validation",
			args: args{field: reflect.ValueOf(InvalidRequest)},
			want: nil,
		},
		{
			name: "MethodNotFound passes validation",
			args: args{field: reflect.ValueOf(MethodNotFound)},
			want: nil,
		},
		{
			name: "InvalidParams passes validation",
			args: args{field: reflect.ValueOf(InvalidParams)},
			want: nil,
		},
		{
			name: "InternalError passes validation",
			args: args{field: reflect.ValueOf(InternalError)},
			want: nil,
		},
		{
			name: "-32769 passes validation",
			args: args{field: reflect.ValueOf(-32769)},
			want: nil,
		},
		{
			name: "-32768 fails validation",
			args: args{field: reflect.ValueOf(-32768)},
			want: "invalid error code - code -32768 is reserved for future use",
		},
		{
			name: "-32767 fails validation",
			args: args{field: reflect.ValueOf(-32767)},
			want: "invalid error code - code -32767 is reserved for future use",
		},

		{
			name: "-32000 fails validation",
			args: args{field: reflect.ValueOf(-32000)},
			want: "invalid error code - code -32000 is reserved for future use",
		},
		{
			name: "-31999 passes validation",
			args: args{field: reflect.ValueOf(-31999)},
			want: nil,
		},
		{
			name: "string fails validation",
			args: args{field: reflect.ValueOf("test")},
			want: "error code must be an integer",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateErrorCode(tt.args.field); !reflect.DeepEqual(got, tt.want) {
				if v, ok := got.(error); ok {
					if v.Error() == tt.want {
						return
					}
				}
				t.Errorf("ValidateErrorCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResponse_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		r       Response
		args    []byte
		wantErr bool
	}{
		{
			name: "response is object",
			r:    Response{},
			args: []byte(`{"jsonrpc":"2.0","result":1,"id":1}`),
		},
		{
			name:    "response is array",
			r:       Response{},
			args:    []byte(`[1,2,3]`),
			wantErr: true,
		},
		{
			name:    "response is string",
			r:       Response{},
			args:    []byte(`"test"`),
			wantErr: true,
		},
		{
			name:    "response is null",
			r:       Response{},
			args:    []byte(`null`),
			wantErr: true,
		},
		{
			name:    "empty bytes",
			r:       Response{},
			args:    []byte(``),
			wantErr: true,
		},
		{
			name:    "response is missing jsonrpc",
			r:       Response{},
			args:    []byte(`{"result":1,"id":1}`),
			wantErr: true,
		},
		{
			name:    "response has bad jsonrpc version",
			r:       Response{},
			args:    []byte(`{"jsonrpc":"1.0","result":1,"id":1}`),
			wantErr: true,
		},
		{
			name:    "response jsonrpc is bad array",
			r:       Response{},
			args:    []byte(`{"jsonrpc":[1,2,3],"result":1,"id":1}`),
			wantErr: true,
		},
		{
			name:    "response is missing result and error",
			r:       Response{},
			args:    []byte(`{"jsonrpc":"2.0","id":1}`),
			wantErr: true,
		},
		{
			name:    "response has nil result and error",
			r:       Response{},
			args:    []byte(`{"jsonrpc":"2.0","result":null,"error":null,"id":1}`),
			wantErr: true,
		},
		{
			name:    "response has both result and error",
			r:       Response{},
			args:    []byte(`{"jsonrpc":"2.0","result":1,"error":1,"id":1}`),
			wantErr: true,
		},
		{
			name:    "response has empty result",
			r:       Response{},
			args:    []byte(`{"jsonrpc":"2.0","result":null,"id":1}`),
			wantErr: true,
		},
		{
			name:    "response has empty error",
			r:       Response{},
			args:    []byte(`{"jsonrpc":"2.0","error":null,"id":1}`),
			wantErr: true,
		},
		{
			name: "response has empty id - null used for unknown id",
			r:    Response{},
			args: []byte(`{"jsonrpc":"2.0","result":1,"id":null}`),
		},
		{
			name: "id is zero",
			r:    Response{},
			args: []byte(`{"jsonrpc":"2.0","result":1,"id":0}`),
		},
		{
			name: "id is zero string",
			r:    Response{},
			args: []byte(`{"jsonrpc":"2.0","result":1,"id":"0"}`),
		},
		{
			name:    "id is an empty object",
			r:       Response{},
			args:    []byte(`{"jsonrpc":"2.0","result":1,"id":{}}`),
			wantErr: true,
		},
		{
			name:    "id is an object",
			r:       Response{},
			args:    []byte(`{"jsonrpc":"2.0","result":1,"id":{"test":1}}`),
			wantErr: true,
		},
		{
			name:    "id is an array",
			r:       Response{},
			args:    []byte(`{"jsonrpc":"2.0","result":1,"id":[1,2,3]}`),
			wantErr: true,
		},
		{
			name:    "id is an empty array",
			r:       Response{},
			args:    []byte(`{"jsonrpc":"2.0","result":1,"id":[]}`),
			wantErr: true,
		},
		{
			name: "id is an empty string",
			r:    Response{},
			args: []byte(`{"jsonrpc":"2.0","result":1,"id":""}`),
		},
		{
			name: "id is null",
			r:    Response{},
			args: []byte(`{"jsonrpc":"2.0","result":1,"id":null}`),
		},
		{
			name:    "id is missing",
			r:       Response{},
			args:    []byte(`{"jsonrpc":"2.0","result":1}`),
			wantErr: true,
		},
		{
			name: "response has error",
			r:    Response{},
			args: []byte(`{"jsonrpc":"2.0","error":{"code":1,"message":"test"},"id":1}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.r.UnmarshalJSON(tt.args)
			if tt.wantErr {
				require.Error(t, err, "unmarshalled object: %+v", tt.r)
			} else {
				require.NoError(t, err, "unmarshalled object: %+v", tt.r)
			}
		})

	}
}

func TestValidateMethod(t *testing.T) {
	type args struct {
		field reflect.Value
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "integer fails validation",
			args: args{field: reflect.ValueOf(1)},
			want: "method must be a string",
		},
		{
			name: "empty string fails validation",
			args: args{field: reflect.ValueOf("")},
			want: "method cannot be empty",
		},
		{
			name: "reserved method fails validation",
			args: args{field: reflect.ValueOf("rpc.")},
			want: "methods beginning with rpc. are reserved",
		},
		{
			name: "other reserved method fails validation",
			args: args{field: reflect.ValueOf("rpc.someMethod")},
			want: "methods beginning with rpc. are reserved",
		},
		{
			name: "valid method passes validation",
			args: args{field: reflect.ValueOf("someMethod")},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateMethod(tt.args.field); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateMethod() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_optional_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	type testCase[T any] struct {
		name    string
		o       optional[T]
		args    args
		wantErr bool
	}
	tests := []testCase[string]{
		{
			name: "value is string",
			o:    optional[string]{},
			args: args{data: []byte(`"test"`)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.o.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_BatchRequest_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		br      BatchRequest
		args    []byte
		wantErr bool
	}{
		{
			name: "batch request is array",
			br:   BatchRequest{},
			args: []byte(`[{"jsonrpc":"2.0","method":"test","params":[1,2,3],"id":1}]`),
		},
		{
			name:    "batch request is object",
			br:      BatchRequest{},
			args:    []byte(`{"jsonrpc":"2.0","method":"test","params":[1,2,3],"id":1}`),
			wantErr: true,
		},
		{
			name:    "batch request is string",
			br:      BatchRequest{},
			args:    []byte(`"test"`),
			wantErr: true,
		},
		{
			name:    "batch request is null",
			br:      BatchRequest{},
			args:    []byte(`null`),
			wantErr: true,
		},
		{
			name:    "empty bytes",
			br:      BatchRequest{},
			args:    []byte(``),
			wantErr: true,
		},
		{
			name: "batch request contains both Request and Notification",
			br:   BatchRequest{},
			args: []byte(`[{"jsonrpc":"2.0","method":"test","params":[1,2,3],"id":1},{"jsonrpc":"2.0","method":"test","params":[1,2,3]}]`),
		},
		{
			name:    "batch request contains a string",
			br:      BatchRequest{},
			args:    []byte(`[{"jsonrpc":"2.0","method":"test","params":[1,2,3],"id":1},"test"]`),
			wantErr: true,
		},
		{
			name:    "batch contains bad Request and bad Notification",
			br:      BatchRequest{},
			args:    []byte(`[{"method":"test","params":[1,2,3],"id":1},{"method":"test","params":[1,2,3]}]`),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.br.UnmarshalJSON(tt.args)
			if tt.wantErr {
				require.Error(t, err, "unmarshalled object: %+v", tt.br)
			} else {
				require.NoError(t, err, "unmarshalled object: %+v", tt.br)
			}
		})
	}
}

func Test_optional_Defined(t *testing.T) {
	type O struct {
		ID optional[ID] `json:"id"`
	}
	b := []byte(`{"id":"test"}`)
	o := O{}
	err := json.Unmarshal(b, &o)
	require.NoError(t, err)
	require.True(t, o.ID.Defined)

	b = []byte(`{"id":null}`)
	o = O{}
	err = json.Unmarshal(b, &o)
	require.NoError(t, err)
	require.True(t, o.ID.Defined)

	b = []byte(`{}`)
	o = O{}
	err = json.Unmarshal(b, &o)
	require.NoError(t, err)
	require.False(t, o.ID.Defined)
}

func Test_Request_with_ParamsArray(t *testing.T) {
	r := Request{}
	err := json.Unmarshal([]byte(`{"jsonrpc":"2.0","method":"test","params":[1,2,3],"id":1}`), &r)
	require.NoError(t, err)
	require.Equal(t, ID("1"), r.ID)
	require.Equal(t, Method("test"), r.Method)
	require.Equal(t, &ParamsArray{1, 2, 3}, r.Params)

}

func Test_Notification_with_ParamsArray(t *testing.T) {
	n := Request{}
	err := json.Unmarshal([]byte(`{"jsonrpc":"2.0","method":"test","params":[1,2,3]}`), &n)
	require.NoError(t, err)
	require.Equal(t, Method("test"), n.Method)
	require.Equal(t, &ParamsArray{1, 2, 3}, n.Params)

}

func Test_BatchRequest_with_Notification(t *testing.T) {
	b := []byte(`[{"jsonrpc":"2.0","method":"test","params":[1,2,3]}]`)
	br := BatchRequest{}
	err := json.Unmarshal(b, &br)
	require.NoError(t, err)
	require.True(t, br[0].isNotification)
	require.Equal(t, Method("test"), br[0].Method)
	require.Equal(t, &ParamsArray{1, 2, 3}, br[0].Params)
}

func Test_BatchRequest_with_Request(t *testing.T) {
	b := []byte(`[{"jsonrpc":"2.0","method":"test","params":[1,2,3],"id":1}]`)
	br := BatchRequest{}
	err := json.Unmarshal(b, &br)
	require.NoError(t, err)
	require.False(t, br[0].isNotification)
	require.Equal(t, ID("1"), br[0].ID)
	require.Equal(t, Method("test"), br[0].Method)
	require.Equal(t, &ParamsArray{1, 2, 3}, br[0].Params)
}

func Test_BatchRequest_with_Mixed(t *testing.T) {
	b := []byte(`[{"jsonrpc":"2.0","method":"test","params":[1,2,3],"id":1},{"jsonrpc":"2.0","method":"test","params":[1,2,3]}]`)
	br := BatchRequest{}
	err := json.Unmarshal(b, &br)
	require.NoError(t, err)
	require.False(t, br[0].isNotification)
	require.True(t, br[1].isNotification)
}

func Test_Request_EndToEnd(t *testing.T) {
	r := Request{}
	err := json.Unmarshal([]byte(`{"jsonrpc":"2.0","method":"test","params":[1,2,3],"id":1}`), &r)
	require.NoError(t, err)
	b, err := json.Marshal(r)
	require.NoError(t, err)
	require.JSONEq(t, `{"jsonrpc":"2.0","method":"test","params":[1,2,3],"id":"1"}`, string(b))
}

func Test_Notification_EndToEnd(t *testing.T) {
	n := Request{}
	err := json.Unmarshal([]byte(`{"jsonrpc":"2.0","method":"test","params":[1,2,3]}`), &n)
	require.NoError(t, err)
	b, err := json.Marshal(n)
	require.NoError(t, err)
	require.JSONEq(t, `{"jsonrpc":"2.0","method":"test","params":[1,2,3]}`, string(b))
}

func Test_Response_EndToEnd(t *testing.T) {
	r := Response{}
	err := json.Unmarshal([]byte(`{"jsonrpc":"2.0","result":1,"id":1}`), &r)
	require.NoError(t, err)
	b, err := json.Marshal(r)
	require.NoError(t, err)
	require.JSONEq(t, `{"jsonrpc":"2.0","result":1,"id":"1"}`, string(b))
}

func TestResponse_without_result_result_not_null(t *testing.T) {
	r := Response{
		ID: ID("1234"),
		Error: &Error{
			Code:    12345,
			Message: "test",
		},
	}
	b, err := r.MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t, `{"jsonrpc":"2.0","error":{"code":12345,"message":"test"},"id":"1234"}`, string(b))
}

func TestResponse_without_result_or_error(t *testing.T) {
	r := Response{
		ID: ID("1234"),
	}
	b, err := r.MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t, `{"jsonrpc":"2.0","result":null,"id":"1234"}`, string(b))
}
