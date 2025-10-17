// Example demonstrating custom event creation and modification
package main

import (
	"fmt"
	"time"

	"github.com/nzions/fdot/pkg/eventstream"
)

func main() {
	fmt.Println("=== Custom Event Manipulation Demo ===")

	// Create a handler for this demo
	handler, err := eventstream.NewHandler()
	if err != nil {
		fmt.Printf("Failed to create handler: %v\n", err)
		return
	}

	// Create a receiver to capture events
	receiver := make(chan *eventstream.Event, 100)
	if err := handler.AddReceiver(receiver); err != nil {
		fmt.Printf("Failed to add receiver: %v\n", err)
		return
	}

	// Method 1: Simple usage
	fmt.Println("1. Simple event sending:")
	handler.Send("Hello from simple API")
	time.Sleep(100 * time.Millisecond) // Let output appear

	// Method 2: Create event and modify before sending
	fmt.Println("\n2. Create and modify event before sending:")
	event := handler.CreateEvent("Custom data")

	// User can modify the event
	event.Data = map[string]any{
		"original":           "Custom data",
		"modified":           "I added this!",
		"timestamp_modified": time.Now(),
	}

	// User can even override caller information if needed
	event.CallingFn = "main.customFunction"
	event.File = "/custom/path/example.go"
	event.Line = 42

	handler.SendEvent(event)
	time.Sleep(100 * time.Millisecond)

	// Method 3: Create multiple events and batch send
	fmt.Println("\n3. Batch event creation:")
	events := make([]*eventstream.Event, 3)
	for i := 0; i < 3; i++ {
		events[i] = handler.CreateEvent(fmt.Sprintf("Batch event %d", i+1))
		// Add custom metadata
		events[i].Data = map[string]any{
			"batch_id": "batch-001",
			"sequence": i + 1,
			"data":     events[i].Data,
		}
	}

	// Send all at once
	for _, event := range events {
		handler.SendEvent(event)
	}

	time.Sleep(100 * time.Millisecond)
	fmt.Println("\nDemo complete!")
}
