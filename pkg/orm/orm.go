package orm

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/yrbb/rain/pkg/utils"

	_ "github.com/go-sql-driver/mysql"
)

type Orm struct {
	db     *sql.DB
	config *Config
	models sync.Map
	exitCh chan struct{}
}

type Config struct {
	Name          string  `toml:"name"`
	Type          string  `toml:"type"` // mysql,postgres,sqllite3
	Addr          string  `toml:"addr"`
	MaxIdleConns  int     `toml:"max_idle_conns"`
	MaxOpenConns  int     `toml:"max_open_conns"`
	MaxLifeTime   int     `toml:"max_life_time"`
	SlowThreshold float64 `toml:"slowThreshold"`
	PoolThreshold int     `toml:"poolThreshold"`
}

func New(c *Config) (*Orm, error) {
	if c.Type == "" {
		return nil, errors.New("data type empty")
	}

	c.PoolThreshold = utils.If(c.PoolThreshold == 0, 80, c.PoolThreshold)

	o := &Orm{
		config: c,
		exitCh: make(chan struct{}),
	}

	db, err := sql.Open(c.Type, c.Addr)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	o.db = db

	if c.MaxIdleConns > 0 {
		o.SetMaxIdleConns(c.MaxIdleConns)
	}

	if c.MaxOpenConns > 0 {
		o.SetMaxOpenConns(c.MaxOpenConns)
	}

	go o.logStatus()

	runtime.SetFinalizer(o, func(o *Orm) {
		o.Close()
	})

	return o, nil
}

func (o *Orm) SetMaxIdleConns(n int) {
	o.db.SetMaxIdleConns(n)
}

func (o *Orm) SetMaxOpenConns(n int) {
	o.config.MaxOpenConns = n
	o.db.SetMaxOpenConns(n)
}

func (o *Orm) SetConnMaxLifetime(d time.Duration) {
	o.db.SetConnMaxLifetime(d)
}

func (o *Orm) Stats() sql.DBStats {
	return o.db.Stats()
}

func (o *Orm) Close() {
	if err := o.db.Close(); err != nil {
		slog.Error("close db err", slog.String("error", err.Error()))
	}

	o.exitCh <- struct{}{}
}

func (o *Orm) DB() *sql.DB {
	return o.db
}

func (o *Orm) NewSession() *Session {
	return &Session{orm: o}
}

func (o *Orm) logStatus() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-o.exitCh:
			return
		case <-ticker.C:
		}

		if o.config.PoolThreshold == -1 {
			return
		}

		used := o.db.Stats().OpenConnections
		tNum := o.config.MaxOpenConns * o.config.PoolThreshold / 100

		if used > tNum {
			percent := float32(used) / float32(o.config.MaxOpenConns) * 100

			slog.Warn(fmt.Sprintf(
				"数据库连接使用率高 [%s], used: %d/%d, percent: %.2f, threshold: %d%%",
				o.config.Name,
				used,
				o.config.MaxOpenConns,
				percent,
				o.config.PoolThreshold,
			))
		}
	}
}
