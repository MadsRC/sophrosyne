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

package jsonrpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// optional is a generic type that represents an optional value in JSON. It is used to represent fields that are
// optional in the JSON-RPC 2.0 specification, such as the "id" field of a [Request].
//
// Implementation is from https://stackoverflow.com/questions/36601367/json-field-set-to-null-vs-field-not-there
type optional[T any] struct {
	Defined bool
	Value   *T
}

// UnmarshalJSON is implemented by deferring to the wrapped type (T).
// It will be called only if the value is defined in the JSON payload.
func (o *optional[T]) UnmarshalJSON(data []byte) error {
	o.Defined = true
	return json.Unmarshal(data, &o.Value)
}

// JSONRPC represents the jsonrpc field of a [Request], [Notification], or [Response]. For the JSON-RPC 2.0
// specification, this field MUST be exactly "2.0" and the use of the [JSONRPC2_0] constant is thus recommended.
type JSONRPC string

const (
	// JSONRPC2_0 is the identifier for the JSON-RPC 2.0 specification.
	JSONRPC2_0 JSONRPC = "2.0"
)

// Method represents the method field of a [Request] or [Notification] as per the JSON-RPC 2.0 specification.
//
// A String containing the name of the method to be invoked. Method names that begin with the word rpc followed by a
// period character (U+002E or ASCII 46) are reserved for rpc-internal methods and extensions and MUST NOT be used for
// anything else.
type Method string

// UnmarshalJSON unmarshals a JSON object into a [Method]. If the value is prefixed with "rpc.", it is considered
// invalid and an error is returned.
func (m *Method) UnmarshalJSON(data []byte) error {
	var value string
	// Ignoring error since this shouldn't error out.
	_ = json.Unmarshal(data, &value)

	if strings.HasPrefix(value, "rpc.") {
		return fmt.Errorf("method names that begin with 'rpc.' are reserved for rpc-internal methods and extensions")
	}

	*m = Method(value)
	return nil
}

// Params represents the Params field of a [Request] or [Notification] as per the JSON-RPC 2.0 specification section
// 4.2.
//
// If present, parameters for the rpc call MUST be provided as a Structured value. Either by-position through an Array
// or by-name through an Object.
//
//	by-position: Params MUST be an Array, containing the values in the Server expected order.
//	by-name: Params MUST be an Object, with member names that match the Server expected parameter names. The absence of
//	expected names MAY result in an error being generated. The names MUST match exactly, including case, to the method's
//	expected parameters.
//
// Implementations include [ParamsObject] and [ParamsArray] to represent the two types of Params.
type Params interface {
	// IsParams is a marker method to determine if the struct is a Params object.
	isParams()
}

// ParamsObject represents a by-name Params object as per the JSON-RPC 2.0 specification section 4.2.
//
// It implements the private [Params] interface, and as such can be used as a value for the Params field of a [Request]
// or [Notification].
type ParamsObject map[string]interface{}

// UnmarshalJSON unmarshalls a JSON object into a [ParamsObject]. This is necessary because the JSON-RPC 2.0
// specification allows for the Params field to be either an object or an array, and the Go JSON unmarshaller cannot
// unmarshal into an interface{}.
//
// If an element is a JSON number, its mathematical value is referenced in order to determine if it is an integer or a
// float. If it is an integer, it is converted into an int64, otherwise it is converted into a float64. This is in
// contrast to the Go JSON unmarshaller, which unmarshals all numbers into float64.
func (p *ParamsObject) UnmarshalJSON(data []byte) error {
	var obj map[string]*json.RawMessage
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return err
	}

	*p = make(ParamsObject)

	for key, raw := range obj {
		var value interface{}
		// Skipping error check since this shouldn't error out.
		_ = json.Unmarshal(*raw, &value)

		switch value := value.(type) {
		case float64:
			if value == float64(int(value)) {
				(*p)[key] = int(value)
			} else {
				(*p)[key] = value
			}
		default:
			(*p)[key] = value
		}
	}

	return nil
}

