package orm

import (
	"reflect"
	"regexp"
	"slices"
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
	key string
	val []any
}

func (s *Session) Table(obj any) *Session {
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

func (s *Session) Where(value ...any) *Session {
	return s.setWhere(logicalAnd, value)
}

func (s *Session) OrWhere(value ...any) *Session {
	return s.setWhere(logicalOr, value)
}

func (s *Session) GroupBy(group string) *Session {
	if s.groupBy == nil {
		s.groupBy = []string{}
	}

	s.groupBy = append(s.groupBy, group)

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
		s.setFromModel(&r, value)
	}

	return s
}

func (s *Session) setFromModel(r *reflect.Value, value []any) {
	fields := map[string]bool{}

	if len(value) > 0 {
		for _, val := range value {
			fields[val.(string)] = true
		}
	} else {
		for idx, field := range s.table.Fields {
			if slices.Contains(s.table.PrimaryKeys, idx) {
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

	s.set[column] = rawStore{key: value}

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
		if idx == s.table.AutoIncrement && utils.ToInt64(vi.Field(idx).Interface()) == 0 {
			continue
		}

		if !hasFields {
			s.fields = append(s.fields, field)
		}

		s.args = append(s.args, vi.Field(idx).Interface())
	}
	s.values++

	return s
}

func (s *Session) setWhere(connector string, values []any) *Session {
	l := len(values)
	if l == 0 {
		return s
	}

	if l == 1 {
		switch val := values[0].(type) {
		case string:
			s.criteria(&s.where, "", "", rawStore{key: val}, connector)
		case map[string]any:
			for k, v := range val {
				s.criteria(&s.where, k, operateEquals, v, connector)
			}
		default:
			s.error = ErrUnsupportedWhereType
		}

		return s
	}

	key, ok := values[0].(string)
	if !ok {
		s.error = ErrUnsupportedWhereType
		return s
	}

	if n := strings.Count(key, "?"); n > 0 {
		if len(values)-1 != n {
			s.error = ErrWhereArgsNotMatch
			return s
		}

		s.criteria(&s.where, "", "", rawStore{key: key, val: values[1:]}, connector)
		return s
	}

	if l > 3 {
		s.error = ErrUnsupportedWhereType
		return s
	}

	operator := operateEquals
	value := values[1]

	if l == 3 {
		hasOperator := false
		for i, v := range values {
			if k, ok := v.(string); ok && slices.Contains(operates, k) {
				hasOperator = true
				operator = values[i].(string)
			} else {
				value = values[i]
			}
		}

		if !hasOperator {
			s.error = ErrUnsupportedWhereType
			return s
		}
	}

	s.criteria(&s.where, key, operator, value, connector)
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
