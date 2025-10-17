package eventstream

import (
	"strings"
	"testing"
	"time"
)

// TestNewHandler tests the creation of a new handler
func TestNewHandler(t *testing.T) {
	h, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler() returned error: %v", err)
	}
	if h == nil {
		t.Fatal("NewHandler() returned nil handler")
	}
}

// TestHandler_AddReceiver tests adding receivers to a handler
func TestHandler_AddReceiver(t *testing.T) {
	h, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler() failed: %v", err)
	}

	ch := make(chan *Event, 10)
	err = h.AddReceiver(ch)
	if err != nil {
		t.Errorf("AddReceiver() failed: %v", err)
	}

	// Test adding multiple receivers
	ch2 := make(chan *Event, 10)
	err = h.AddReceiver(ch2)
	if err != nil {
		t.Errorf("AddReceiver() failed for second receiver: %v", err)
	}
}

// TestHandler_AddReceiver_Closed tests adding receivers to a closed handler
func TestHandler_AddReceiver_Closed(t *testing.T) {
	h, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler() failed: %v", err)
	}

	h.Close()

	ch := make(chan *Event, 10)
	err = h.AddReceiver(ch)
	if err == nil {
		t.Error("AddReceiver() should fail on closed handler")
	}
}

// TestHandler_RemoveReceiver tests removing receivers from a handler
func TestHandler_RemoveReceiver(t *testing.T) {
	h, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler() failed: %v", err)
	}

	ch1 := make(chan *Event, 10)
	ch2 := make(chan *Event, 10)

	h.AddReceiver(ch1)
	h.AddReceiver(ch2)

	// Remove first receiver
	h.RemoveReceiver(ch1)

	// Send event - only ch2 should receive
	h.Send("test")

	select {
	case <-ch1:
		t.Error("Removed receiver should not receive events")
	case <-time.After(50 * time.Millisecond):
		// Expected - ch1 should not receive
	}

	select {
	case event := <-ch2:
		if event.Data != "test" {
			t.Errorf("Expected 'test', got %v", event.Data)
		}
	case <-time.After(50 * time.Millisecond):
		t.Error("ch2 should have received the event")
	}
}

// TestHandler_Send tests sending events
func TestHandler_Send(t *testing.T) {
	h, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler() failed: %v", err)
	}

	ch := make(chan *Event, 10)
	h.AddReceiver(ch)

	testData := "test message"
	h.Send(testData)

	select {
	case event := <-ch:
		if event.Data != testData {
			t.Errorf("Expected %v, got %v", testData, event.Data)
		}
		if event.DataType != "string" {
			t.Errorf("Expected DataType 'string', got %v", event.DataType)
		}
		if event.Timestamp.IsZero() {
			t.Error("Timestamp should be set")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Event should have been received")
	}
}

// TestHandler_Send_MultipleReceivers tests sending to multiple receivers
func TestHandler_Send_MultipleReceivers(t *testing.T) {
	h, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler() failed: %v", err)
	}

	ch1 := make(chan *Event, 10)
	ch2 := make(chan *Event, 10)

	h.AddReceiver(ch1)
	h.AddReceiver(ch2)

	testData := "broadcast test"
	h.Send(testData)

	// Both channels should receive the event
	for i, ch := range []chan *Event{ch1, ch2} {
		select {
		case event := <-ch:
			if event.Data != testData {
				t.Errorf("Channel %d: Expected %v, got %v", i, testData, event.Data)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Channel %d should have received the event", i)
		}
	}
}

// TestHandler_Send_NilData tests sending nil data
func TestHandler_Send_NilData(t *testing.T) {
	h, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler() failed: %v", err)
	}

	ch := make(chan *Event, 10)
	h.AddReceiver(ch)

	h.Send(nil)

	select {
	case event := <-ch:
		if event.Data != nil {
			t.Errorf("Expected nil data, got %v", event.Data)
		}
		if event.DataType != "nil" {
			t.Errorf("Expected DataType 'nil', got %v", event.DataType)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Event should have been received")
	}
}