func (*ParamsObject) isParams() {
	// Used to implement the [Params] interface.
}

// ParamsArray represents a by-position Params array as per the JSON-RPC 2.0 specification section 4.2.
//
// It implements the private [Params] interface, and as such can be used as a value for the Params field of a [Request]
// or [Notification].
type ParamsArray []interface{}

func (*ParamsArray) isParams() {
	// Used to implement the [Params] interface.
}

// UnmarshalJSON unmarshals a JSON object into a [ParamsArray]. This is necessary because the JSON-RPC 2.0 specification
// allows for the Params field to be either an object or an array, and the Go JSON unmarshaller cannot unmarshal into an
// interface{}.
//
// If an element is a JSON number, its mathematical value is referenced in order to determine if it is an integer or a
// float. If it is an integer, it is converted into an int64, otherwise it is converted into a float64. This is in
// contrast to the Go JSON unmarshaller, which unmarshals all numbers into float64.
func (p *ParamsArray) UnmarshalJSON(data []byte) error {
	var arr []json.RawMessage
	err := json.Unmarshal(data, &arr)
	if err != nil {
		return err
	}

	for _, raw := range arr {
		var value interface{}
		// Skipping error check since this shouldn't error out.
		_ = json.Unmarshal(raw, &value)

		switch value := value.(type) {
		case float64:
			if value == float64(int(value)) {
				*p = append(*p, int(value))
			} else {
				*p = append(*p, value)
			}
		default:
			*p = append(*p, value)
		}
	}

	return nil
}

// ID represents the id field of a [Request] or [Response] as per the JSON-RPC 2.0 specification.
//
// The specification mandates that the id field MUST contain a String, Number, or Null value.  The use of a Null is
// discouraged, as it is used for Responses with an unknown id and thus can cause confusion. Number should not contain
// fractions to avoid issues with binary fractions.
//
// To simplify the implementation, this library uses a string type for the id field. When marshalling, the value is
// always marshalled into a string.
type ID string

// UnmarshalJSON unmarshals a JSON object into an [ID]. If the value is "null", it is unmarshalled into an empty string.
// If the value is a number, it is unmarshalled into a string.
func (id *ID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*id = ""
		return nil
	}

	var value string
	err := json.Unmarshal(data, &value)
	if err != nil {
		var number float64
		err = json.Unmarshal(data, &number)
		if err != nil {
			if strings.HasPrefix(string(data), `{`) || strings.HasPrefix(string(data), `[`) {
				return fmt.Errorf("id must be a string, number, or null")
			}
		}

		*id = ID(fmt.Sprintf("%v", int(number)))
		return nil
	}

	*id = ID(value)
	return nil
}

// Request represents a Request object as per the JSON-RPC 2.0 specification.
//
// A rpc call is represented by sending a Request object to a Server. The Request object has the following members:
//
// jsonrpc
//
//	A String specifying the version of the JSON-RPC protocol. MUST be exactly "2.0".
//
// method
//
//	A String containing the name of the method to be invoked. Method names that begin with the word rpc followed by a
//	period character (U+002E or ASCII 46) are reserved for rpc-internal methods and extensions and MUST NOT be used for
//	anything else.
//
// Params
//
//	A Structured value that holds the parameter values to be used during the invocation of the method. This member MAY
//	be omitted.
//
// id
//
//	An identifier established by the Client that MUST contain a String, Number, or NULL value if included. If it is not
//	included it is assumed to be a notification. The value SHOULD normally not be Null [1] and Numbers SHOULD NOT
//	contain fractional parts [2]
//
// The Server MUST reply with the same value in the [Response] object if included. This member is used to correlate the
// context between the two objects.
//
// [1] The use of Null as a value for the id member in a Request object is discouraged, because this specification uses
// a value of Null for Responses with an unknown id. Also, because JSON-RPC 1.0 uses an id value of Null for
// Notifications this could cause confusion in handling.
//
// [2] Fractional parts may be problematic, since many decimal fractions cannot be represented exactly as binary
// fractions.
type Request struct {
	isNotification bool
	Method         Method `json:"method" validate:"required"`
	Params         Params `json:"params,omitempty"`
	ID             ID     `json:"id"`
}

