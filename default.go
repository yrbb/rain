package rain

import (
	"errors"
	"log/slog"
	"runtime"

	"github.com/panjf2000/ants/v2"
	"github.com/redis/go-redis/v9"

	"github.com/yrbb/rain/pkg/orm"
)

var rainIns *Rain

func Instance() *Rain {
	return rainIns
}

func Go(task func()) error {
	err := rainIns.worker.Submit(task)

	if errors.Is(err, ants.ErrPoolOverload) {
		slog.Warn("添加任务错误", slog.String("error", err.Error()))

		go func() {
			defer func() {
				if p := recover(); p != nil {
					var buf [4096]byte
					n := runtime.Stack(buf[:], false)
					slog.Error("任务异常", slog.Any("error", p), slog.String("stack", string(buf[:n])))
				}
			}()

			task()
		}()

		return nil
	}

	if err != nil {
		slog.Error("添加任务异常", slog.String("error", err.Error()))
	}

	return err
}

func Worker() *ants.Pool {
	return rainIns.worker
}

func Orm(name ...string) (*orm.Orm, error) {
	return rainIns.database.Get(name...)
}

func MustOrm(name ...string) *orm.Orm {
	s, _ := Orm(name...)
	return s
}

func Redis(name ...string) (*redis.Client, error) {
	return rainIns.redis.Get(name...)
}

func MustRedis(name ...string) *redis.Client {
	r, _ := Redis(name...)
	return r
}

func GetConfig(name string) any {
	return rainIns.config.Get(name)
}
