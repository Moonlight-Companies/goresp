package connection

import (
	"github.com/Moonlight-Companies/goresp/resp"
)

func ParseMessage(value resp.RESPValue) (*BusMessage, bool) {
	if value == nil {
		return nil, false // No value to parse
	}

	array, ok := value.(*resp.RESPArray)
	if !ok {
		return nil, false
	}

	if len(array.Items) < 3 {
		return nil, false // Not enough data to form a message
	}

	messageType, ok := array.Items[0].(*resp.RESPBulkString)
	if !ok || (messageType.String() != "message" && messageType.String() != "pmessage") {
		return nil, false // Not a message or pmessage
	}

	busMessage := BusMessage{
		Channel: "",
		Pattern: "",
	}

	switch messageType.String() {
	case "message":
		if len(array.Items) != 3 {
			return nil, false // Incorrect format for message
		}
		channel, ok := array.Items[1].(*resp.RESPBulkString)
		if !ok {
			return nil, false
		}
		data, ok := array.Items[2].(*resp.RESPBulkString)
		if !ok {
			return nil, false
		}
		busMessage.Channel = channel.String()
		busMessage.Data = []byte(data.String())

	case "pmessage":
		if len(array.Items) != 4 {
			return nil, false // Incorrect format for pmessage
		}
		pattern, ok := array.Items[1].(*resp.RESPBulkString)
		if !ok {
			return nil, false
		}
		channel, ok := array.Items[2].(*resp.RESPBulkString)
		if !ok {
			return nil, false
		}
		data, ok := array.Items[3].(*resp.RESPBulkString)
		if !ok {
			return nil, false
		}
		busMessage.Pattern = pattern.String()
		busMessage.Channel = channel.String()
		busMessage.Data = []byte(data.String())
	default:
		return nil, false
	}

	return &busMessage, true
}
