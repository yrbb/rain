package orm

import (
	"context"
	"database/sql"
	"reflect"
	"time"
)

func (s *Session) Update(obj ...any) (n int64, err error) {
	defer s.reset()

	if s.error != nil {
		return 0, s.error
	}

	if len(obj) == 0 {
		return 0, ErrMissingModel
	}

	if err = s.processUpdateParams(obj); err != nil {
		return 0, err
	}

	if n, err = s.updateDelete(s.BuildUpdateSQL()); err != nil || n == 0 {
		return 0, err
	}

	oi := reflect.Indirect(reflect.ValueOf(obj[0]))

	for idx, field := range s.table.Fields {
		if val, ok := s.set[field]; ok {
			if _, ok := val.(rawStore); ok {
				continue
			}

			_ = convertAssign(oi.Field(idx+1).Addr().Interface(), val)
		}
	}

	return n, nil
}

func (s *Session) processUpdateParams(obj []any) error {
	if s.table == nil {
		table, err := s.orm.GetModelInfo(obj[0])
		if err != nil {
			return err
		}

		s.table = table
	}

	if len(obj) > 1 {
		s.Set(obj[1])
	} else if s.set == nil || len(s.set) == 0 {
		s.Set(obj[0])
	}

	if s.error != nil {
		return s.error
	}

	if s.where == nil || len(s.where) == 0 {
		if model, ok := obj[0].(IModel); ok && model.Original() != nil {
			if err := s.parseWhereCondition(model); err != nil {
				return err
			}
		}
	}

	if len(s.where) == 0 {
		return ErrWhereEmpty
	}

	return nil
}

func (s *Session) Delete(obj any) (n int64, err error) {
	defer s.reset()

	if s.error != nil {
		return 0, s.error
	}

	if s.table, err = s.orm.GetModelInfo(obj); err != nil {
		return 0, err
	}

	if n, err = s.updateDelete(s.BuildDeleteSQL()); err != nil {
		return 0, nil
	}

	return n, nil
}

func (s *Session) updateDelete(sqlStr string, values []any, err error) (int64, error) {
	if err != nil {
		return 0, err
	}

	if s.params == nil || len(s.params) == 0 {
		s.params = map[string]any{}
		for _, v := range s.where {
			if v.Operator != "=" {
				continue
			}

			s.params[v.Column] = v.Value
		}
	}

	res, err := s.Statement(sqlStr, values...)
	if err != nil {
		return 0, err
	}

	if s.rowsAffected, err = res.RowsAffected(); err != nil {
		return 0, err
	}

	return s.rowsAffected, nil
}

func (s *Session) Insert(obj ...any) (n int64, err error) {
	defer s.reset()

	if s.error != nil {
		return 0, s.error
	}

	if err = s.processInsertParams(obj); err != nil {
		return 0, err
	}

	sqlStr, values, err := s.BuildInsertSQL()

	if s.error != nil {
		return 0, s.error
	}

	if err != nil {
		return 0, err
	}

	res, err := s.Statement(sqlStr, values...)
	if err != nil {
		return 0, err
	}

	insertID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if s.tx == nil || s.insertID == 0 {
		s.insertID = insertID
	}

	if s.tx == nil || s.rowsAffected == 0 {
		s.rowsAffected, _ = res.RowsAffected()
	}

	return insertID, nil
}

func (s *Session) processInsertParams(obj []any) error {
	lo := len(obj)
	if lo == 0 {
		return ErrMissingModel
	}

	if s.table == nil {
		table, err := s.orm.GetModelInfo(obj[0])
		if err != nil {
			return err
		}

		s.table = table
	}

	// TODO 一次插多条的情况未考虑
	if exists := s.args != nil && len(s.args) > 0; lo > 1 {
		if exists {
			return ErrDuplicateValues
		}

		s.Values(obj[1])
	} else if !exists {
		s.Values(obj[0])
	}

	if s.error != nil {
		return s.error
	}

	return nil
}

func (s *Session) Statement(sqlStr string, values ...any) (res sql.Result, err error) {
	s.queryStart = time.Now()

	if s.tx != nil {
		res, err = s.tx.Exec(sqlStr, values...)
	} else {
		if s.queryTimeout == 0 {
			res, err = s.orm.db.Exec(sqlStr, values...)
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), s.queryTimeout)
			res, err = s.orm.db.ExecContext(ctx, sqlStr, values...)
			cancel()
		}
	}

	s.after("statement", err == nil)

	return
}

func (s *Session) Save(obj any, params map[string]any) (int64, error) {
	model, ok := obj.(IModel)
	if !ok {
		return 0, ErrMissingModel
	}

	if model.Original() == nil {
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
		if s.table, err = s.orm.GetModelInfo(obj); err != nil {
			return 0, err
		}
	}

	if s.where == nil || len(s.where) == 0 {
		if err = s.parseWhereCondition(model); err != nil {
			return 0, err
		}
	}

	s.params = params

	// bind params -> struct
	for idx, field := range s.table.Fields {
		if val, ok := params[field]; ok {
			_ = convertAssign(ov.Elem().Field(idx+1).Addr().Interface(), val)

			s.Set(field, val)
		}
	}

	n, err := s.updateDelete(s.BuildUpdateSQL())
	if err != nil || n == 0 {
		return 0, err
	}

	return n, nil
}

func (s *Session) parseWhereCondition(m IModel) error {
	ori := *(m.Original())

	args := map[string]any{}
	for _, idx := range m.ColIdx() {
		args[s.table.Fields[idx-1]] = ori.Elem().Field(idx).Interface()
	}

	find := false
	where := map[string]any{}

	// has primary key
	if len(s.table.PrimaryKeys) > 0 {
		for _, pk := range s.table.PrimaryKeys {
			if v, ok := args[*pk]; ok {
				find = true
				where[*pk] = v

				continue
			}

			find = false

			break
		}
	}

	// has unique key
	if !find && len(s.table.UniqueKeys) > 0 {
		for _, uqs := range s.table.UniqueKeys {
			where = map[string]any{}

			for _, uq := range uqs {
				if v, ok := args[*uq]; ok {
					find = true
					where[*uq] = v

					continue
				}

				find = false

				break
			}

			if find {
				break
			}
		}
	}

	if !find {
		return ErrNoPrimaryAndUnique
	}

	for k, v := range where {
		s.Where(k, v)
	}

	return nil
}
