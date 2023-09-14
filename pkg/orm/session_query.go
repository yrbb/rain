package orm

import (
	"context"
	"database/sql"
	"reflect"
	"time"
)

func (s *Session) Query(sqlStr string, values []any, errs ...error) (rows *sql.Rows, err error) {
	defer func() {
		s.after(err)
	}()

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

func (s *Session) Get(obj IModel) (bool, error) {
	defer s.reset()

	if s.error != nil {
		return false, s.error
	}

	_, find, err := s.getRow(obj)

	return find, err
}

func (s *Session) GetMap(obj IModel, m map[string]any) (bool, error) {
	defer s.reset()

	if s.error != nil {
		return false, s.error
	}

	v, find, err := s.getRow(obj)
	if err != nil {
		return find, err
	}

	if find {
		for _, idx := range s.colIdx {
			m[s.table.Fields[idx-1]] = v.Elem().Field(idx).Interface()
		}
	}

	return find, err
}

func (s *Session) getRow(obj IModel) (v reflect.Value, find bool, err error) {
	s.Limit(1)

	if v, err = s.queryCheck(obj); err != nil {
		return
	}

	ptrs := make([]any, len(s.colIdx))
	for i, idx := range s.colIdx {
		ptrs[i] = v.Elem().Field(idx).Addr().Interface()
	}

	var rows *sql.Rows
	rows, err = s.Query(s.buildSelectSQL())
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
		if s.table, err = s.orm.getModelInfo(et, true); err != nil {
			return
		}
	}

	if s.colIdx == nil {
		s.makeSelectFields()
	}

	var rows *sql.Rows

	rows, err = s.Query(s.buildSelectSQL())
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

func (s *Session) FindMap(obj IModel, m *[]map[string]any) (find bool, err error) {
	defer s.reset()

	if s.error != nil {
		return find, s.error
	}

	var v reflect.Value
	if v, err = s.queryCheck(obj); err != nil {
		return
	}

	ptrs := make([]any, len(s.colIdx))
	for i, idx := range s.colIdx {
		ptrs[i] = v.Elem().Field(idx).Addr().Interface()
	}

	var rows *sql.Rows
	rows, err = s.Query(s.buildSelectSQL())
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

func (s *Session) queryCheck(obj IModel) (v reflect.Value, err error) {
	v = reflect.ValueOf(obj)

	if s.table == nil {
		if s.table, err = s.orm.getModelInfo(obj); err != nil {
			return
		}
	}

	if s.colIdx == nil {
		s.makeSelectFields()
	}

	return
}

func (s *Session) makeSelectFields() {
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

func (s *Session) Count(obj ...IModel) (i int64, err error) {
	s.columns = []string{}

	s.Columns("count(1) as count")

	err = s.Pluck(&i, obj...)

	return
}

func (s *Session) Max(field string, obj ...IModel) (i int64, err error) {
	s.columns = []string{}

	s.Columns("max(" + field + ") as max")

	err = s.Pluck(&i, obj...)

	return
}

func (s *Session) Sum(field string, obj ...IModel) (i int64, err error) {
	s.columns = []string{}

	s.Columns("sum(" + field + ") as sum")

	err = s.Pluck(&i, obj...)

	return
}

func (s *Session) Pluck(v any, obj ...IModel) (err error) {
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
