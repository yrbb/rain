package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/yrbb/rain/pkg/logger"
	"github.com/yrbb/rain/pkg/orm"
)

const (
	TypeMySQL    = "mysql"
	TypePostgres = "postgres"
	TypSQLLite3  = "sqllite3"
)

type dbInstance struct {
	db  *sql.DB
	orm *orm.Orm
}

type Database struct {
	list sync.Map // name => *dbInstance
}

func New(configs []Config) (*Database, error) {
	m := &Database{}

	initLen, initNum := 0, 0
	for _, v := range configs {
		if v.Disable {
			continue
		}

		initLen++
	}

	for _, v := range configs {
		v := v
		if v.Disable {
			continue
		}

		v.validate()

		initNum++

		logger.M().Info(fmt.Sprintf("初始化数据库 [%d/%d] %s", initNum, initLen, v.Name))

		if err := m.create(&v); err != nil {
			return nil, err
		}
	}

	return m, nil
}

func (m *Database) WithDebug(d bool) {}

func (m *Database) create(c *Config) error {
	db, err := orm.Open(&c.Config)
	if err != nil {
		return fmt.Errorf("orm open connect (%s) error: %v", c.Name, err)
	}

	sdb := db.DB()

	m.setOptions(sdb, c)

	m.list.Store(c.Name, &dbInstance{
		db:  sdb,
		orm: db,
	})

	return nil
}

func (m *Database) setOptions(db *sql.DB, c *Config) {
	if c.MaxOpenConns > 0 {
		db.SetMaxOpenConns(c.MaxOpenConns)
	}

	if c.MaxIdleConns > 0 {
		db.SetMaxIdleConns(c.MaxIdleConns)
	}

	if c.MaxLifeTime > 0 {
		db.SetConnMaxLifetime(time.Duration(c.MaxLifeTime) * time.Second)
	}
}

func (m *Database) Get(name ...string) (*orm.Orm, error) {
	if len(name) == 0 {
		name = []string{"default"}
	}

	if tmp, ok := m.list.Load(name[0]); ok {
		return tmp.(*dbInstance).orm, nil
	}

	return nil, fmt.Errorf("数据库资源未找到: %s", name[0])
}

func (m *Database) UpdateConfig(c []Config) bool {
	for _, v := range c {
		v := v
		v.validate()

		tmp, ok := m.list.Load(v.Name)
		if !ok {
			continue
		}

		m.setOptions(tmp.(*dbInstance).db, &v)
	}

	return true
}

func (m *Database) Close() {
	m.list.Range(func(key, v any) bool {
		if err := v.(*dbInstance).db.Close(); err != nil {
			logger.M().Error("关闭数据库异常", slog.String(key.(string), err.Error()))
		}

		m.list.Delete(key)

		return true
	})
}
