package decode

import (
	"testing"

	"github.com/Moonlight-Companies/goresp/resp"
)

func TestDecodeStream(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected resp.RESPValue
		wantErr  bool
	}{
		{
			name:     "Invalid Opcode",
			input:    "^",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Simple String",
			input:    "+OK\r\n",
			expected: &resp.RESPSimpleString{Value: "OK"},
		},
		{
			name:     "Error",
			input:    "-Error message\r\n",
			expected: &resp.RESPError{Value: "Error message"},
		},
		{
			name:     "Integer",
			input:    ":1000\r\n",
			expected: &resp.RESPInteger{Value: 1000},
		},
		{
			name:     "Bulk String",
			input:    "$5\r\nhello\r\n",
			expected: &resp.RESPBulkString{Value: []byte("hello")},
		},
		{
			name:  "Array",
			input: "*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n",
			expected: &resp.RESPArray{
				Items: []resp.RESPValue{
					&resp.RESPBulkString{Value: []byte("hello")},
					&resp.RESPBulkString{Value: []byte("world")},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dec := resp.NewDecode()
			var got resp.RESPValue
			var err error

			for i := 0; i < len(tt.input); i++ {
				dec.Provide([]byte{tt.input[i]})
				got, err = dec.Parse()

				if tt.wantErr {
					if err == nil {
						t.Errorf("Decode() at byte %d expected error, got nil", i)
					}
					return
				}

				if i < len(tt.input)-1 {
					if got != nil || err != nil {
						t.Errorf("Decode() at byte %d = %v, %v; want nil, nil", i, got, err)
						return
					}
				}
			}

			if tt.wantErr {
				t.Errorf("Decode() expected error, got nil")
				return
			}

			if err != nil {
				t.Errorf("Decode() error = %v", err)
				return
			}

			if got == nil {
				t.Errorf("Decode() returned nil, expected %v", tt.expected)
				return
			}

			if !got.Equal(tt.expected) {
				t.Errorf("Decode() = %v, want %v", got, tt.expected)
			}
		})
	}
}
