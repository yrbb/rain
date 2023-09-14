package orm

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/yrbb/rain/pkg/utils"
)

type conditionStore struct {
	Column    string
	Value     any
	Operator  string
	Connector string
	Bracket   string
}

type rawStore struct {
	value string
}

func (s *Session) Table(obj IModel) *Session {
	if ti, err := s.orm.getModelInfo(obj); err != nil {
		s.error = err
	} else {
		s.table = ti
	}

	return s
}

func (s *Session) Columns(columns ...string) *Session {
	if s.columns == nil {
		s.columns = []string{}
	}

	s.columns = append(s.columns, columns...)

	return s
}

func (s *Session) Options(option string) *Session {
	if s.options == nil {
		s.options = []string{}
	}

	s.options = append(s.options, option)

	return s
}

func (s *Session) ForceIndex(index string) *Session {
	s.forceIndex = index

	return s
}

func (s *Session) Where(column string, value any, operators ...string) *Session {
	operator := operateEquals
	if len(operators) > 0 {
		operator = operators[0]
	}

	s.criteria(&s.where, column, operator, value, logicalAnd)

	return s
}

func (s *Session) OrWhere(column string, value any, operators ...string) *Session {
	operator := operateEquals
	if len(operators) > 0 {
		operator = operators[0]
	}

	s.criteria(&s.where, column, operator, value, logicalOr)

	return s
}

func (s *Session) WhereMap(wheres map[string]any) *Session {
	for k, v := range wheres {
		s.Where(k, v)
	}

	return s
}

func (s *Session) WhereRaw(raw string) *Session {
	if raw == "" {
		return s
	}

	s.criteria(&s.where, "", "", rawStore{value: raw}, logicalAnd)

	return s
}

func (s *Session) GroupBy(grainp string) *Session {
	if s.grainpBy == nil {
		s.grainpBy = []string{}
	}

	s.grainpBy = append(s.grainpBy, grainp)

	return s
}

func (s *Session) OrderBy(column string, orders ...string) *Session {
	if s.orderBy == nil {
		s.orderBy = []string{}
	}

	order := operateOrderByAsc
	if len(orders) > 0 {
		order = orders[0]
	}

	s.orderBy = append(s.orderBy, column+" "+order)

	return s
}

func (s *Session) Limit(limit int, offset ...int) *Session {
	s.limit = limit

	if len(offset) > 0 {
		s.Offset(offset[0])
	}

	return s
}

func (s *Session) Offset(offset int) *Session {
	s.offset = offset

	return s
}

func (s *Session) Bracket(callback func(*Session), connectors ...string) *Session {
	if s.where == nil {
		s.where = []conditionStore{}
	}

	connector := logicalAnd
	if len(connectors) > 0 {
		connector = connectors[0]
	}

	s.where = append(s.where, conditionStore{
		Bracket:   bracketOpen,
		Connector: connector,
	})

	callback(s)

	s.where = append(s.where, conditionStore{
		Bracket:   bracketClose,
		Connector: connector,
	})

	return s
}

func (s *Session) Set(column any, value ...any) *Session {
	if s.set == nil {
		s.set = map[string]any{}
	}

	switch v := column.(type) {
	case string:
		if len(value) > 0 {
			s.set[v] = value[0]
		} else {
			s.set[v] = nil
		}
	case map[string]any:
		for key, val := range v {
			s.set[key] = val
		}
	default:
		if s.table == nil {
			var err error

			if s.table, err = s.orm.getModelInfo(column); err != nil {
				s.error = err

				return s
			}
		}

		r := reflect.Indirect(reflect.ValueOf(column))
		m := r.Field(0).Interface().(Model)

		if m.original != nil {
			s.setFromCompare(&r, m)
		} else {
			s.setFromModel(&r, value)
		}
	}

	return s
}

func (s *Session) setFromCompare(r *reflect.Value, m Model) {
	for _, idx := range m.colIdx {
		nv := (*r).Field(idx).Interface()
		ov := (*m.original).Elem().Field(idx).Interface()

		if nv != ov {
			s.set[s.table.Fields[idx-1]] = nv
		}
	}

	if len(s.set) == 0 {
		s.error = ErrNotSetUpdateField
	}
}

func (s *Session) setFromModel(r *reflect.Value, value []any) {
	fields := map[string]bool{}

	if len(value) > 0 {
		for _, val := range value {
			fields[val.(string)] = true
		}
	} else {
		exps := map[string]bool{}

		// FIXME 主键不一定是自增的
		if len(s.table.PrimaryKeys) > 0 {
			for _, val := range s.table.PrimaryKeys {
				exps[*val] = true
			}
		}

		for _, field := range s.table.Fields {
			if _, ok := exps[field]; ok {
				continue
			}

			fields[field] = true
		}
	}

	for idx, field := range s.table.Fields {
		if _, ok := fields[field]; ok {
			s.set[field] = (*r).Field(idx).Interface()
		}
	}
}

// (a=a+1)
func (s *Session) SetRaw(column, value string) *Session {
	if s.set == nil {
		s.set = map[string]any{}
	}

	s.set[column] = rawStore{value: value}

	return s
}

func (s *Session) Values(value any) *Session {
	hasFields := len(s.fields) > 0

	if kv, ok := value.(map[string]any); ok {
		value = []map[string]any{kv}
	}

	if kvs, ok := value.([]map[string]any); ok {
		for _, kv := range kvs {
			if hasFields && len(kv) != len(s.fields) {
				s.error = ErrFieldsNotMatch
				return s
			}

			for k, v := range kv {
				if !hasFields {
					s.fields = append(s.fields, k)
				}

				s.args = append(s.args, v)
			}

			hasFields = true
			s.values++
		}

		return s
	}

	if s.table == nil {
		if s.table, s.error = s.orm.getModelInfo(value); s.error != nil {
			return s
		}
	}

	vi := reflect.Indirect(reflect.ValueOf(value))

	for idx, field := range s.table.Fields {
		if s.table.AutoIncrement != nil && field == *s.table.AutoIncrement && utils.ToInt64(vi.Field(idx+1).Interface()) == 0 {
			continue
		}

		if !hasFields {
			s.fields = append(s.fields, field)
		}

		s.args = append(s.args, vi.Field(idx+1).Interface())
	}
	s.values++

	return s
}

func (s *Session) criteria(store *[]conditionStore, column string, operator string, value any, connector string) {
	if *store == nil {
		*store = []conditionStore{}
	}

	if matched, _ := regexp.MatchString("[!=<>]", operator); !matched {
		operator = strings.ToUpper(operator)

		if operator == operateIn || operator == operateNotIn {
			var v reflect.Value

			if v = reflect.ValueOf(value); v.Kind() == reflect.Ptr {
				v = v.Elem()
			}

			l := v.Len()
			ret := make([]any, l)

			for i := 0; i < l; i++ {
				ret[i] = v.Index(i).Interface()
			}

			value = ret
		}
	}

	column = strings.Trim(column, " ")

	*store = append(*store, conditionStore{
		Column:    column,
		Value:     value,
		Operator:  operator,
		Connector: connector,
	})
}
