package orm

import (
	"database/sql"
	"reflect"
)

func (s *Session) Get(obj any) (bool, error) {
	defer s.reset()

	if s.error != nil {
		return false, s.error
	}

	_, find, err := s.getRow(obj)

	return find, err
}

func (s *Session) GetMap(obj any) (map[string]any, error) {
	defer s.reset()

	if s.error != nil {
		return nil, s.error
	}

	v, find, err := s.getRow(obj)
	if err != nil {
		return nil, err
	}

	m := map[string]any{}
	if find {
		for _, idx := range s.colIdx {
			m[s.table.Fields[idx]] = v.Elem().Field(idx).Interface()
		}
	}

	return m, nil
}

func (s *Session) getRow(obj any) (v reflect.Value, find bool, err error) {
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

	return
}

func (s *Session) queryCheck(obj any) (v reflect.Value, err error) {
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
		if s.table, err = s.orm.getModelInfo(obj); err != nil {
			return
		}
	}

	if len(s.colIdx) == 0 {
		s.makeSelectFields()
	}

	if len(s.where) == 0 {
		_ = s.makeWhereCondition(v, true)
	}

	return
}

func (s *Session) makeSelectFields() {
	s.colIdx = []int{}

	if len(s.columns) > 0 {
		for _, c := range s.columns {
			for i, f := range s.table.Fields {
				if f == c {
					s.colIdx = append(s.colIdx, i)
					break
				}
			}
		}

		return
	}

	s.columns = []string{}

	for i, f := range s.table.Fields {
		s.columns = append(s.columns, f)
		s.colIdx = append(s.colIdx, i)
	}
}
