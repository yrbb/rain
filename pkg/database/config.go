package database

import "github.com/yrbb/rain/pkg/orm"

const (
	MaxLifetime  = 2 * 3600 // 单位 time.Second
	MaxOpenConns = 100      // 设置数据库连接池最大连接数
	MaxIdleConns = 5        // 连接池最大允许的空闲连接数，如果没有 sql 任务需要执行的连接数大于5，超过的连接会被连接池关闭
)

type Config struct {
	orm.Config
	Disable bool `toml:"disable"`
}

func (c *Config) validate() {
	if c.MaxOpenConns <= 0 {
		c.MaxOpenConns = MaxOpenConns
	}

	if c.MaxIdleConns <= 0 {
		c.MaxIdleConns = MaxIdleConns
	}

	if c.MaxLifeTime <= 0 {
		c.MaxLifeTime = MaxLifetime
	}

	if c.SlowThreshold <= 0 {
		c.SlowThreshold = 1000
	}
}
