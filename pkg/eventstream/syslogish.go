package eventstream

import "fmt"

// LogLevel represents the severity level of a log message.
type LogLevel int

const (
	// Log for general-purpose logging (used by Logf)
	Log LogLevel = iota
	// TraceLevel for very detailed debug information
	TraceLevel
	// DebugLevel for debug information
	DebugLevel
	// InfoLevel for informational messages
	InfoLevel
	// WarnLevel for warning messages
	WarnLevel
	// ErrorLevel for error messages
	ErrorLevel
)

func (l LogLevel) String() string {
	switch l {
	case Log:
		return "LOG"
	case TraceLevel:
		return "TRACE"
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// SysLogIsh represents a structured log message.
type SysLogIsh struct {
	Level   LogLevel
	Message string
}

// String returns the string representation of a SysLog.
func (s SysLogIsh) String() string {
	return fmt.Sprintf("[%s] %s", s.Level, s.Message)
}

// Logf sends a formatted log message with the specified level.
// This is a general-purpose logging function that accepts any LogLevel.
// For convenience, use Tracef, Debugf, Infof, Warnf, or Errorf for specific levels.
func (h *Handler) log(level LogLevel, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	syslog := SysLogIsh{
		Level:   level,
		Message: msg,
	}
	h.Send(syslog)
}

// Tracef sends a trace-level log message.
func (h *Handler) Logf(format string, args ...any) {
	h.log(Log, format, args...)
}

// Tracef sends a trace-level log message.
func (h *Handler) Tracef(format string, args ...any) {
	h.log(TraceLevel, format, args...)
}

// Debugf sends a debug-level log message.
func (h *Handler) Debugf(format string, args ...any) {
	h.log(DebugLevel, format, args...)
}

// Infof sends an info-level log message.
func (h *Handler) Infof(format string, args ...any) {
	h.log(InfoLevel, format, args...)
}

// Warnf sends a warning-level log message.
func (h *Handler) Warnf(format string, args ...any) {
	h.log(WarnLevel, format, args...)
}

// Errorf sends an error-level log message.
func (h *Handler) Errorf(format string, args ...any) {
	h.log(ErrorLevel, format, args...)
}