func (r Request) IsNotification() bool {
	return r.isNotification
}

func (r *Request) AsNotification() *Request {
	r.isNotification = true
	return r
}

// MarshalJSON marshals a [Request] object into a JSON object. The field "jsonrpc" is added and set to the value of
// [JSONRPC2_0].
func (r Request) MarshalJSON() ([]byte, error) {
	if r.isNotification {
		return json.Marshal(&struct {
			JSONRPC JSONRPC `json:"jsonrpc"`
			Method  Method  `json:"method" validate:"required"`
			Params  Params  `json:"params,omitempty"`
		}{
			JSONRPC: JSONRPC2_0,
			Method:  r.Method,
			Params:  r.Params,
		})
	} else {
		return json.Marshal(&struct {
			JSONRPC JSONRPC `json:"jsonrpc"`
			Method  Method  `json:"method" validate:"required"`
			Params  Params  `json:"params,omitempty"`
			ID      ID      `json:"id"`
		}{
			JSONRPC: JSONRPC2_0,
			Method:  r.Method,
			Params:  r.Params,
			ID:      r.ID,
		})
	}
}

func (r *Request) UnmarshalJSON(data []byte) error {
	var dat map[string]*json.RawMessage
	err := json.Unmarshal(data, &dat)
	if err != nil {
		return err
	}

	v, ok := dat["jsonrpc"]
	if ok {
		var version JSONRPC
		err = json.Unmarshal(*v, &version)
		if err != nil {
			return err
		}
		if version != JSONRPC2_0 {
			return fmt.Errorf("invalid JSON-RPC version: %s", version)
		}
	} else {
		return fmt.Errorf("jsonrpc field is required")
	}

	if _, ok := dat["id"]; ok {
		err = json.Unmarshal(*dat["id"], &r.ID)
		if err != nil {
			return err
		}
	} else {
		r.isNotification = true
	}

	if _, ok := dat["method"]; ok {
		err = json.Unmarshal(*dat["method"], &r.Method)
		if err != nil {
			return err
		}
	}

	if r.Method == "" {
		return fmt.Errorf("method is required")
	}

	// decode Params into a ParamsObject if it is an object, otherwise decode it into a ParamsArray.
	if _, ok := dat["params"]; ok {
		if dat["params"] != nil {
			var obj ParamsObject
			err = json.Unmarshal(*dat["params"], &obj)
			if err == nil {
				r.Params = &obj
			} else {
				var arr ParamsArray
				err = json.Unmarshal(*dat["params"], &arr)
				if err == nil {
					r.Params = &arr
				}
			}
		}
	}

	return nil
}

// Response represents a Response object as per the JSON-RPC 2.0 specification.
//
// When a rpc call is made, the Server MUST reply with a Response, except for in the case of Notifications. The Response
// is expressed as a single JSON Object, with the following members:
//
// jsonrpc
//
//	A String specifying the version of the JSON-RPC protocol. MUST be exactly "2.0".
//
// result
//
//	This member is REQUIRED on success.
//	This member MUST NOT exist if there was an error invoking the method.
//	The value of this member is determined by the method invoked on the Server.
//
// error
//
//	This member is REQUIRED on error.
//	This member MUST NOT exist if there was no error triggered during invocation.
//	The value for this member MUST be an Object as defined in section 5.1.
//
// id
//
//	This member is REQUIRED.
//	It MUST be the same as the value of the id member in the Request Object.
//	If there was an error in detecting the id in the Request object (e.g. Parse error/Invalid Request), it MUST be Null.
//
// Either the result member or error member MUST be included, but both members MUST NOT be included.
type Response struct {
	Result interface{} `json:"result,omitempty"`
	Error  *Error      `json:"error,omitempty" validate:"required_without=Result,excluded_with=Result"`
	ID     ID          `json:"id"`
}

