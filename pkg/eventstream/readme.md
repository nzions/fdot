# eventstream - Event Streaming and Logging Package

A Go package providing event streaming with pub/sub pattern and structured logging with automatic caller information capture.

## Features

- **Event streaming with metadata**: Automatic capture of timestamp, data type, caller function, file, and line number
- **Pub/Sub pattern**: Multiple receivers can subscribe to events from a single handler
- **Structured logging**: Built-in log levels (Trace, Debug, Info, Warn, Error)
- **Concurrent-safe**: Thread-safe handler operations with mutex protection
- **Default handler**: Global handler that prints to STDOUT/STDERR for convenience
- **Non-blocking sends**: Prevents deadlocks with full receiver channels

## Installation

```bash
go get github.com/nzions/fdot/pkg/eventstream
```

## Quick Start

```go
import "github.com/nzions/fdot/pkg/eventstream"

// Use global convenience functions
eventstream.Infof("Application started")
eventstream.Debugf("Processing %d items", count)
eventstream.Errorf("Failed to connect: %v", err)

// Send custom data
eventstream.Send(map[string]any{"status": "ok", "count": 100})
```

## API Reference

### Types

#### Event
Represents an event with metadata about its origin.

```go
type Event struct {
    Timestamp  time.Time  // When the event was created
    Data       any        // The event payload
    DataType   string     // Reflected type name of Data
    CallingFn  string     // Name of the calling function
    File       string     // Source file path
    Line       int        // Line number in source file
}
```

#### SysLog
Represents a structured log message.

```go
type SysLog struct {
    Level   LogLevel  // Severity level
    Message string    // Log message
}
```

#### LogLevel
Log severity levels:
- `TraceLevel` - Very detailed debug information
- `DebugLevel` - Debug information
- `InfoLevel` - Informational messages
- `WarnLevel` - Warning messages
- `ErrorLevel` - Error messages

#### Handler
Manages event distribution to multiple receivers.

```go
type Handler struct {
    // ... (internal fields)
}
```

### Functions

#### Handler Creation

```go
func NewHandler() *Handler
```
Creates a new Handler instance.

#### Handler Methods

```go
func (h *Handler) AddReceiver(ch chan *Event) error
```
Adds a receiver channel. Returns error if handler is closed.

```go
func (h *Handler) RemoveReceiver(ch chan *Event)
```
Removes a receiver channel from the handler.

```go
func (h *Handler) Send(data any)
```
Sends an event to all registered receivers with automatic metadata decoration.

```go
func (h *Handler) Logf(level LogLevel, format string, args ...any)
```
Sends a formatted log message with the specified level.

```go
func (h *Handler) Tracef(format string, args ...any)
func (h *Handler) Debugf(format string, args ...any)
func (h *Handler) Infof(format string, args ...any)
func (h *Handler) Warnf(format string, args ...any)
func (h *Handler) Errorf(format string, args ...any)
```
Convenience methods for sending log messages at specific levels.

```go
func (h *Handler) Close()
```
Closes the handler and prevents new events from being sent.

#### Global Functions

Convenience functions that use the DefaultHandler:

```go
func Send(data any)
func Tracef(format string, args ...any)
func Debugf(format string, args ...any)
func Infof(format string, args ...any)
func Warnf(format string, args ...any)
func Errorf(format string, args ...any)
```

### DefaultHandler

```go
var DefaultHandler *Handler
```
Global handler that prints events to STDOUT/STDERR. Errors go to STDERR, all other messages to STDOUT.

## Usage Examples

### Basic Logging

```go
import "github.com/nzions/fdot/pkg/eventstream"

eventstream.Infof("Server started on port %d", 8080)
eventstream.Debugf("Loaded %d configuration items", len(config))
eventstream.Warnf("Cache miss rate: %.2f%%", missRate)
eventstream.Errorf("Database connection failed: %v", err)
```

### Custom Handler with Receivers

```go
handler := eventstream.NewHandler()

// Create a receiver channel
receiver := make(chan *eventstream.Event, 100)
handler.AddReceiver(receiver)

// Send events
go func() {
    handler.Infof("Processing started")
    handler.Send(customData)
}()

// Process events
for event := range receiver {
    if syslog, ok := event.Data.(eventstream.SysLog); ok {
        fmt.Printf("[%s] %s at %s:%d\n", 
            syslog.Level, syslog.Message, event.File, event.Line)
    }
}
```

### Pub/Sub Pattern (Multiple Receivers)

```go
handler := eventstream.NewHandler()

// Multiple subscribers
receiver1 := make(chan *eventstream.Event, 10)
receiver2 := make(chan *eventstream.Event, 10)

handler.AddReceiver(receiver1)
handler.AddReceiver(receiver2)

// Broadcast to all receivers
handler.Infof("Broadcast message")

// Both receivers get the same event
event1 := <-receiver1
event2 := <-receiver2
```

### Event Metadata Inspection

```go
handler := eventstream.NewHandler()
receiver := make(chan *eventstream.Event, 10)
handler.AddReceiver(receiver)

handler.Send(myData)

event := <-receiver
fmt.Printf("Timestamp: %s\n", event.Timestamp)
fmt.Printf("Type: %s\n", event.DataType)
fmt.Printf("Called from: %s at %s:%d\n", 
    event.CallingFn, event.File, event.Line)
```

## Example Application

See `examples/eventstream/eventstream_demo.go` for a comprehensive example demonstrating:
1. Global convenience functions
2. Custom handlers with receivers
3. Multiple receivers (pub/sub pattern)
4. Event metadata inspection
5. All log levels

Run the example:
```bash
cd examples/eventstream
go run eventstream_demo.go
```

## Testing

Run the test suite:
```bash
cd pkg/eventstream
go test -v
```

Run benchmarks:
```bash
go test -bench=. -benchmem
```

## Thread Safety

All Handler operations are thread-safe and can be called concurrently from multiple goroutines. The implementation uses mutex locks to protect internal state.

## Performance Considerations

- Event sends are non-blocking - if a receiver's channel is full, the send is skipped
- Use buffered channels for receivers to prevent blocking
- The DefaultHandler uses a background goroutine to process events asynchronously
- Caller information capture uses `runtime.Caller()` which has minimal overhead

## License

Part of the fdot project.