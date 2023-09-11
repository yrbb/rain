package writer

import "log/slog"

type Writer interface {
	Write(lvl slog.Level, data []byte) error
	Close()
}
