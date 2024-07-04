# GoRESP: RESP (Redis Serialization Protocol) Implementation in Go

GoRESP is an implementation of the Redis Serialization Protocol (RESP) in Go, with a focus on providing a non-blocking stream of events from Redis pubsub channels. It offers a solution for encoding and decoding RESP data streams, along with a PubSub connector featuring auto-reconnection capabilities.

## Features

- RESP protocol support (Simple Strings, Errors, Integers, Bulk Strings, Arrays)
- Streaming decoder for processing data streams
- Error handling and recovery
- PubSub connector with automatic reconnection and resubscription
- Health checks to handle silent stream stoppages (no syscall errors)
- Non-blocking event stream for consuming pubsub messages

## Components

### RESP Decoder

The core of GoRESP is a streaming RESP decoder capable of handling both complete messages and partial streams:

- Processes data from byte-by-byte to large chunk partial messages
- Supports all RESP data types
- Handles malformed input
- Resumes parsing from incomplete data

### PubSub Connector (`NewReconnecting`)

Provides a high-level interface for subscribing to Redis channels:

- Automatic connection management
- Resubscription to channels after reconnection
- Configurable health checks for detecting message inactivity
- Non-blocking channel for receiving published messages, ignoring non-pubsub messages

## Usage

### RESP Decoding with Decoder-Managed Buffer

```go
decoder := resp.NewDecode()
decoder.Provide([]byte("+OK\r\n"))
value, err := decoder.Parse()
if err != nil {
    // Handle error
}
if value == nil {
    // Provide additional data
}
// Use value
```

### Direct RESP Decoding

```go
value, consumed, err := resp.DecodeValue([]byte("+OK\r\n"), 0)
if err != nil {
    // Handle error
}
if value == nil {
    // Provide additional data
}
// Use value
// Remember to consume 'consumed' bytes from the start of your buffer `buf.Next(consumed)`
```

### Encoding

```go
respSimpleString := resp.RESPSimpleString{Value: "hello world"}
buf := bytes.Buffer{}
respSimpleString.Encode(&buf) // encode appending to the buf
```

### Formatting Redis Commands

```go
// Format a simple SUBSCRIBE command
subCmd := resp.FormatCommand("SUBSCRIBE", "chan")
fmt.Printf("Encoded SUBSCRIBE command: %q\n", subCmd)
```

### PubSub Connector (Non-blocking Event Stream)

```go
reconn := resp.NewReconnecting("127.0.0.1:6379")
reconn.Subscribe("chan")

go func() {
    for msg := range reconn.Messages {
        fmt.Println(msg.Channel, msg.Pattern, len(msg.Data))
        temp, err := msg.IntoMap()
        // Process message asynchronously
    }
}()

// Your main application logic continues here...
```

## Customization

### Custom Connection Implementation

Implement your own connection by calling `.Provide([]byte)` on the decoder:

```go
d := resp.Decode{}
// ...
d.Provide(buf)
for {
    item, err := d.Parse()
    if err != nil {
        break // unrecoverable
    }
    if item == nil {
        break // needs more data
    }
    handle(item) // a RESPValue
}
```

## Testing

The test suite includes:

- Unit tests for RESP data types
- Streaming decode tests for messages broken across multiple reads
- Error handling and recovery tests
- Permutation tests for protocol compliance checking
- Failure tests for unrecoverable conditions

Run tests with:

```
go test ./...
```

## Non-standard Testing

Use `socat` to simulate various network conditions:

```bash
socat -v -x TCP4-LISTEN:6379,fork,reuseaddr TCP4:bus:6379
```

Manipulate the socat child process to test connection handling:

```bash
kill -SIGSTOP <pid of socat child>  # Pause data flow
kill -SIGCONT <pid of socat child>  # Resume data flow
kill -SIGKILL <pid of socat child>  # Force a reconnection
```

This allows testing of health check mechanisms and ensuring we discard partial state on reconnect.

## Contributing

Contributions are welcome! Please submit a Pull Request or open an Issue for discussion.