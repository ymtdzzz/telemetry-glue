package logger

import "log"

// StdoutLogger is a logger that writes logs to standard output
type StdoutLogger struct{}

func NewStdoutLogger() *StdoutLogger {
	return &StdoutLogger{}
}

func (l *StdoutLogger) Log(message string) error {
	log.Println(message)
	return nil
}
