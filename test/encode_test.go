package decode

import (
	"bytes"
	"testing"

	"github.com/Moonlight-Companies/goresp/resp"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		name     string
		input    resp.RESPValue
		expected []byte
	}{
		{
			name:     "Simple String",
			input:    &resp.RESPSimpleString{Value: "OK"},
			expected: []byte("+OK\r\n"),
		},
		{
			name:     "Error",
			input:    &resp.RESPError{Value: "Error message"},
			expected: []byte("-Error message\r\n"),
		},
		{
			name:     "Integer",
			input:    &resp.RESPInteger{Value: 1000},
			expected: []byte(":1000\r\n"),
		},
		{
			name:     "Bulk String",
			input:    &resp.RESPBulkString{Value: []byte("hello")},
			expected: []byte("$5\r\nhello\r\n"),
		},
		{
			name:     "Empty Bulk String",
			input:    &resp.RESPBulkString{Value: []byte("")},
			expected: []byte("$0\r\n\r\n"),
		},
		{
			name:     "Null Bulk String",
			input:    &resp.RESPBulkString{Value: nil},
			expected: []byte("$-1\r\n"),
		},
		{
			name: "Array",
			input: &resp.RESPArray{
				Items: []resp.RESPValue{
					&resp.RESPSimpleString{Value: "OK"},
					&resp.RESPInteger{Value: 1000},
					&resp.RESPBulkString{Value: []byte("hello")},
				},
			},
			expected: []byte("*3\r\n+OK\r\n:1000\r\n$5\r\nhello\r\n"),
		},
		{
			name:     "Empty Array",
			input:    &resp.RESPArray{Items: []resp.RESPValue{}},
			expected: []byte("*0\r\n"),
		},
		{
			name:     "Null Array",
			input:    &resp.RESPArray{Items: nil},
			expected: []byte("*-1\r\n"),
		},
		{
			name: "Nested Array",
			input: &resp.RESPArray{
				Items: []resp.RESPValue{
					&resp.RESPArray{
						Items: []resp.RESPValue{
							&resp.RESPSimpleString{Value: "nested"},
							&resp.RESPInteger{Value: 42},
						},
					},
					&resp.RESPBulkString{Value: []byte("outer")},
				},
			},
			expected: []byte("*2\r\n*2\r\n+nested\r\n:42\r\n$5\r\nouter\r\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			err := tt.input.Encode(buf)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tt.expected) {
				t.Errorf("Encode() = %v, want %v", buf.Bytes(), tt.expected)
			}
		})
	}
}
