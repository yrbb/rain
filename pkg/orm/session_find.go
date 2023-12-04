package orm

import (
	"database/sql"
	"reflect"
)

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

		if isPtr {
			sv.Set(reflect.Append(sv, nv.Elem().Addr()))
		} else {
			sv.Set(reflect.Append(sv, nv.Elem()))
		}
	}

	return
}

func (s *Session) FindMap(obj any) (m []map[string]any, err error) {
	defer s.reset()

	if s.error != nil {
		return nil, s.error
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

		mp := map[string]any{}

		for _, idx := range s.colIdx {
			key := s.table.Fields[idx]
			mp[key] = v.Elem().Field(idx).Interface()
		}

		m = append(m, mp)
	}

	return
}
