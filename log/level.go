package log

const (
	// All level is used when logging everything.
	All Level = iota
	// Error level is used when logging errors.
	Error
	// Warn level is used when logging warnings.
	Warn
	// Info level is used when logging info data.
	Info
)

var levels = map[Level]string{
	Error: "error",
	Warn:  "warn",
	Info:  "info",
}

// AddLevel adds custom level's name to levels map.
func AddLevel(lvl Level, name string) {
	levels[lvl] = name
}

// Level represents severity level while logging.
type Level int

func (lvl Level) String() string {
	return levels[lvl]
}
