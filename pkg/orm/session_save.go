package orm

import "reflect"

func (s *Session) Save(obj IModel, params map[string]any) (int64, error) {
	if obj.Original() == nil {
		return s.Set(params).Update(obj)
	}

	if len(params) == 0 {
		return 0, ErrUpdateParamsEmpty
	}

	ov := reflect.ValueOf(obj)

	if ov.Kind() != reflect.Ptr {
		return 0, ErrNeedPointer
	}

	var err error

	if s.table == nil {
		if s.table, err = s.orm.getModelInfo(obj); err != nil {
			return 0, err
		}
	}

	if s.where == nil || len(s.where) == 0 {
		if err = s.makeWhereCondition(obj); err != nil {
			return 0, err
		}
	}

	// bind params -> struct
	for idx, field := range s.table.Fields {
		if val, ok := params[field]; ok {
			_ = convertAssign(ov.Elem().Field(idx+1).Addr().Interface(), val)

			s.Set(field, val)
		}
	}

	n, err := s.updateDelete(s.buildUpdateSQL())
	if err != nil || n == 0 {
		return 0, err
	}

	return n, nil
}
