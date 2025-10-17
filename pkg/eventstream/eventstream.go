// Package eventstream provides an event streaming and logging system.
// It allows sending events to multiple receivers via channels and includes
// structured logging with automatic caller information capture.
package eventstream

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Event represents an event with metadata about its origin.
type Event struct {
	Timestamp time.Time
	Data      any
	DataType  string
	CallingFn string
	File      string
	Line      int
}

// Handler manages event distribution to multiple receivers.
type Handler struct {
	mu            sync.RWMutex
	receivers     []chan *Event
	closed        bool
	packagePrefix string
}

// NewHandler creates a new Handler instance.
func NewHandler() (*Handler, error) {
	h := &Handler{
		receivers: make([]chan *Event, 0),
	}
	return h.detectPackagePrefix()
}
func (h *Handler) detectPackagePrefix() (*Handler, error) {
	// Detect our package name dynamically by looking at this function's name
	pc, _, _, ok := runtime.Caller(0)
	if ok {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			funcName := fn.Name()
			// Extract package path from function name
			// Format: github.com/nzions/fdot/pkg/eventstream.(*Handler).detectPackagePrefix
			// We want: github.com/nzions/fdot/pkg/eventstream
			if strings.Contains(funcName, ".(*Handler).") {
				// Handle method functions: remove .(*Handler).detectPackagePrefix suffix
				if methodIdx := strings.Index(funcName, ".(*Handler)."); methodIdx != -1 {
					h.packagePrefix = funcName[:methodIdx]
				}
			} else if lastDot := strings.LastIndexByte(funcName, '.'); lastDot != -1 {
				h.packagePrefix = funcName[:lastDot]
			}
		}
	}
	return h, nil
}

// AddReceiver adds a new receiver channel to the handler.
// Returns an error if the handler is closed.
func (h *Handler) AddReceiver(ch chan *Event) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return fmt.Errorf("handler is closed")
	}

	h.receivers = append(h.receivers, ch)
	return nil
}

// RemoveReceiver removes a receiver channel from the handler.
func (h *Handler) RemoveReceiver(ch chan *Event) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i, receiver := range h.receivers {
		if receiver == ch {
			h.receivers = append(h.receivers[:i], h.receivers[i+1:]...)
			return
		}
	}
}

// Send sends an event to all registered receivers.
// It automatically decorates the event with metadata (timestamp, data type, caller info).
func (h *Handler) Send(data any) {
	event := h.CreateEvent(data)
	h.sendEvent(event)
}

// SendEvent sends a pre-created event to all receivers.
// Use this with NewEvent when you want to modify an event before sending.
func (h *Handler) SendEvent(event *Event) {
	h.sendEvent(event)
}

// sendEvent sends a pre-created event to all receivers.
func (h *Handler) sendEvent(event *Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.closed {
		return
	}

	for _, receiver := range h.receivers {
		// Non-blocking send to prevent deadlocks
		select {
		case receiver <- event:
		default:
			// Receiver's channel is full, skip
		}
	}
}

// CreateEvent creates an Event with caller information.
// It walks the call stack to find the first caller outside the eventstream package.
func (h *Handler) CreateEvent(data any) *Event {
	event := &Event{
		Timestamp: time.Now(),
		Data:      data,
	}

	// Capture data type
	if data != nil {
		event.DataType = reflect.TypeOf(data).String()
	} else {
		event.DataType = "nil"
	}

	// Capture caller information by walking the stack
	// Find the first caller that is NOT in the eventstream package
	const maxDepth = 10 // Reduced from 25 - most calls are shallow

	// Pre-allocate for common case to reduce allocations
	var targetFile, targetFunc string
	var targetLine int

	for i := 1; i < maxDepth; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		funcName := fn.Name()

		// Fast path: check if function name starts with our detected package path
		// This is more precise and faster than Contains()
		if !strings.HasPrefix(funcName, h.packagePrefix) {
			// This is our caller - store and break immediately
			targetFile = file
			targetLine = line
			targetFunc = funcName
			break
		}

		// If it starts with our package prefix, check if it's exactly our package
		// (not a sub-package or similar named package)
		if len(funcName) > len(h.packagePrefix) {
			nextChar := funcName[len(h.packagePrefix)]
			if nextChar != '.' && nextChar != '/' {
				// Different package that happens to start with our prefix
				targetFile = file
				targetLine = line
				targetFunc = funcName
				break
			}
		}
	}

	// Assign captured values
	event.File = targetFile
	event.Line = targetLine
	event.CallingFn = targetFunc

	return event
}

// Close closes the handler and prevents new events from being sent.
func (h *Handler) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.closed = true
}
