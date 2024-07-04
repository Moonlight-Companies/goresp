# GoRESP: Robust RESP (Redis Serialization Protocol) Implementation in Go

GoRESP is a comprehensive implementation of the Redis Serialization Protocol (RESP) in Go. It offers a flexible and robust solution for encoding and decoding RESP data streams, along with a PubSub connector featuring auto-reconnection capabilities.

## Features

- Full RESP protocol support (Simple Strings, Errors, Integers, Bulk Strings, Arrays)
- Efficient streaming decoder for processing large data streams
- Comprehensive error handling and recovery
- PubSub connector with automatic reconnection and resubscription
- Health checks to handle silent stream stoppages (no syscall errors)
- Extensible design allowing custom connection implementations for encoding and decoding

## Components

### RESP Decoder

The core of GoRESP is a high-performance RESP decoder capable of handling both complete messages and partial streams:

- Processes data from byte-by-byte to large chunk partial messages
- Supports all RESP data types
- Robust error handling for malformed input
- Resumes parsing from incomplete data

### PubSub Connector (`NewReconnecting`)

Provides a high-level interface for subscribing to Redis channels:

- Automatic connection management
- Resubscription to channels after reconnection
- Configurable health checks for detecting message inactivity
- Channel for receiving published messages in a usable format, ignoring non-pubsub messages like ping responses

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

## Formatting Redis Commands

```go
// Format a simple SUBSCRIBE command
subCmd := resp.FormatCommand("SUBSCRIBE", "chan")
fmt.Printf("Encoded SUBSCRIBE command: %q\n", subCmd)
```

### PubSub Connector

```go
reconn := resp.NewReconnecting("127.0.0.1:6379")
reconn.Subscribe("chan")

for msg := range reconn.Messages {
    fmt.Println(msg.Channel, msg.Pattern, len(msg.Data))
    temp, err := msg.IntoMap()
}
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

## Performance

GoRESP is optimized for performance, processing data byte-by-byte to minimize memory allocations and enable efficient streaming of large datasets.

## Testing

The comprehensive test suite includes:

- Unit tests for all RESP data types
- Streaming decode tests for large messages broken across multiple reads
- Error handling and recovery tests
- Permutation tests for thorough protocol compliance checking
- Failure tests for unrecoverable conditions:
  1. Failed integer parsing where expected
  2. Invalid opcode

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

## License

[MIT License](LICENSE)