// MarshalJSON marshals a [Response] object into a JSON object. The field "jsonrpc" is added and set to the value of
// [JSONRPC2_0].
func (r Response) MarshalJSON() ([]byte, error) {
	if r.Error != nil {
		r.Result = nil // Error takes precedence over Result
	}

	if r.Result == nil && r.Error == nil {
		r.Result = json.RawMessage("null")
	}

	type Alias Response
	return json.Marshal(&struct {
		JSONRPC JSONRPC `json:"jsonrpc"`
		*Alias
	}{
		JSONRPC: JSONRPC2_0,
		Alias:   (*Alias)(&r),
	})
}

func (r *Response) UnmarshalJSON(data []byte) error {
	type temp struct {
		ID     optional[ID]          `json:"id"`
		Result optional[interface{}] `json:"result"`
		Error  optional[Error]       `json:"error"`
	}
	var dat map[string]*json.RawMessage
	err := json.Unmarshal(data, &dat)
	if err != nil {
		return err
	}

	v, ok := dat["jsonrpc"]
	if ok {
		var version JSONRPC
		err = json.Unmarshal(*v, &version)
		if err != nil {
			return err
		}
		if version != JSONRPC2_0 {
			return fmt.Errorf("invalid JSON-RPC version: %s", version)
		}
	} else {
		return fmt.Errorf("jsonrpc field is required")
	}

	var tmp temp
	err = json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	if !tmp.Result.Defined && !tmp.Error.Defined {
		return fmt.Errorf("either result or error is required")
	}
	if tmp.Result.Value == nil && tmp.Error.Value == nil {
		return fmt.Errorf("either result or error is required")
	}
	if tmp.Result.Defined {
		r.Result = tmp.Result.Value
	}
	if tmp.Error.Defined {
		r.Error = tmp.Error.Value

	}

	if tmp.ID.Defined {
		if tmp.ID.Value == nil {
			r.ID = ""
		} else {
			r.ID = *tmp.ID.Value
		}
	} else {
		return fmt.Errorf("id is required")
	}

	return nil
}

// Error represents an Error object as per the JSON-RPC 2.0 specification section 5.1.
//
// When a rpc call encounters an error, the [Response] Object MUST contain the error member with a value that is a
// Object with the following members:
//
// code
//
//	A Number that indicates the error type that occurred.
//	This MUST be an integer.
//
// message
//
//	A String providing a short description of the error.
//	The message SHOULD be limited to a concise single sentence.
//
// data
//
//	A Primitive or Structured value that contains additional information about the error.
//	This may be omitted.
//	The value of this member is defined by the Server (e.g. detailed error information, nested errors etc.).
//
// The error codes from and including -32768 to -32000 are reserved for pre-defined errors. Any code within this range,
// but not defined explicitly below is reserved for future use. The error codes are nearly the same as those suggested
// for XML-RPC at the following url: http://xmlrpc-epi.sourceforge.net/specs/rfc.fault_codes.php
//
// code               message             meaning
// -32700 	          Parse error         Invalid JSON was received by the server. An error occurred on the server while parsing the JSON text.
// -32600             Invalid Request     The JSON sent is not a valid Request object.
// -32601             Method not found    The method does not exist / is not available.
// -32602             Invalid Params      Invalid method parameter(s).
// -32603             Internal error      Internal JSON-RPC error.
// -32000 to -32099   Server error        Reserved for implementation-defined server-errors.
//
// The remainder of the space is available for application defined errors.
type Error struct {
	Code    RPCErrorCode `json:"code" validate:"required"`
	Message string       `json:"message" validate:"required"`
	Data    interface{}  `json:"data,omitempty"`
}

