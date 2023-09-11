package orm

import (
	"database/sql"
	"log/slog"
	"time"
)

type Session struct {
	orm *Orm

	tx   *sql.Tx
	txID int

	queryTimeout time.Duration

	status bool
	error  error

	insertID     int64
	rowsAffected int64

	queryStart time.Time
	queryTime  float64

	table *ModelInfo

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

	colIdx []int          // for select
	params map[string]any // for insert、update cb
}

func (s *Session) WithTimeout(t time.Duration) *Session {
	s.queryTimeout = t

	return s
}

func (s *Session) after(op string, status bool) {
	s.queryTime = time.Since(s.queryStart).Seconds()
	s.status = status

	if s.orm.longQueryTime > 0 && s.queryTime >= s.orm.longQueryTime {
		// s.orm.log().Warnf(
		// 	"long query [%.6f], sql: %s, args: %v",
		// 	s.queryTime, s.GetSQL(), s.params,
		// )
	}

	// var table string
	// if s.table != nil {
	// 	table = s.table.Name
	// }

	// fields := []any{
	// 	slog.Float64("took", float64(elapsed.Nanoseconds())/1e6),
	// 	slog.Int64("rows", rows),
	// 	slog.String("sql", sql),
	// }

	// FIXME
	slog.Debug("query")
	// .Fields(map[string]any{
	// 	"Table":        table,
	// 	"SQL":          s.sql,
	// 	"Args":         s.args,
	// 	"Params":       s.params,
	// 	"InsertId":     s.insertID,
	// 	"RowsAffected": s.rowsAffected,
	// 	"StartTime":    s.queryStart.Format("2006-01-02 15:04:05.000000"),
	// 	"QueryTime":    s.queryTime,
	// 	"Status":       s.status,
	// }).Debug()
}

func (s *Session) reset() {
	s.queryTimeout = 0
	s.error = nil

	if s.tx == nil {
		s.insertID = 0
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
	s.args = nil

	s.table = nil
	s.colIdx = nil
	s.params = nil
}
