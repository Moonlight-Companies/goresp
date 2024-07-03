package decode

import (
	"testing"

	"github.com/iceisfun/goresp/resp"
)

func TestDecode(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected resp.RESPValue
		wantErr  bool
		wantMore bool
	}{
		{
			name:     "Simple String",
			input:    []byte("+OK\r\n"),
			expected: &resp.RESPSimpleString{Value: "OK"},
			wantErr:  false,
		},
		{
			name:     "Error",
			input:    []byte("-Error message\r\n"),
			expected: &resp.RESPError{Value: "Error message"},
			wantErr:  false,
		},
		{
			name:     "Integer",
			input:    []byte(":1000\r\n"),
			expected: &resp.RESPInteger{Value: 1000},
			wantErr:  false,
		},
		{
			name:     "Bulk String",
			input:    []byte("$5\r\nhello\r\n"),
			expected: &resp.RESPBulkString{Value: []byte("hello")},
			wantErr:  false,
		},
		{
			name:     "Null Bulk String",
			input:    []byte("$-1\r\n"),
			expected: &resp.RESPBulkString{Value: []byte("")},
			wantErr:  false,
		},
		{
			name:     "Empty Bulk String",
			input:    []byte("$0\r\n\r\n"),
			expected: &resp.RESPBulkString{Value: []byte("")},
			wantErr:  false,
		},
		{
			name:  "Array",
			input: []byte("*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n"),
			expected: &resp.RESPArray{
				Items: []resp.RESPValue{
					&resp.RESPBulkString{Value: []byte("hello")},
					&resp.RESPBulkString{Value: []byte("world")},
				},
			},
			wantErr: false,
		},
		{
			name:     "Empty Array",
			input:    []byte("*0\r\n"),
			expected: &resp.RESPArray{Items: []resp.RESPValue{}},
			wantErr:  false,
		},
		{
			name:     "Null Array",
			input:    []byte("*-1\r\n"),
			expected: &resp.RESPArray{Items: nil},
			wantErr:  false,
		},
		{
			name:  "Nested Array",
			input: []byte("*2\r\n*2\r\n+OK\r\n:1000\r\n$5\r\nhello\r\n"),
			expected: &resp.RESPArray{
				Items: []resp.RESPValue{
					&resp.RESPArray{
						Items: []resp.RESPValue{
							&resp.RESPSimpleString{Value: "OK"},
							&resp.RESPInteger{Value: 1000},
						},
					},
					&resp.RESPBulkString{Value: []byte("hello")},
				},
			},
			wantErr: false,
		},
		{
			name:     "Invalid input - no CRLF",
			input:    []byte("+OK"),
			expected: nil,
			wantErr:  false,
			wantMore: true,
		},
		{
			name:     "Invalid input - unknown type",
			input:    []byte("X123\r\n"),
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Invalid input - incomplete bulk string",
			input:    []byte("$5\r\nhell"),
			expected: nil,
			wantErr:  false,
			wantMore: true,
		},
		{
			name:     "Invalid input - incomplete array",
			input:    []byte("*2\r\n$5\r\nhello\r\n"),
			expected: nil,
			wantErr:  false,
			wantMore: true,
		},
		{
			name:     "Incomplete input - need more data",
			input:    []byte("$"),
			expected: nil,
			wantErr:  false,
			wantMore: true,
		},
		{
			name:     "Incomplete input - partial bulk string",
			input:    []byte("$5\r\nhel"),
			expected: nil,
			wantErr:  false,
			wantMore: true,
		},
		{
			name:     "Invalid opcode",
			input:    []byte("%"),
			expected: nil,
			wantErr:  true,
			wantMore: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dec := resp.NewDecode()
			dec.Provide(tt.input)
			got, err := dec.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantMore {
				if got != nil || err != nil {
					t.Errorf("Decode() = %v, %v; want nil, nil for incomplete input", got, err)
				}
				return
			}

			if got == nil && tt.expected != nil {
				t.Errorf("Decode() returned nil, expected %v", tt.expected)
				return
			}

			if got != nil && tt.expected == nil {
				t.Errorf("Decode() returned %v, expected nil", got)
				return
			}

			if got != nil && tt.expected != nil && !got.Equal(tt.expected) {
				t.Errorf("Decode() = %v, want %v", got, tt.expected)
			}
		})
	}
}
