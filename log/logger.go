package log

// Logger specifies logging API.
type Logger interface {
	// SetLevel sets minimum logging level.
	SetLevel(Level)
	// Info logs any object in JSON format on info level.
	Info(interface{})
	// Warn logs any object in JSON format on warning level.
	Warn(interface{})
	// Error logs any object in JSON format on error level.
	Error(interface{})
	// Log logs anything with custom level.
	Log(Level, interface{})
}
