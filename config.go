package rain

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/fsnotify/fsnotify"
	"github.com/panjf2000/ants/v2"

	"github.com/yrbb/rain/pkg/database"
	"github.com/yrbb/rain/pkg/redis"
	"github.com/yrbb/rain/pkg/utils"
)

type Config struct {
	mu sync.Mutex

	Debug    bool              `toml:"debug"`
	Project  string            `toml:"project"`
	Logger   logConfig         `toml:"logger"`
	Server   serverConfig      `toml:"server"`
	Worker   workerConfig      `toml:"worker"`
	Database []database.Config `toml:"database"`
	Redis    []redis.Config    `toml:"redis"`
	Custom   map[string]any    `toml:"custom"`
}

type serverConfig struct {
	Listen       string        `toml:"listen"`        // 监听, eg: 11.*:80
	ReadTimeout  time.Duration `toml:"read_timeout"`  // second
	WriteTimeout time.Duration `toml:"write_timeout"` // second
	StopTimeout  time.Duration `toml:"stop_timeout"`  // second
	EnablePProf  bool          `toml:"enable_pprof"`
}

func (s *serverConfig) validate() error {
	host, port, err := parseServerHostPort(s.Listen)
	if err != nil {
		return err
	}

	s.Listen = fmt.Sprintf("%s:%d", host, port)

	s.ReadTimeout *= time.Second
	s.WriteTimeout *= time.Second

	// 默认停止超时时间 10s
	if s.StopTimeout == 0 {
		s.StopTimeout = 10
	}
	s.StopTimeout *= time.Second

	return nil
}

type logConfig struct {
	Path      string        `toml:"path"`
	Level     string        `toml:"level"`
	SplitTime time.Duration `toml:"split_time"`
}

func (l *logConfig) validate() error {
	if l.Path == "" {
		l.Path = "/tmp"
	}

	if l.Level == "" {
		l.Level = "debug"
	}

	if l.SplitTime == 0 {
		l.SplitTime = 60
	}

	if ok, _ := utils.IsWritable(l.Path); !ok {
		return errors.New("日志目录不存在或者不可写")
	}

	return nil
}

type workerConfig struct {
	Capacity         int           `toml:"capacity"`
	ExpireTime       time.Duration `toml:"expire_time"`
	PreAlloc         bool          `toml:"pre_alloc"`
	MaxBlockingTasks int           `toml:"max_blocking_tasks"`
	Nonblocking      bool          `toml:"non_blocking"`
}

func (w *workerConfig) Options() []ants.Option {
	var opts []ants.Option

	if w.ExpireTime > 0 {
		opts = append(opts, ants.WithExpiryDuration(w.ExpireTime*time.Second))
	}

	if w.MaxBlockingTasks > 0 {
		opts = append(opts, ants.WithMaxBlockingTasks(w.MaxBlockingTasks))
	}

	opts = append(opts, ants.WithPreAlloc(w.PreAlloc))
	opts = append(opts, ants.WithNonblocking(w.Nonblocking))
	// opts = append(opts, ants.WithLogger(&internal.Logger{}))

	return opts
}

func parseConfig(file string) (*Config, error) {
	// file = getConfigPath() + file

	if !utils.FileExists(file) {
		return nil, fmt.Errorf("配置文件不存在: %s", file)
	}

	var cfg Config
	if _, err := toml.DecodeFile(file, &cfg); err != nil {
		return nil, err
	}

	if cfg.Project == "" {
		return nil, errors.New("配置项目名不能为空")
	}

	if err := cfg.Logger.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func watchConfig(watcher *fsnotify.Watcher, file string, callback func(*Config)) error {
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				switch event.Op {
				case fsnotify.Write:
					fmt.Printf("[CONFIG] Watch %s write ...\n\n", file)
					var cfg Config
					if _, err := toml.DecodeFile(file, &cfg); err != nil {
						continue
					}

					callback(&cfg)
				case fsnotify.Rename:
					fmt.Printf("[CONFIG] Watch %s rename to %s ...\n\n", file, event.Name)
					if event.Name != file {
						continue
					}

					var cfg Config
					if _, err := toml.DecodeFile(file, &cfg); err != nil {
						continue
					}

					callback(&cfg)
					watcher.Remove(file)
					watcher.Add(file)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					continue
				}

				if err != nil {
					fmt.Printf("[CONFIG] Watch error: %s\n", err)
					return
				}
			}
		}
	}()

	return watcher.Add(file)
}

func (c *Config) Get(name string) any {
	c.mu.Lock()
	defer c.mu.Unlock()

	if v, ok := c.Custom[name]; ok {
		if val, ok := v.(string); ok {
			if res, err := strconv.ParseInt(v.(string), 10, 64); err == nil {
				return res
			}

			return val
		}

		return v
	}

	return nil
}

func (c *Config) Set(name string, value any) {
	c.mu.Lock()
	c.Custom[name] = value
	c.mu.Unlock()
}
