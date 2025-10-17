// Example usage of the eventstream package
package main

import (
	"fmt"
	"time"

	"github.com/nzions/fdot/pkg/eventstream"
)

func main() {
	fmt.Println("=== eventstream Package Demo ===")
	fmt.Println()

	// 1. Creating a handler for logging
	fmt.Println("1. Using handler for logging:")
	handler, err := eventstream.NewHandler()
	if err != nil {
		fmt.Printf("Failed to create handler: %v\n", err)
		return
	}

	handler.Infof("Application started at %s", time.Now().Format(time.RFC3339))
	handler.Debugf("Debug information: processing %d items", 42)
	handler.Warnf("Warning: resource usage at %.1f%%", 85.3)
	handler.Errorf("Error: failed to connect to %s", "database")
	handler.Send(map[string]any{"status": "ok", "count": 100})
	time.Sleep(50 * time.Millisecond) // Let messages process

	// 2. Creating custom handler with receivers
	fmt.Println("\n2. Creating custom handler with receivers:")
	customHandler, err := eventstream.NewHandler()
	if err != nil {
		fmt.Printf("Failed to create custom handler: %v\n", err)
		return
	}

	// Create a receiver that collects events
	receiver := make(chan *eventstream.Event, 100)
	if err := customHandler.AddReceiver(receiver); err != nil {
		fmt.Printf("Error adding receiver: %v\n", err)
		return
	}

	// Send events in a goroutine
	go func() {
		customHandler.Infof("Custom handler message 1")
		customHandler.Debugf("Custom handler message 2")
		customHandler.Send("Custom data event")
	}()

	// Receive and process events
	timeout := time.After(100 * time.Millisecond)
	eventCount := 0
	for {
		select {
		case event := <-receiver:
			eventCount++
			if syslog, ok := event.Data.(eventstream.SysLogIsh); ok {
				fmt.Printf("   [%s] %s (from %s:%d)\n",
					syslog.Level,
					syslog.Message,
					event.File,
					event.Line)
			} else {
				fmt.Printf("   [Event] Type=%s Data=%v (from %s:%d)\n",
					event.DataType,
					event.Data,
					event.File,
					event.Line)
			}
		case <-timeout:
			goto done
		}
	}
done:
	fmt.Printf("   Received %d events\n", eventCount)

	// 3. Multiple receivers example
	fmt.Println("\n3. Multiple receivers (pub/sub pattern):")
	pubsub, err := eventstream.NewHandler()
	if err != nil {
		fmt.Printf("Failed to create handler: %v\n", err)
		return
	}

	// Create multiple receivers
	receiver1 := make(chan *eventstream.Event, 10)
	receiver2 := make(chan *eventstream.Event, 10)

	pubsub.AddReceiver(receiver1)
	pubsub.AddReceiver(receiver2)

	// Send an event
	pubsub.Infof("Broadcast message to all subscribers")

	// Both receivers get the same event
	event1 := <-receiver1
	event2 := <-receiver2

	if syslog1, ok := event1.Data.(eventstream.SysLogIsh); ok {
		if syslog2, ok := event2.Data.(eventstream.SysLogIsh); ok {
			fmt.Printf("   Receiver 1: %s\n", syslog1.Message)
			fmt.Printf("   Receiver 2: %s\n", syslog2.Message)
		}
	}

	// 4. Event metadata inspection
	fmt.Println("\n4. Event metadata inspection:")
	metaHandler, err := eventstream.NewHandler()
	if err != nil {
		fmt.Printf("Failed to create handler: %v\n", err)
		return
	}
	metaReceiver := make(chan *eventstream.Event, 10)
	metaHandler.AddReceiver(metaReceiver)

	metaHandler.Send(struct {
		Name  string
		Value int
	}{Name: "test", Value: 123})

	metaEvent := <-metaReceiver
	fmt.Printf("   Timestamp:  %s\n", metaEvent.Timestamp.Format("15:04:05.000"))
	fmt.Printf("   DataType:   %s\n", metaEvent.DataType)
	fmt.Printf("   CallingFn:  %s\n", metaEvent.CallingFn)
	fmt.Printf("   File:       %s\n", metaEvent.File)
	fmt.Printf("   Line:       %d\n", metaEvent.Line)
	fmt.Printf("   Data:       %+v\n", metaEvent.Data)

	// 5. Log levels demonstration
	fmt.Println("\n5. All log levels:")
	levelHandler, err := eventstream.NewHandler()
	if err != nil {
		fmt.Printf("Failed to create handler: %v\n", err)
		return
	}
	levelReceiver := make(chan *eventstream.Event, 10)
	levelHandler.AddReceiver(levelReceiver)

	levels := []struct {
		name string
		fn   func(string, ...any)
	}{
		{"Trace", levelHandler.Tracef},
		{"Debug", levelHandler.Debugf},
		{"Info", levelHandler.Infof},
		{"Warn", levelHandler.Warnf},
		{"Error", levelHandler.Errorf},
	}

	for _, level := range levels {
		level.fn("This is a %s message", level.name)
		event := <-levelReceiver
		if syslog, ok := event.Data.(eventstream.SysLogIsh); ok {
			fmt.Printf("   %-5s: %s\n", syslog.Level, syslog.Message)
		}
	}

	fmt.Println("\n=== Demo Complete ===")
}
