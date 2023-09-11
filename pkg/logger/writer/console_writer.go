package writer

import (
	"log/slog"
	"os"
)

var _ Writer = &ConsoleWriter{}

func NewConsoleWriter() *ConsoleWriter {
	return &ConsoleWriter{}
}

type ConsoleWriter struct{}

func (c *ConsoleWriter) Write(_ slog.Level, data []byte) error {
	os.Stdout.Write(data)
	return nil
}

func (c *ConsoleWriter) Close() {}
