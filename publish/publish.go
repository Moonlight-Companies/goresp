package publish

import (
	"encoding/json"
	"errors"

	"github.com/Moonlight-Companies/goresp/command"
	"github.com/Moonlight-Companies/goresp/connection"
)

var ErrorQueueFull = errors.New("queue full")

func Publish(channel string, message map[string]interface{}) error {
	json, err := json.Marshal(message)
	if err != nil {
		return err
	}

	select {
	case publish_message_queue <- publishCommand{Channel: channel, Message: json}:
		return nil
	default:
		return ErrorQueueFull
	}
}

func PublishWithEvent(channel string, event string, message map[string]interface{}) error {
	clone := make(map[string]interface{}, len(message)+1)
	for k, v := range message {
		clone[k] = v
	}
	clone["Event"] = event
	return Publish(channel, clone)
}

type publishCommand struct {
	Channel string
	Message []byte
}

var conn *connection.Reconnecting = connection.NewReconnecting("bus:6379")
var publish_message_queue = make(chan publishCommand, 1000)

func init() {
	go func() {
		for msg := range publish_message_queue {
			payload := command.FormatCommand("publish", msg.Channel, string(msg.Message))
			conn.Send(payload)
		}
	}()
}
