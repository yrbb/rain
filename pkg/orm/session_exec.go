package orm

import (
	"context"
	"database/sql"
	"reflect"
	"time"
)

func (s *Session) Update(obj ...any) (rowsAffected int64, err error) {
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

				_ = convertAssign(oi.Field(idx).Addr().Interface(), val)
			}
		}
	}

	return rowsAffected, nil
}

func (s *Session) makeUpdateParams(obj []any) (err error) {
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

	if len(s.where) == 0 {
		if err = s.makeWhereCondition(obj[0]); err != nil {
			return err
		}
	}

	if len(s.where) == 0 {
		return ErrWhereEmpty
	}

	return nil
}

func (s *Session) Delete(obj ...any) (rowsAffected int64, err error) {
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

		if len(s.where) == 0 {
			if err = s.makeWhereCondition(obj[0]); err != nil {
				return 0, err
			}
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

func (s *Session) Insert(obj ...any) (insertId int64, err error) {
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

	if s.values == 1 && len(obj) > 0 && s.table.AutoIncrement > -1 {
		oi := reflect.Indirect(reflect.ValueOf(obj[0]))
		_ = convertAssign(oi.Field(s.table.AutoIncrement).Addr().Interface(), insertId)
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
	defer s.after("exec", err)

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

func (s *Session) makeWhereCondition(m any) error {
	rv := reflect.Indirect(reflect.ValueOf(m))

	find := false
	where := map[string]any{}

	// has primary key
	if len(s.table.PrimaryKeys) > 0 {
		for _, i := range s.table.PrimaryKeys {
			if val := rv.Field(i).Interface(); isZero(val) {
				find = false
			} else {
				find = true
				where[s.table.Fields[i]] = val
			}

			if !find {
				break
			}
		}
	}

	// has unique key
	if !find && len(s.table.UniqueKeys) > 0 {
		for _, uqs := range s.table.UniqueKeys {
			where = map[string]any{}

			for _, uq := range uqs {
				if val := rv.Field(uq).Interface(); isZero(val) {
					find = false
				} else {
					find = true
					where[s.table.Fields[uq]] = val
				}

				if !find {
					break
				}
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