// TestHandler_Send_Closed tests sending to closed handler
func TestHandler_Send_Closed(t *testing.T) {
	h, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler() failed: %v", err)
	}

	ch := make(chan *Event, 10)
	h.AddReceiver(ch)
	h.Close()

	h.Send("test")

	// Should not receive anything since handler is closed
	select {
	case <-ch:
		t.Error("Closed handler should not send events")
	case <-time.After(50 * time.Millisecond):
		// Expected
	}
}

// TestHandler_Logf tests the Logf functionality
func TestHandler_Logf(t *testing.T) {
	h, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler() failed: %v", err)
	}

	ch := make(chan *Event, 10)
	h.AddReceiver(ch)

	h.Logf("test message %d", 123)

	select {
	case event := <-ch:
		syslog, ok := event.Data.(SysLogIsh)
		if !ok {
			t.Fatalf("Logf() event.Data is not SysLogIsh, got %T", event.Data)
		}
		if syslog.Level != Log {
			t.Errorf("Logf() syslog.Level = %v, want %v", syslog.Level, Log)
		}
		if syslog.Message != "test message 123" {
			t.Errorf("Logf() syslog.Message = %q, want %q", syslog.Message, "test message 123")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Logf() timeout waiting for event")
	}
}

// TestHandler_LogLevels tests different log levels
func TestHandler_LogLevels(t *testing.T) {
	h, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler() failed: %v", err)
	}

	ch := make(chan *Event, 10)
	h.AddReceiver(ch)

	tests := []struct {
		name     string
		logFunc  func(string, ...any)
		expected LogLevel
	}{
		{"Tracef", h.Tracef, TraceLevel},
		{"Debugf", h.Debugf, DebugLevel},
		{"Infof", h.Infof, InfoLevel},
		{"Warnf", h.Warnf, WarnLevel},
		{"Errorf", h.Errorf, ErrorLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.logFunc("test %s", tt.name)

			select {
			case event := <-ch:
				syslog, ok := event.Data.(SysLogIsh)
				if !ok {
					t.Fatalf("%s event.Data is not SysLogIsh, got %T", tt.name, event.Data)
				}
				if syslog.Level != tt.expected {
					t.Errorf("%s level = %v, want %v", tt.name, syslog.Level, tt.expected)
				}
				expectedMessage := "test " + tt.name
				if syslog.Message != expectedMessage {
					t.Errorf("%s message = %q, want %q", tt.name, syslog.Message, expectedMessage)
				}
			case <-time.After(100 * time.Millisecond):
				t.Fatalf("%s timeout waiting for event", tt.name)
			}
		})
	}
}

// TestHandler_Close tests closing a handler
func TestHandler_Close(t *testing.T) {
	h, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler() failed: %v", err)
	}

	ch := make(chan *Event, 10)
	h.AddReceiver(ch)

	h.Close()

	// Try to add receiver to closed handler
	ch2 := make(chan *Event, 10)
	err = h.AddReceiver(ch2)
	if err == nil {
		t.Error("AddReceiver() should fail on closed handler")
	}

	// Try to send to closed handler
	h.Send("test")
	select {
	case <-ch:
		t.Error("Closed handler should not send events")
	case <-time.After(50 * time.Millisecond):
		// Expected
	}
}

// TestEvent_CallerInfo tests caller information capture
func TestEvent_CallerInfo(t *testing.T) {
	h, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler() failed: %v", err)
	}

	ch := make(chan *Event, 10)
	h.AddReceiver(ch)

	h.Send("test caller info")

	select {
	case event := <-ch:
		// Should capture caller information
		if event.CallingFn == "" {
			t.Error("CallingFn should not be empty")
		}
		if event.File == "" {
			t.Error("File should not be empty")
		}
		if event.Line == 0 {
			t.Error("Line should not be zero")
		}

		// CallingFn should not contain the eventstream package (verify package detection works)
		if strings.Contains(event.CallingFn, "/pkg/eventstream") {
			t.Errorf("CallingFn should not contain eventstream package path, got: %s", event.CallingFn)
		}

		// Verify it's not pointing to our eventstream package files
		if strings.Contains(event.File, "/pkg/eventstream/") && !strings.HasSuffix(event.File, "_test.go") {
			t.Errorf("File should not point to eventstream package files, got: %s", event.File)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Should have received event")
	}
}

