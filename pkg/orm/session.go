package orm

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

type Session struct {
	orm *Orm

	error error
	table *modelInfo

	queryTimeout time.Duration
	queryStart   time.Time
	queryTime    float64

	tx   *sql.Tx
	txId int

	insertId     int64
	rowsAffected int64

	sql  string
	args []any

	options    []string
	columns    []string
	orderBy    []string
	grainpBy   []string
	forceIndex string
	where      []conditionStore
	limit      int
	offset     int

	set    map[string]any // for update
	fields []string       // for insert
	values int            // for insert
	colIdx []int          // for select
}

func (s *Session) SetTimeout(t time.Duration) *Session {
	s.queryTimeout = t

	return s
}

func (s *Session) after(err error) {
	s.queryTime = float64(time.Since(s.queryStart).Milliseconds())

	if s.orm.config.SlowThreshold > 0 && s.queryTime >= s.orm.config.SlowThreshold {
		slog.Warn(fmt.Sprintf(
			"long query [%.6f], sql: %s, args: %v",
			s.queryTime, s.sql, s.args,
		))
	}

	var tableName string
	if s.table != nil {
		tableName = s.table.Name
	}

	fields := []any{
		slog.String("table", tableName),
		slog.String("sql", s.sql),
		slog.Any("args", s.args),
		// slog.Any("params", s.params),
		slog.Float64("took", s.queryTime),
		slog.Int64("rowsAffected", s.rowsAffected),
		slog.Int64("insertId", s.insertId),
	}

	slog.Debug("query", fields...)
}

func (s *Session) reset() {
	s.queryTimeout = 0
	s.error = nil

	if s.tx == nil {
		s.insertId = 0
		s.rowsAffected = 0
	}

	s.error = nil
	s.sql = ""
	s.limit = 0
	s.offset = 0
	s.options = nil
	s.columns = nil
	s.orderBy = nil
	s.grainpBy = nil
	s.forceIndex = ""
	s.where = nil
	s.set = nil
	s.fields = nil
	s.values = 0
	s.args = nil

	s.table = nil
	s.colIdx = nil
}
