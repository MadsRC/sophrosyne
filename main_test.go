package sophrosyne

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAPIErrorMarshal(t *testing.T) {
	cases := []struct {
		name                    string
		input                   APIError
		ExpectedMarshalResult   string
		ExpectedUnmarshalResult APIError
	}{
		{
			name: "error with no suberrors",
			input: APIError{
				Code:    500,
				Message: "message",
			},
			ExpectedMarshalResult: `{"code":500,"message":"message"}`,
			ExpectedUnmarshalResult: APIError{
				Code:    500,
				Message: "message",
			},
		},
		{
			name: "error with single suberror",
			input: APIError{
				Code:    500,
				Message: "message",
				Errors: []APISubError{
					{Message: "error"},
				},
			},
			ExpectedMarshalResult: `{"code":500,"message":"error","errors":[{"message":"error"}]}`,
			ExpectedUnmarshalResult: APIError{
				Code:    500,
				Message: "error",
				Errors: []APISubError{
					{Message: "error"},
				},
			},
		},
		{
			name: "error with multiple suberrors",
			input: APIError{
				Code:    500,
				Message: "message",
				Errors: []APISubError{
					{Message: "error1"},
					{Message: "error2"},
				},
			},
			ExpectedMarshalResult: `{"code":500,"message":"error1","errors":[{"message":"error1"},{"message":"error2"}]}`,
			ExpectedUnmarshalResult: APIError{
				Code:    500,
				Message: "error1",
				Errors: []APISubError{
					{Message: "error1"},
					{Message: "error2"},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.input)
			require.NoError(t, err)
			require.NotNil(t, b)
			require.Equal(t, tc.ExpectedMarshalResult, string(b))

			newVal := APIError{}
			err = json.Unmarshal(b, &newVal)
			require.NoError(t, err)
			require.Equal(t, tc.ExpectedUnmarshalResult, newVal)
		})
	}
}

func FuzzAPIErrorMarshal(f *testing.F) {
	f.Fuzz(func(t *testing.T, input string) {
		val := APIError{
			Message: "something",
			Errors: []APISubError{
				{Message: input},
			},
		}
		b, err := json.Marshal(val)
		require.NoError(t, err)
		require.NotNil(t, b)

		fuzzValAsJSON, err := json.Marshal(input)
		require.NoError(t, err)
		t.Logf("fuzzValAsJSON: %s", fuzzValAsJSON)
		t.Logf("b: %s", b)
		if string(fuzzValAsJSON) == `""` {
			t.Log("a")
			require.Equal(t, []byte(`{"errors":[{}]}`), b)
		} else {
			t.Log("b")
			require.Equal(t, []byte(`{"message":`+string(fuzzValAsJSON)+`,"errors":[{"message":`+string(fuzzValAsJSON)+`}]}`), b)
		}
	})
}

func TestAPIResponseMarshal(t *testing.T) {
	cases := []struct {
		name                  string
		input                 APIResponse
		ExpectedMarshalResult string
		ExpectUnmarshalResult APIResponse
	}{
		{
			name: "response with no error",
			input: APIResponse{
				APIVersion: "1.0",
				Data:       &APIResponseData{Kind: "i should be included"},
			},
			ExpectedMarshalResult: `{"apiVersion":"1.0","data":{"kind":"i should be included"}}`,
			ExpectUnmarshalResult: APIResponse{
				APIVersion: "1.0",
				Data:       &APIResponseData{Kind: "i should be included"},
			},
		},
		{
			name: "response with error",
			input: APIResponse{
				APIVersion: "1.0",
				Data:       &APIResponseData{Kind: "i should be ignored"},
				Error: &APIError{
					Code:    500,
					Message: "error",
				},
			},
			ExpectedMarshalResult: `{"apiVersion":"1.0","error":{"code":500,"message":"error"}}`,
			ExpectUnmarshalResult: APIResponse{
				APIVersion: "1.0",
				Error: &APIError{
					Code:    500,
					Message: "error",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.input)
			require.NoError(t, err)
			require.NotNil(t, b)
			require.Equal(t, tc.ExpectedMarshalResult, string(b))

			newVal := APIResponse{}
			err = json.Unmarshal(b, &newVal)
			require.NoError(t, err)
			require.Equal(t, tc.ExpectUnmarshalResult, newVal)
		})
	}
}
