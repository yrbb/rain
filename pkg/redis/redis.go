package redis

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/yrbb/rain/pkg/logger"
)

type Config struct {
	Disable         bool          `toml:"disable"`
	Name            string        `toml:"name"`
	Addr            string        `toml:"addr"`           // redis://<user>:<pass>@localhost:6379/<db>
	DialTimeout     time.Duration `toml:"dial_timeout"`   // default 5s
	ReadTimeout     time.Duration `toml:"read_timeout"`   // default 3s
	WriteTimeout    time.Duration `toml:"write_timeout"`  // default 3s
	PoolFIFO        bool          `toml:"pool_fifo"`      // LIFO, FIFO
	PoolSize        int           `toml:"pool_size"`      // default 10 * runtime.GOMAXPROCS
	MinIdleConns    int           `toml:"min_idle_conns"` // default 0
	MaxIdleConns    int           `toml:"max_idle_conns"` // default 0
	ConnMaxIdleTime time.Duration `toml:"max_open_conns"` // default 30分钟
	ConnMaxLifeTime time.Duration `toml:"delay_connect"`  // default 0
}

type Redis struct {
	hooks sync.Map // name => *LogHook
	list  sync.Map // name => *redis.Client
}

func New(configs []Config) (*Redis, error) {
	// set default logger
	redis.SetLogger(&Logger{})

	m := &Redis{}

	initLen, initNum := 0, 0
	for _, v := range configs {
		if !v.Disable {
			initLen++
		}
	}

	for _, v := range configs {
		v := v

		if v.Disable {
			continue
		}

		initNum++

		logger.M().Info(fmt.Sprintf("初始化 Redis [%d/%d] %s", initNum, initLen, v.Name))

		if err := m.create(&v); err != nil {
			return nil, err
		}
	}

	return m, nil
}

func (m *Redis) create(c *Config) error {
	opts, err := redis.ParseURL(c.Addr)
	if err != nil {
		return err
	}

	m.setOptions(opts, c)

	rdb := redis.NewClient(opts)

	cmd := rdb.Ping(context.Background())
	if cmd.Err() != nil {
		return cmd.Err()
	}

	logHook := &LogHook{Name: c.Name}
	rdb.AddHook(logHook)

	m.hooks.Store(c.Name, logHook)
	m.list.Store(c.Name, rdb)

	return nil
}

func (m *Redis) setOptions(opts *redis.Options, c *Config) {
	opts.PoolFIFO = c.PoolFIFO

	if c.DialTimeout > 0 {
		opts.DialTimeout = time.Second * c.DialTimeout
	}

	if c.ReadTimeout > 0 {
		opts.ReadTimeout = time.Second * c.ReadTimeout
	}

	if c.WriteTimeout > 0 {
		opts.WriteTimeout = time.Second * c.WriteTimeout
	}

	if c.PoolSize > 0 {
		opts.PoolSize = c.PoolSize
	}

	if c.MinIdleConns > 0 {
		opts.MinIdleConns = c.MinIdleConns
	}

	if c.MaxIdleConns > 0 {
		opts.MaxIdleConns = c.MaxIdleConns
	}

	if c.ConnMaxIdleTime > 0 {
		opts.ConnMaxIdleTime = time.Second * c.ConnMaxIdleTime
	}

	if c.ConnMaxLifeTime > 0 {
		opts.ConnMaxLifetime = time.Second * c.ConnMaxLifeTime
	}
}

func (m *Redis) Get(name ...string) (*redis.Client, error) {
	if len(name) == 0 {
		name = []string{"default"}
	}

	if tmp, ok := m.list.Load(name[0]); ok {
		return tmp.(*redis.Client), nil
	}

	return nil, fmt.Errorf("Redis 资源不存在: %s", name[0])
}

func (m *Redis) UpdateConfig(c []Config) bool {
	for _, v := range c {
		v := v

		tmp, ok := m.list.Load(v.Name)
		if !ok {
			continue
		}

		rc := tmp.(*redis.Client)

		if err := m.create(&v); err != nil {
			logger.M().Error("初始化 Redis 资源异常", slog.String("name", v.Name), slog.String("error", err.Error()))
			m.list.Store(v.Name, rc)

			continue
		}

		logger.M().Info(fmt.Sprintf("初始化 Redis %s", v.Name))

		if err := rc.Close(); err != nil {
			logger.M().Error("关闭 Redis 资源异常", slog.String("name", v.Name), slog.String("error", err.Error()))
		}
	}

	return true
}

func (m *Redis) SetDebug(d bool) {
	m.hooks.Range(func(_, v any) bool {
		v.(*LogHook).Debug = d
		return true
	})
}

func (m *Redis) Close() {
	m.list.Range(func(k, v any) bool {
		if err := v.(*redis.Client).Close(); err != nil {
			logger.M().Error("关闭 Redis 资源异常", slog.String("name", k.(string)), slog.String("error", err.Error()))
		}

		return true
	})
}
