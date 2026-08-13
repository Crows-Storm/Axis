package logger

// Config holds logger configuration
type Config struct {
	// Level is the log level: trace, debug, info, warn, error, fatal, panic
	Level string
	// ServiceName is added to every log entry
	ServiceName string
	// LogDir is the directory for log files (empty = no file output)
	LogDir string
	// LogFileName overrides the default date-based filename
	LogFileName string
	// ConsoleOutput enables stdout output (default: true)
	ConsoleOutput *bool
}

// SetDefaults fills in zero values with sensible defaults
func (c *Config) SetDefaults() {
	if c.Level == "" {
		c.Level = "info"
	}
	if c.LogDir == "" {
		c.LogDir = "data/logs"
	}
	if c.ConsoleOutput == nil {
		t := true
		c.ConsoleOutput = &t
	}
}
