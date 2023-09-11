package redis

import (
	"context"
	"fmt"
	"log/slog"
)

type Logger struct{}

func (l *Logger) Printf(ctx context.Context, format string, v ...interface{}) {
	slog.Info(fmt.Sprintf(format, v...))
}
