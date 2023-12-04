package orm

import (
	"context"
	"database/sql"
	"reflect"
	"strings"
	"time"

	"github.com/yrbb/rain/pkg/utils"
)

func (s *Session) Query(sqlStr string, values []any, errs ...error) (rows *sql.Rows, err error) {
	defer s.after("query", err)

	if len(errs) > 0 && errs[0] != nil {
		return nil, errs[0]
	}

	if sqlStr == "" {
		return nil, ErrSQLEmpty
	}

	if s.sql == "" {
		s.sql = sqlStr
		s.args = values
	}

	s.queryStart = time.Now()

	ctx, cancel := context.Background(), context.CancelFunc(func() {})
	if s.queryTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, s.queryTimeout)
	}
	defer cancel()

	if s.tx != nil {
		rows, err = s.tx.QueryContext(ctx, sqlStr, values...)
	} else {
		rows, err = s.orm.db.QueryContext(ctx, sqlStr, values...)
	}

	return
}

func (s *Session) QueryMap(sqlStr string, values []any) ([]map[string]any, error) {
	rows, err := s.Query(sqlStr, values)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return convertRows(rows)
}

func (s *Session) QueryStruct(sqlStr string, values []any, obj any) error {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr {
		return ErrNeedPointer
	}

	sv := reflect.Indirect(v)
	if sv.Kind() != reflect.Slice {
		return ErrNeedPtrToSlice
	}

	et := sv.Type().Elem()

	var isPtr bool
	if et.Kind() == reflect.Ptr {
		isPtr = true
		et = et.Elem()
	}

	rows, err := s.Query(sqlStr, values)
	if err != nil {
		return err
	}

	defer rows.Close()

	cls, err := rows.Columns()
	if err != nil {
		return err
	}

	colIdx := []int{}
	for i := 0; i < et.NumField(); i++ {
		col := et.Field(i).Tag.Get("db")
		if col != "" {
			col = strings.Fields(col)[0]
		} else {
			col = et.Field(i).Tag.Get("json")
			if col != "" {
				col = strings.Split(col, ",")[0]
			}
		}

		if col == "" {
			col = et.Field(i).Name
		}

		if utils.SliceIn(col, cls) {
			colIdx = append(colIdx, i)
		}
	}

	for rows.Next() {
		nv := reflect.New(et)
		ni := reflect.Indirect(nv)

		var ptrs []any

		for _, idx := range colIdx {
			ptrs = append(ptrs, ni.Field(idx).Addr().Interface())
		}

		if err = rows.Scan(ptrs...); err != nil {
			return err
		}

		if isPtr {
			sv.Set(reflect.Append(sv, nv.Elem().Addr()))
		} else {
			sv.Set(reflect.Append(sv, nv.Elem()))
		}
	}

	return nil
}

func (s *Session) Count(obj ...any) (i int64, err error) {
	s.columns = []string{}

	s.Columns("count(1) as count")

	err = s.Pluck(&i, obj...)

	return
}

func (s *Session) Max(field string, obj ...any) (i int64, err error) {
	s.columns = []string{}

	s.Columns("max(" + field + ") as max")

	err = s.Pluck(&i, obj...)

	return
}

func (s *Session) Sum(field string, obj ...any) (i int64, err error) {
	s.columns = []string{}

	s.Columns("sum(" + field + ") as sum")

	err = s.Pluck(&i, obj...)

	return
}

func (s *Session) Pluck(v any, obj ...any) (err error) {
	defer s.reset()

	if s.table == nil {
		if len(obj) == 0 {
			return ErrMissingModel
		}

		s.table, err = s.orm.getModelInfo(obj[0])
		if err != nil {
			return
		}
	}

	var rows *sql.Rows
	if rows, err = s.Query(s.buildSelectSQL()); err != nil {
		return
	}

	defer func() {
		_ = rows.Close()
	}()

	var r any

	if rows.Next() {
		if err = rows.Scan(&r); err != nil {
			return
		}

		if r != nil {
			if err = convertAssign(v, r); err != nil {
				return
			}
		}
	}

	return
}
