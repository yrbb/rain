package orm

import (
	"context"
	"database/sql"
	"reflect"
	"time"
)

func (s *Session) Query(sqlStr string, values []any, errs ...error) (rows *sql.Rows, err error) {
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

	if s.tx != nil {
		rows, err = s.tx.Query(sqlStr, values...)
	} else {
		if s.queryTimeout == 0 {
			rows, err = s.orm.db.Query(sqlStr, values...)
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), s.queryTimeout)
			rows, err = s.orm.db.QueryContext(ctx, sqlStr, values...)
			cancel()
		}
	}

	s.after("query", err == nil)

	return
}

func (s *Session) Get(obj any, m ...map[string]any) (bool, error) {
	defer s.reset()

	if len(m) > 0 {
		return s.GetMap(obj, m[0])
	}

	_, find, err := s.getOne(obj)

	return find, err
}

func (s *Session) GetMap(obj any, m map[string]any, defa ...bool) (bool, error) {
	defer s.reset()

	v, find, err := s.getOne(obj)
	if err != nil {
		return find, err
	}

	if find || (len(defa) > 0 && defa[0]) {
		for _, idx := range s.colIdx {
			m[s.table.Fields[idx-1]] = v.Elem().Field(idx).Interface()
		}
	}

	return find, nil
}

func (s *Session) getOne(obj any) (v reflect.Value, find bool, err error) {
	if s.error != nil {
		err = s.error
		return
	}

	s.Limit(1)

	if v, err = s.findCheck(obj); err != nil {
		return
	}

	ptrs := make([]any, len(s.colIdx))
	for i, idx := range s.colIdx {
		ptrs[i] = v.Elem().Field(idx).Addr().Interface()
	}

	var rows *sql.Rows
	rows, err = s.Query(s.BuildSelectSQL())
	if err != nil {
		return
	}

	defer func() {
		_ = rows.Close()
	}()

	if rows.Next() {
		if err = rows.Scan(ptrs...); err != nil {
			return
		}

		find = true
	}

	if err == nil {
		newValue := reflect.New(v.Elem().Type())

		for _, idx := range s.colIdx {
			oi := newValue.Elem().Field(idx).Addr().Interface()
			vi := v.Elem().Field(idx).Interface()

			_ = convertAssign(oi, vi)
		}

		v.Elem().Field(0).Set(reflect.ValueOf(Model{
			colIdx:   s.colIdx,
			original: &newValue,
		}))
	}

	return
}

func (s *Session) Find(obj any) (find bool, err error) {
	defer s.reset()

	if s.error != nil {
		return find, s.error
	}

	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr {
		return find, ErrNeedPointer
	}

	sv := reflect.Indirect(v)
	if sv.Kind() != reflect.Slice {
		return find, ErrNeedPtrToSlice
	}

	et := sv.Type().Elem()

	var isPtr bool
	if et.Kind() == reflect.Ptr {
		isPtr = true
		et = et.Elem()
	}

	if s.table == nil {
		if s.table, err = s.orm.GetModelInfo(et, true); err != nil {
			return find, err
		}
	}

	if s.colIdx == nil {
		s.parseSelectFields()
	}

	var rows *sql.Rows

	rows, err = s.Query(s.BuildSelectSQL())
	if err != nil {
		return
	}

	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		find = true

		nv := reflect.New(et)
		ni := reflect.Indirect(nv)

		var ptrs []any

		for _, idx := range s.colIdx {
			ptrs = append(ptrs, ni.Field(idx).Addr().Interface())
		}

		if err = rows.Scan(ptrs...); err != nil {
			return false, err
		}

		md := reflect.ValueOf(Model{
			colIdx:   s.colIdx,
			original: &nv,
		})

		nv.Elem().Field(0).Set(md)

		if isPtr {
			sv.Set(reflect.Append(sv, nv.Elem().Addr()))
		} else {
			sv.Set(reflect.Append(sv, nv.Elem()))
		}
	}

	return
}

func (s *Session) FindMap(obj any, m *[]map[string]any) (find bool, err error) {
	defer s.reset()

	if s.error != nil {
		err = s.error
		return
	}

	var v reflect.Value

	if v, err = s.findCheck(obj); err != nil {
		return
	}

	ptrs := make([]any, len(s.colIdx))
	for i, idx := range s.colIdx {
		ptrs[i] = v.Elem().Field(idx).Addr().Interface()
	}

	var rows *sql.Rows
	rows, err = s.Query(s.BuildSelectSQL())

	if err != nil {
		return
	}

	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		if err = rows.Scan(ptrs...); err != nil {
			return
		}

		find = true

		mp := map[string]any{}

		for _, idx := range s.colIdx {
			key := s.table.Fields[idx-1]
			mp[key] = v.Elem().Field(idx).Interface()
		}

		*m = append(*m, mp)
	}

	return
}

func (s *Session) findCheck(obj any) (v reflect.Value, err error) {
	v = reflect.ValueOf(obj)

	if v.Kind() != reflect.Ptr {
		err = ErrNeedPointer
		return
	}

	if k := reflect.Indirect(v).Kind(); k != reflect.Struct {
		err = ErrElementNeedStruct
		return
	}

	if s.table == nil {
		if s.table, err = s.orm.GetModelInfo(obj); err != nil {
			return
		}
	}

	if s.colIdx == nil {
		s.parseSelectFields()
	}

	return
}

func (s *Session) parseSelectFields() {
	s.colIdx = []int{}

	if s.columns != nil && len(s.columns) > 0 {
		for _, c := range s.columns {
			for i, f := range s.table.Fields {
				if f == c {
					s.colIdx = append(s.colIdx, i+1)
					break
				}
			}
		}

		return
	}

	s.columns = []string{}

	for i, f := range s.table.Fields {
		s.columns = append(s.columns, f)
		s.colIdx = append(s.colIdx, i+1)
	}
}

func (s *Session) Count(obj any) (i int64, err error) {
	s.columns = []string{}

	s.Columns("count(1) as count")

	err = s.Pluck(obj, &i)

	return
}

func (s *Session) Max(obj any, field string) (i int64, err error) {
	s.columns = []string{}

	s.Columns("max(" + field + ") as max")

	err = s.Pluck(obj, &i)

	return
}

func (s *Session) Sum(obj any, field string) (i int64, err error) {
	s.columns = []string{}

	s.Columns("sum(" + field + ") as sum")

	err = s.Pluck(obj, &i)

	return
}

func (s *Session) Pluck(obj any, v any) (err error) {
	defer s.reset()

	if s.table == nil {
		s.table, err = s.orm.GetModelInfo(obj)
		if err != nil {
			return
		}
	}

	var rows *sql.Rows

	if rows, err = s.Query(s.BuildSelectSQL()); err != nil {
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
