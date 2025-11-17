package logger

// Loggable represents an entity that can output log messages
type Loggable interface {
	Log(message string) error
}
