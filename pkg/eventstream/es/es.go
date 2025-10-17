package es

import (
	"fmt"
	"os"

	"github.com/nzions/fdot/pkg/eventstream"
)

// DefaultHandler is a global handler that prints events to STDOUT/STDERR.
var DefaultHandler *eventstream.Handler

func init() {
	dh, err := eventstream.NewHandler()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize DefaultHandler: %v", err))
	}
	DefaultHandler = dh

	// Create a receiver that prints to stdout/stderr
	receiver := make(chan *eventstream.Event, 100)
	err = DefaultHandler.AddReceiver(receiver)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize DefaultHandler: %v", err))
	}

	// Start goroutine to process events
	go func() {
		for event := range receiver {
			// Check if it's a SysLog
			if syslog, ok := event.Data.(eventstream.SysLogIsh); ok {
				output := os.Stdout
				if syslog.Level >= eventstream.ErrorLevel {
					output = os.Stderr
				}
				fmt.Fprintf(output, "%s [%s:%d] %s\n",
					event.Timestamp.Format("2006-01-02 15:04:05"),
					event.File,
					event.Line,
					syslog.String())
			} else {
				// Print generic event data
				fmt.Fprintf(os.Stdout, "%s [%s:%d] Type=%s Data=%v\n",
					event.Timestamp.Format("2006-01-02 15:04:05"),
					event.File,
					event.Line,
					event.DataType,
					event.Data)
			}
		}
	}()
}

// Convenience functions that use DefaultHandler

// Send sends an event using the DefaultHandler.
func Send(data any) {
	DefaultHandler.Send(data)
}

// CreateEvent creates an Event using the DefaultHandler without sending it.
// This allows users to modify the event before sending it with SendEvent.
func CreateEvent(data any) *eventstream.Event {
	return DefaultHandler.CreateEvent(data)
}

// SendEvent sends a pre-created event using the DefaultHandler.
func SendEvent(event *eventstream.Event) {
	DefaultHandler.SendEvent(event)
}

// Tracef sends a trace-level log message using the DefaultHandler.
func Tracef(format string, args ...any) {
	DefaultHandler.Tracef(format, args...)
}

// Debugf sends a debug-level log message using the DefaultHandler.
func Debugf(format string, args ...any) {
	DefaultHandler.Debugf(format, args...)
}

// Infof sends an info-level log message using the DefaultHandler.
func Infof(format string, args ...any) {
	DefaultHandler.Infof(format, args...)
}

// Warnf sends a warning-level log message using the DefaultHandler.
func Warnf(format string, args ...any) {
	DefaultHandler.Warnf(format, args...)
}

// Errorf sends an error-level log message using the DefaultHandler.
func Errorf(format string, args ...any) {
	DefaultHandler.Errorf(format, args...)
}
