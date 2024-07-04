package connection_test

import (
	"reflect"
	"testing"

	"github.com/Moonlight-Companies/goresp/connection"
	"github.com/Moonlight-Companies/goresp/resp"
)

func TestParseMessage(t *testing.T) {
	tests := []struct {
		name           string
		input          resp.RESPValue
		expectedMsg    *connection.BusMessage
		expectedReturn bool
	}{
		{
			name:           "Nil input",
			input:          nil,
			expectedMsg:    nil,
			expectedReturn: false,
		},
		{
			name:           "Simple String",
			input:          &resp.RESPSimpleString{Value: "OK"},
			expectedMsg:    nil,
			expectedReturn: false,
		},
		{
			name:           "Error",
			input:          &resp.RESPError{Value: "Error occurred"},
			expectedMsg:    nil,
			expectedReturn: false,
		},
		{
			name:           "Integer",
			input:          &resp.RESPInteger{Value: 42},
			expectedMsg:    nil,
			expectedReturn: false,
		},
		{
			name:           "Bulk String",
			input:          &resp.RESPBulkString{Value: []byte("test")},
			expectedMsg:    nil,
			expectedReturn: false,
		},
		{
			name: "Valid message",
			input: &resp.RESPArray{
				Items: []resp.RESPValue{
					&resp.RESPBulkString{Value: []byte("message")},
					&resp.RESPBulkString{Value: []byte("channel1")},
					&resp.RESPBulkString{Value: []byte("Hello, World!")},
				},
			},
			expectedMsg: &connection.BusMessage{
				Channel: "channel1",
				Data:    []byte("Hello, World!"),
			},
			expectedReturn: true,
		},
		{
			name: "Valid pmessage",
			input: &resp.RESPArray{
				Items: []resp.RESPValue{
					&resp.RESPBulkString{Value: []byte("pmessage")},
					&resp.RESPBulkString{Value: []byte("pattern1")},
					&resp.RESPBulkString{Value: []byte("channel1")},
					&resp.RESPBulkString{Value: []byte("Hello, Pattern!")},
				},
			},
			expectedMsg: &connection.BusMessage{
				Pattern: "pattern1",
				Channel: "channel1",
				Data:    []byte("Hello, Pattern!"),
			},
			expectedReturn: true,
		},
		{
			name: "Subscribe message",
			input: &resp.RESPArray{
				Items: []resp.RESPValue{
					&resp.RESPBulkString{Value: []byte("subscribe")},
					&resp.RESPBulkString{Value: []byte("channel1")},
					&resp.RESPInteger{Value: 1},
				},
			},
			expectedMsg:    nil,
			expectedReturn: false,
		},
		{
			name: "PSubscribe message",
			input: &resp.RESPArray{
				Items: []resp.RESPValue{
					&resp.RESPBulkString{Value: []byte("psubscribe")},
					&resp.RESPBulkString{Value: []byte("pattern1")},
					&resp.RESPInteger{Value: 1},
				},
			},
			expectedMsg:    nil,
			expectedReturn: false,
		},
		{
			name: "Invalid array length",
			input: &resp.RESPArray{
				Items: []resp.RESPValue{
					&resp.RESPBulkString{Value: []byte("message")},
					&resp.RESPBulkString{Value: []byte("channel1")},
				},
			},
			expectedMsg:    nil,
			expectedReturn: false,
		},
		{
			name: "Invalid message type",
			input: &resp.RESPArray{
				Items: []resp.RESPValue{
					&resp.RESPBulkString{Value: []byte("invalid")},
					&resp.RESPBulkString{Value: []byte("channel1")},
					&resp.RESPBulkString{Value: []byte("data")},
				},
			},
			expectedMsg:    nil,
			expectedReturn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, ok := connection.ParseMessage(tt.input)
			if ok != tt.expectedReturn {
				t.Errorf("ParseMessage() returned %v, want %v", ok, tt.expectedReturn)
			}
			if !reflect.DeepEqual(msg, tt.expectedMsg) {
				t.Errorf("ParseMessage() = %v, want %v", msg, tt.expectedMsg)
			}
		})
	}
}