// BatchRequest represents a Batch Request as per the JSON-RPC 2.0 specification.
//
// To send several [Request] objects at the same time, the Client MAY send an Array filled with [Request] objects.
//
// The Server should respond with an Array containing the corresponding [Response] objects, after all of the batch
// [Request] objects have been processed. A [Response] object SHOULD exist for each [Request] object, except that there
// SHOULD NOT be any [Response] objects for notifications. The Server MAY process a batch rpc call as a set of
// concurrent tasks, processing them in any order and with any width of parallelism.
//
// The [Response] objects being returned from a batch call MAY be returned in any order within the Array. The Client
// SHOULD match contexts between the set of [Request] objects and the resulting set of [Response] objects based on the
// id member within each Object.
//
// If the batch rpc call itself fails to be recognized as an valid JSON or as an Array with at least one value, the
// response from the Server MUST be a single [Response] object. If there are no [Response] objects contained within the
// [Response] array as it is to be sent to the client, the server MUST NOT return an empty Array and should return
// nothing at all.
type BatchRequest []Request

func (b *BatchRequest) UnmarshalJSON(data []byte) error {
	var arr []json.RawMessage
	err := json.Unmarshal(data, &arr)
	if err != nil {
		return err
	}

	type O struct {
		ID optional[ID] `json:"id"`
	}

	if arr == nil {
		return fmt.Errorf("batch request must be an array")
	}

	var me error

	for i, raw := range arr {
		var obj O
		err = json.Unmarshal(raw, &obj)
		if err != nil {
			me = errors.Join(fmt.Errorf("error unmarshalling object at index %d: %v", i, err)) // nolint:errorlint
			continue
		}
		var req Request
		if !obj.ID.Defined {
			// It is a notification
			req.isNotification = true
		}
		err = json.Unmarshal(raw, &req)
		if err != nil {
			me = errors.Join(fmt.Errorf("error unmarshalling object at index %d into Request: %v", i, err)) // nolint:errorlint
		} else {
			*b = append(*b, req)
		}
	}

	return me
}

// BatchResponse represents a Batch Response as per the JSON-RPC 2.0 specification.
//
// To send several [Request] objects at the same time, the Client MAY send an Array filled with [Request] objects.
//
// The Server should respond with an Array containing the corresponding [Response] objects, after all of the batch
// [Request] objects have been processed. A [Response] object SHOULD exist for each [Request] object, except that there
// SHOULD NOT be any [Response] objects for notifications. The Server MAY process a batch rpc call as a set of
// concurrent tasks, processing them in any order and with any width of parallelism.
//
// The [Response] objects being returned from a batch call MAY be returned in any order within the Array. The Client
// SHOULD match contexts between the set of [Request] objects and the resulting set of [Response] objects based on the
// id member within each Object.
//
// If the batch rpc call itself fails to be recognized as an valid JSON or as an Array with at least one value, the
// response from the Server MUST be a single [Response] object. If there are no [Response] objects contained within the
// [Response] array as it is to be sent to the client, the server MUST NOT return an empty Array and should return
// nothing at all.
type BatchResponse []Response

// RPCErrorCode represents an error code as per the JSON-RPC 2.0 specification section 5.1.
type RPCErrorCode int

