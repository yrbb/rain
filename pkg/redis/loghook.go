package redis

import (
	"context"
	"log/slog"
	"net"
	"time"

	"github.com/redis/go-redis/v9"
)

var _ redis.Hook = &LogHook{}

type LogHook struct {
	Name  string
	Debug bool
}

func (h *LogHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

func (h *LogHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmd)

		if err != nil && err != redis.Nil {
			slog.Error(
				"redis-query",
				slog.String("name", h.Name),
				slog.String("cmd", cmd.Name()),
				slog.Any("args", cmd.Args()),
				slog.Int("took", int(time.Since(start).Milliseconds())),
				slog.String("error", err.Error()),
			)
		} else if h.Debug {
			slog.Debug(
				"redis-query",
				slog.String("name", h.Name),
				slog.String("cmd", cmd.Name()),
				slog.Any("args", cmd.Args()),
				slog.Int("took", int(time.Since(start).Milliseconds())),
			)
		}

		return err
	}
}
func (h *LogHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		return next(ctx, cmds)
	}
}
