package connection

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/Moonlight-Companies/goresp/command"
	"github.com/Moonlight-Companies/goresp/logging"
	"github.com/Moonlight-Companies/goresp/resp"
)

const (
	healthCheckInterval = 5 * time.Second
	maxReconnectDelay   = 30 * time.Second
)

type BusMessage struct {
	Channel string
	Data    []byte
	Pattern string
}

func (m *BusMessage) IntoMap() (output map[string]interface{}, err error) {
	if err = json.Unmarshal(m.Data, &output); err != nil {
		return nil, err
	}
	return output, nil
}

type ReconnectingChannel struct {
	Channel string
	Kind    string
}

type Reconnecting struct {
	logger         *logging.Logger
	addr           string
	conn           net.Conn
	decoder        *resp.Decode
	lastData       time.Time
	connected      bool
	mutex          sync.Mutex
	done           chan struct{}
	data           chan []byte
	commands       chan []byte
	reconnectDelay time.Duration
	channels       sync.Map
	Messages       chan BusMessage
}

func NewReconnecting(addr string) *Reconnecting {
	result := &Reconnecting{
		logger:         logging.NewLogger(logging.LogLevelInfo),
		addr:           addr,
		decoder:        &resp.Decode{},
		done:           make(chan struct{}),
		reconnectDelay: time.Second,
		data:           make(chan []byte, 255),
		commands:       make(chan []byte, 255),
		Messages:       make(chan BusMessage, 255),
	}

	go result.handleReconnect()
	go result.handleHealthCheck()
	go result.handleData()
	go result.handleSend()

	return result
}

func (r *Reconnecting) Close() {
	close(r.done)
	r.disconnect()
}

func (r *Reconnecting) onConnect() {
	r.Send(command.FormatCommand("PING"))

	r.channels.Range(func(key, value interface{}) bool {
		channelItem := value.(ReconnectingChannel)
		switch channelItem.Kind {
		case "SUBSCRIBE":
			r.subscribe(channelItem.Channel)
		case "PSUBSCRIBE":
			r.psubscribe(channelItem.Channel)
		}
		return true
	})
}

func (r *Reconnecting) onDisconnect() {
	r.logger.Info("Disconnected from Redis")
}

func (r *Reconnecting) Subscribe(channels ...string) {
	for _, channel := range channels {
		if _, loaded := r.channels.LoadOrStore(channel, ReconnectingChannel{Channel: channel, Kind: "SUBSCRIBE"}); !loaded {
			r.subscribe(channel)
		}
	}
}

func (r *Reconnecting) subscribe(channel string) {
	cmd := command.FormatCommand("SUBSCRIBE", channel)
	r.Send(cmd)
}

func (r *Reconnecting) PSubscribe(patterns ...string) {
	for _, pattern := range patterns {
		if _, loaded := r.channels.LoadOrStore(pattern, ReconnectingChannel{Channel: pattern, Kind: "PSUBSCRIBE"}); !loaded {
			r.psubscribe(pattern)
		}
	}
}

func (r *Reconnecting) psubscribe(pattern string) {
	cmd := command.FormatCommand("PSUBSCRIBE", pattern)
	r.Send(cmd)
}

func (r *Reconnecting) Unsubscribe(channels ...string) {
	for _, channel := range channels {
		if _, loaded := r.channels.LoadAndDelete(channel); loaded {
			r.unsubscribe(channel)
		}
	}
}

func (r *Reconnecting) unsubscribe(channel string) {
	cmd := command.FormatCommand("UNSUBSCRIBE", channel)
	r.Send(cmd)
}

func (r *Reconnecting) PUnsubscribe(patterns ...string) {
	for _, pattern := range patterns {
		if _, loaded := r.channels.LoadAndDelete(pattern); loaded {
			r.punsubscribe(pattern)
		}
	}
}

func (r *Reconnecting) punsubscribe(pattern string) {
	cmd := command.FormatCommand("PUNSUBSCRIBE", pattern)
	r.Send(cmd)
}

func (r *Reconnecting) Send(cmd []byte) {
	select {
	case r.commands <- cmd:
		r.logger.Debug("Sent command: %s", cmd)
	default:
		r.logger.Warn("Command queue full, dropping command: %s", cmd)
	}
}

func (r *Reconnecting) handleReconnect() {
	for {
		select {
		case <-r.done:
			return
		default:
			if !r.isConnected() {
				if err := r.connect_and_produce_data(); err != nil {
					r.logger.Error("Failed to connect: %v", err)
					time.Sleep(r.reconnectDelay)
					r.reconnectDelay = min(r.reconnectDelay*2, maxReconnectDelay)
				} else {
					r.reconnectDelay = time.Second
				}
			}
			time.Sleep(time.Second)
		}
	}
}