const (
	// ParseError signals that an invalid JSON was received by the server and that an error occurred on the server while
	// parsing the JSON text.
	ParseError RPCErrorCode = -32700
	// InvalidRequest signals that the JSON sent is not a valid Request object.
	InvalidRequest RPCErrorCode = -32600
	// MethodNotFound signals that the method does not exist / is not available.
	MethodNotFound RPCErrorCode = -32601
	// InvalidParams signals that invalid method parameter(s) was given.
	InvalidParams RPCErrorCode = -32602
	// InternalError signals an internal JSON-RPC error.
	InternalError RPCErrorCode = -32603
	// ServerError0 to ServerError99 are reserved for implementation-defined server-errors.
	ServerError0  RPCErrorCode = -32000
	ServerError1  RPCErrorCode = -32001
	ServerError2  RPCErrorCode = -32002
	ServerError3  RPCErrorCode = -32003
	ServerError4  RPCErrorCode = -32004
	ServerError5  RPCErrorCode = -32005
	ServerError6  RPCErrorCode = -32006
	ServerError7  RPCErrorCode = -32007
	ServerError8  RPCErrorCode = -32008
	ServerError9  RPCErrorCode = -32009
	ServerError10 RPCErrorCode = -32010
	ServerError11 RPCErrorCode = -32011
	ServerError12 RPCErrorCode = -32012
	ServerError13 RPCErrorCode = -32013
	ServerError14 RPCErrorCode = -32014
	ServerError15 RPCErrorCode = -32015
	ServerError16 RPCErrorCode = -32016
	ServerError17 RPCErrorCode = -32017
	ServerError18 RPCErrorCode = -32018
	ServerError19 RPCErrorCode = -32019
	ServerError20 RPCErrorCode = -32020
	ServerError21 RPCErrorCode = -32021
	ServerError22 RPCErrorCode = -32022
	ServerError23 RPCErrorCode = -32023
	ServerError24 RPCErrorCode = -32024
	ServerError25 RPCErrorCode = -32025
	ServerError26 RPCErrorCode = -32026
	ServerError27 RPCErrorCode = -32027
	ServerError28 RPCErrorCode = -32028
	ServerError29 RPCErrorCode = -32029
	ServerError30 RPCErrorCode = -32030
	ServerError31 RPCErrorCode = -32031
	ServerError32 RPCErrorCode = -32032
	ServerError33 RPCErrorCode = -32033
	ServerError34 RPCErrorCode = -32034
	ServerError35 RPCErrorCode = -32035
	ServerError36 RPCErrorCode = -32036
	ServerError37 RPCErrorCode = -32037
	ServerError38 RPCErrorCode = -32038
	ServerError39 RPCErrorCode = -32039
	ServerError40 RPCErrorCode = -32040
	ServerError41 RPCErrorCode = -32041
	ServerError42 RPCErrorCode = -32042
	ServerError43 RPCErrorCode = -32043
	ServerError44 RPCErrorCode = -32044
	ServerError45 RPCErrorCode = -32045
	ServerError46 RPCErrorCode = -32046
	ServerError47 RPCErrorCode = -32047
	ServerError48 RPCErrorCode = -32048
	ServerError49 RPCErrorCode = -32049
	ServerError50 RPCErrorCode = -32050
	ServerError51 RPCErrorCode = -32051
	ServerError52 RPCErrorCode = -32052
	ServerError53 RPCErrorCode = -32053
	ServerError54 RPCErrorCode = -32054
	ServerError55 RPCErrorCode = -32055
	ServerError56 RPCErrorCode = -32056
	ServerError57 RPCErrorCode = -32057
	ServerError58 RPCErrorCode = -32058
	ServerError59 RPCErrorCode = -32059
	ServerError60 RPCErrorCode = -32060
	ServerError61 RPCErrorCode = -32061
	ServerError62 RPCErrorCode = -32062
	ServerError63 RPCErrorCode = -32063
	ServerError64 RPCErrorCode = -32064
	ServerError65 RPCErrorCode = -32065
	ServerError66 RPCErrorCode = -32066
	ServerError67 RPCErrorCode = -32067
	ServerError68 RPCErrorCode = -32068
	ServerError69 RPCErrorCode = -32069
	ServerError70 RPCErrorCode = -32070
	ServerError71 RPCErrorCode = -32071
	ServerError72 RPCErrorCode = -32072
	ServerError73 RPCErrorCode = -32073
	ServerError74 RPCErrorCode = -32074
	ServerError75 RPCErrorCode = -32075
	ServerError76 RPCErrorCode = -32076
	ServerError77 RPCErrorCode = -32077
	ServerError78 RPCErrorCode = -32078
	ServerError79 RPCErrorCode = -32079
	ServerError80 RPCErrorCode = -32080
	ServerError81 RPCErrorCode = -32081
	ServerError82 RPCErrorCode = -32082
	ServerError83 RPCErrorCode = -32083
	ServerError84 RPCErrorCode = -32084
	ServerError85 RPCErrorCode = -32085
	ServerError86 RPCErrorCode = -32086
	ServerError87 RPCErrorCode = -32087
	ServerError88 RPCErrorCode = -32088
	ServerError89 RPCErrorCode = -32089
	ServerError90 RPCErrorCode = -32090
	ServerError91 RPCErrorCode = -32091
	ServerError92 RPCErrorCode = -32092
	ServerError93 RPCErrorCode = -32093
	ServerError94 RPCErrorCode = -32094
	ServerError95 RPCErrorCode = -32095
	ServerError96 RPCErrorCode = -32096
	ServerError97 RPCErrorCode = -32097
	ServerError98 RPCErrorCode = -32098
	ServerError99 RPCErrorCode = -32099
)