// TestEvent_CallerInfo_FromLogf tests caller info from log functions
func TestEvent_CallerInfo_FromLogf(t *testing.T) {
	h, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler() failed: %v", err)
	}

	ch := make(chan *Event, 10)
	h.AddReceiver(ch)

	h.Infof("test caller from logf")

	select {
	case event := <-ch:
		// Should capture caller, not the Infof method itself
		if event.CallingFn == "" {
			t.Error("CallingFn should not be empty")
		}
		// Should not contain the eventstream package path (verifies package detection)
		if strings.Contains(event.CallingFn, "/pkg/eventstream") {
			t.Errorf("CallingFn should not contain eventstream package path, got: %s", event.CallingFn)
		}
		// Verify it's not pointing to our eventstream package files (but test framework is OK)
		if strings.Contains(event.File, "/pkg/eventstream/") && !strings.HasSuffix(event.File, "_test.go") {
			t.Errorf("File should not point to eventstream package files, got: %s", event.File)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Should have received event")
	}
}

// TestLogLevel_String tests LogLevel string representation
func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		expected string
	}{
		{"log", Log, "LOG"},
		{"trace", TraceLevel, "TRACE"},
		{"debug", DebugLevel, "DEBUG"},
		{"info", InfoLevel, "INFO"},
		{"warn", WarnLevel, "WARN"},
		{"error", ErrorLevel, "ERROR"},
		{"unknown", LogLevel(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.level.String()
			if got != tt.expected {
				t.Errorf("LogLevel.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestSysLogIsh_String tests SysLogIsh string representation
func TestSysLogIsh_String(t *testing.T) {
	tests := []struct {
		name     string
		syslog   SysLogIsh
		contains []string
	}{
		{
			name:     "info message",
			syslog:   SysLogIsh{Level: InfoLevel, Message: "test message"},
			contains: []string{"INFO", "test message"},
		},
		{
			name:     "error message",
			syslog:   SysLogIsh{Level: ErrorLevel, Message: "error occurred"},
			contains: []string{"ERROR", "error occurred"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.syslog.String()
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("SysLogIsh.String() = %q should contain %q", got, want)
				}
			}
		})
	}
}

// TestDetectPackagePrefix tests that package prefix detection works
func TestDetectPackagePrefix(t *testing.T) {
	h, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler() failed: %v", err)
	}

	ch := make(chan *Event, 10)
	h.AddReceiver(ch)

	// Send from this test
	h.Send("test package detection")

	select {
	case event := <-ch:
		// Verify basic caller info is captured
		if event.CallingFn == "" {
			t.Error("CallingFn should not be empty")
		}
		if event.File == "" {
			t.Error("File should not be empty")
		}
		if event.Line == 0 {
			t.Error("Line should not be zero")
		}

		// The calling function should not contain the eventstream package path
		// This verifies that package prefix detection is working
		if strings.Contains(event.CallingFn, "/pkg/eventstream") {
			t.Errorf("Package prefix detection failed - CallingFn contains eventstream package path: %s", event.CallingFn)
		}

		// Verify it's not pointing to our eventstream package files (key test for package detection)
		if strings.Contains(event.File, "/pkg/eventstream/") && !strings.HasSuffix(event.File, "_test.go") {
			t.Errorf("Package prefix detection failed - pointing to eventstream package files: %s", event.File)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Should have received event")
	}
}

// Benchmark tests
func BenchmarkHandler_Send(b *testing.B) {
	h, err := NewHandler()
	if err != nil {
		b.Fatalf("NewHandler() failed: %v", err)
	}

	ch := make(chan *Event, 1000)
	h.AddReceiver(ch)

	// Consume events in background
	go func() {
		for range ch {
		}
	}()

	testData := "benchmark data"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h.Send(testData)
	}
}

func BenchmarkHandler_Logf(b *testing.B) {
	h, err := NewHandler()
	if err != nil {
		b.Fatalf("NewHandler() failed: %v", err)
	}

	ch := make(chan *Event, 1000)
	h.AddReceiver(ch)

	// Consume events in background
	go func() {
		for range ch {
		}
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h.Logf("benchmark log message %d", i)
	}
}

func BenchmarkHandler_CreateEvent(b *testing.B) {
	h, err := NewHandler()
	if err != nil {
		b.Fatalf("NewHandler() failed: %v", err)
	}

	testData := "benchmark test data"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = h.CreateEvent(testData)
	}
}