func (r *Reconnecting) handleHealthCheck() {
	ticker := time.NewTicker(healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.done:
			return
		case <-ticker.C:
			r.healthCheck()
		}
	}
}

func (r *Reconnecting) healthCheck() {
	if !r.isConnected() {
		return
	}

	if time.Since(r.lastData) > 4*healthCheckInterval {
		r.logger.Warn("No data received for a while, disconnecting")
		r.disconnect()
		return
	}

	if time.Since(r.lastData) > healthCheckInterval {
		randomString := fmt.Sprintf("%d", rand.Int())
		pingCmd := command.FormatCommand("PING", randomString)
		r.Send(pingCmd)
	}
}

func (r *Reconnecting) handleSend() {
	for cmd := range r.commands {
		if r.isConnected() {
			_, err := r.conn.Write([]byte(cmd))
			if err != nil {
				r.logger.Error("Failed to send command: %v", err)
				r.disconnect()
			}
		}
	}
}

func (r *Reconnecting) connect_and_produce_data() error {
	conn, err := net.DialTimeout("tcp", r.addr, 10*time.Second)
	if err != nil {
		return err
	}

	r.mutex.Lock()
	r.conn = conn
	r.connected = true
	r.lastData = time.Now()
	r.decoder.Reset()
	r.mutex.Unlock()

	defer func() {
		r.conn.Close()
		r.conn = nil
		r.connected = false
		r.onDisconnect()
	}()

	r.logger.Info("Connected to Redis")
	r.onConnect()

	for {
		buffer := make([]byte, 16384)
		n, err := r.conn.Read(buffer)
		if err != nil {
			r.logger.Error("Read failed: %v", err)
			return err
		}

		r.lastData = time.Now()
		r.logger.Debug("RECEIVED %s", string(buffer[:n]))

		select {
		case r.data <- buffer[:n]:
		default:
			r.logger.Warn("Data queue full, aborting connection")
			return nil
		}
	}
}

func (r *Reconnecting) disconnect() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if c := r.conn; c != nil {
		c.Close()
		r.decoder.Reset()
	}
}

func (r *Reconnecting) isConnected() bool {
	return r.conn != nil && r.connected
}

func (r *Reconnecting) handleData() {
	for data := range r.data {
		if r.isConnected() {
			r.decoder.Provide(data)
			r.parse()
		} else {
			r.decoder.Reset()
		}
	}
}

func (r *Reconnecting) parse() error {
	for {
		value, err := r.decoder.Parse()
		if err != nil {
			r.logger.Error("Error parsing data: %v", err)
			r.disconnect()
			return err
		}

		if value == nil {
			return nil
		}

		array, ok := value.(*resp.RESPArray)
		if !ok {
			r.logger.Debug("Received non-array value: %s, ignoring %v", value.Type(), value)
			continue
		}

		if len(array.Items) > 0 {
			messageType, ok := array.Items[0].(*resp.RESPBulkString)
			if !ok {
				continue
			}

			switch messageType.String() {
			case "pong":
				continue
			}
		}

		if len(array.Items) < 3 {
			continue
		}

		messageType, ok := array.Items[0].(*resp.RESPBulkString)
		if !ok {
			r.logger.Warn("Invalid message type, ignoring")
			continue
		}

		var busMessage BusMessage

		switch messageType.String() {
		case "message":
			if len(array.Items) != 3 {
				r.logger.Warn("Invalid MESSAGE format, ignoring")
				continue
			}
			channel, ok := array.Items[1].(*resp.RESPBulkString)
			if !ok {
				r.logger.Warn("Invalid channel format, ignoring")
				continue
			}
			data, ok := array.Items[2].(*resp.RESPBulkString)
			if !ok {
				r.logger.Warn("Invalid data format, ignoring")
				continue
			}
			busMessage = BusMessage{
				Channel: channel.String(),
				Data:    []byte(data.String()),
			}
		case "pmessage":
			if len(array.Items) != 4 {
				r.logger.Warn("Invalid PMESSAGE format, ignoring")
				continue
			}
			pattern, ok := array.Items[1].(*resp.RESPBulkString)
			if !ok {
				r.logger.Warn("Invalid pattern format, ignoring")
				continue
			}
			channel, ok := array.Items[2].(*resp.RESPBulkString)
			if !ok {
				r.logger.Warn("Invalid channel format, ignoring")
				continue
			}
			data, ok := array.Items[3].(*resp.RESPBulkString)
			if !ok {
				r.logger.Warn("Invalid data format, ignoring")
				continue
			}
			busMessage = BusMessage{
				Channel: channel.String(),
				Data:    []byte(data.String()),
				Pattern: pattern.String(),
			}
		default:
			continue
		}

		select {
		case r.Messages <- busMessage:
		default:
			r.logger.Warn("Producer queue full, dropping message. Queue length: %d", len(r.Messages))
		}
	}
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