const ServerErrorMessage = "Server error"

// RPCErrorMessage represents an error message as per the JSON-RPC 2.0 specification section 5.1.
type RPCErrorMessage string

const (
	// ParseErrorMessage is the message for [ParseError].
	ParseErrorMessage RPCErrorMessage = "Parse error"
	// InvalidRequestMessage is the message for [InvalidRequest].
	InvalidRequestMessage RPCErrorMessage = "Invalid Request"
	// MethodNotFoundMessage is the message for [MethodNotFound].
	MethodNotFoundMessage RPCErrorMessage = "Method not found"
	// InvalidParamsMessage is the message for [InvalidParams].
	InvalidParamsMessage RPCErrorMessage = "Invalid Params"
	// InternalErrorMessage is the message for [InternalError].
	InternalErrorMessage RPCErrorMessage = "Internal error"
	// ServerErrorMessage0 to ServerErrorMessage99 are reserved for implementation-defined server-errors.
	ServerErrorMessage0  RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage1  RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage2  RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage3  RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage4  RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage5  RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage6  RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage7  RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage8  RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage9  RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage10 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage11 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage12 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage13 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage14 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage15 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage16 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage17 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage18 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage19 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage20 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage21 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage22 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage23 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage24 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage25 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage26 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage27 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage28 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage29 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage30 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage31 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage32 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage33 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage34 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage35 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage36 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage37 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage38 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage39 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage40 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage41 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage42 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage43 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage44 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage45 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage46 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage47 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage48 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage49 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage50 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage51 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage52 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage53 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage54 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage55 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage56 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage57 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage58 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage59 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage60 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage61 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage62 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage63 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage64 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage65 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage66 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage67 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage68 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage69 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage70 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage71 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage72 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage73 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage74 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage75 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage76 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage77 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage78 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage79 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage80 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage81 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage82 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage83 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage84 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage85 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage86 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage87 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage88 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage89 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage90 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage91 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage92 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage93 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage94 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage95 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage96 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage97 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage98 RPCErrorMessage = ServerErrorMessage
	ServerErrorMessage99 RPCErrorMessage = ServerErrorMessage
)

// ValidateMethod validates the method field of a [Request], [Notification], or [Response].
//
// If [github.com/go-playground/validator] is used, this function can be used as a custom validation function.
//
// Example:
//
//	validate = validator.New()
//	validate.RegisterValidation(jsonrpc.ValidateMethod, jsonrpc.Method)
func ValidateMethod(field reflect.Value) interface{} {
	if field.Kind() != reflect.String {
		return "method must be a string"
	}
	if field.String() == "" {
		return "method cannot be empty"
	}

	if strings.HasPrefix(field.String(), "rpc.") {
		return "methods beginning with rpc. are reserved"
	}

	return nil
}

// ValidateErrorCode validates an [RPCErrorCode].
//
// If [github.com/go-playground/validator] is used, this function can be used as a custom validation function.
//
// Example:
//
//	validate = validator.New()
//	validate.RegisterValidation(jsonrpc.ValidateErrorCode, jsonrpc.RPCErrorCode)
func ValidateErrorCode(field reflect.Value) interface{} {
	if field.Kind() != reflect.Int {
		return "error code must be an integer"
	}

	for _, code := range []RPCErrorCode{ParseError, InvalidRequest, MethodNotFound, InvalidParams, InternalError} {
		if RPCErrorCode(field.Int()) == code {
			return nil
		}
	}

	if field.Int() <= -32000 && field.Int() >= -32768 {
		// This code is reserved for future use.
		return fmt.Sprintf("invalid error code - code %d is reserved for future use", field.Int())
	}

	return nil
}
