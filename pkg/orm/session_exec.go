package orm

import (
	"context"
	"database/sql"
	"reflect"
	"time"
)

func (s *Session) Update(obj ...IModel) (rowsAffected int64, err error) {
	defer s.reset()

	if s.error != nil {
		return 0, s.error
	}

	if err = s.makeUpdateParams(obj); err != nil {
		return 0, err
	}

	if rowsAffected, err = s.updateDelete(s.buildUpdateSQL()); err != nil || rowsAffected == 0 {
		return 0, err
	}

	if len(obj) > 0 {
		oi := reflect.Indirect(reflect.ValueOf(obj[0]))

		for idx, field := range s.table.Fields {
			if val, ok := s.set[field]; ok {
				if _, ok := val.(rawStore); ok {
					continue
				}

				_ = convertAssign(oi.Field(idx+1).Addr().Interface(), val)
			}
		}
	}

	return rowsAffected, nil
}

func (s *Session) makeUpdateParams(obj []IModel) (err error) {
	lo := len(obj)
	if lo == 0 {
		if s.table == nil || len(s.set) == 0 {
			return ErrMissingModel
		}

		if len(s.where) == 0 {
			return ErrWhereEmpty
		}

		return
	}

	if s.table == nil {
		s.table, err = s.orm.getModelInfo(obj[0])
		if err != nil {
			return err
		}
	}

	if len(s.set) == 0 {
		s.Set(obj[0])
	}

	if s.error != nil {
		return s.error
	}

	if len(s.where) == 0 && obj[0].Original() != nil {
		if err = s.makeWhereCondition(obj[0]); err != nil {
			return err
		}
	}

	if len(s.where) == 0 {
		return ErrWhereEmpty
	}

	return nil
}

func (s *Session) Delete(obj ...IModel) (rowsAffected int64, err error) {
	defer s.reset()

	if s.error != nil {
		return 0, s.error
	}

	if s.table == nil {
		if len(obj) == 0 {
			return 0, ErrMissingModel
		}

		if s.table, err = s.orm.getModelInfo(obj[0]); err != nil {
			return 0, err
		}
	}

	if rowsAffected, err = s.updateDelete(s.buildDeleteSQL()); err != nil {
		return 0, nil
	}

	return rowsAffected, nil
}

func (s *Session) updateDelete(sqlStr string, values []any, err error) (int64, error) {
	if err != nil {
		return 0, err
	}

	res, err := s.Exec(sqlStr, values...)
	if err != nil {
		return 0, err
	}

	s.rowsAffected, err = res.RowsAffected()

	return s.rowsAffected, err
}

func (s *Session) Insert(obj ...IModel) (insertId int64, err error) {
	defer s.reset()

	if s.error != nil {
		return 0, s.error
	}

	if len(obj) == 0 {
		if s.table == nil || len(s.args) == 0 {
			return 0, ErrMissingModel
		}
	} else {
		if s.table == nil {
			s.table, err = s.orm.getModelInfo(obj[0])
			if err != nil {
				return 0, err
			}
		}

		if len(s.args) == 0 {
			s.Values(obj[0])
		}

		if s.error != nil {
			return 0, s.error
		}
	}

	sqlStr, values, err := s.buildInsertSQL()

	if s.error != nil {
		return 0, s.error
	}

	if err != nil {
		return 0, err
	}

	res, err := s.Exec(sqlStr, values...)
	if err != nil {
		return 0, err
	}

	insertId, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if s.values == 1 && len(obj) > 0 && s.table.AutoIncrement != nil {
		oi := reflect.Indirect(reflect.ValueOf(obj[0]))

		aIdx := 0
		for idx, field := range s.table.Fields {
			if field == *s.table.AutoIncrement {
				aIdx = idx
				break
			}
		}

		_ = convertAssign(oi.Field(aIdx+1).Addr().Interface(), insertId)
	}

	if s.tx == nil || s.insertId == 0 {
		s.insertId = insertId
	}

	if s.tx == nil || s.rowsAffected == 0 {
		s.rowsAffected, _ = res.RowsAffected()
	}

	return insertId, nil
}

func (s *Session) Exec(sqlStr string, values ...any) (res sql.Result, err error) {
	defer func() {
		s.after(err)
	}()

	s.queryStart = time.Now()

	ctx, cancel := context.Background(), context.CancelFunc(func() {})
	if s.queryTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, s.queryTimeout)
	}
	defer cancel()

	if s.tx != nil {
		res, err = s.tx.ExecContext(ctx, sqlStr, values...)
	} else {
		res, err = s.orm.db.ExecContext(ctx, sqlStr, values...)
	}

	return
}

func (s *Session) makeWhereCondition(m IModel) error {
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
