package orm

import (
	"fmt"
	"strings"

	"github.com/yrbb/rain/pkg/utils"
)

func (s *Session) BuildSelectSQL() (string, []any, error) {
	if s.sql == "" {
		where := s.buildWhereString()
		if where == "" {
			return "", nil, ErrWhereEmpty
		}

		from := s.buildFromString(false)
		if from == "" {
			return "", nil, ErrFromEmpty
		}

		parts := utils.SliceStringFilter([]string{
			s.buildSelectString(),
			from,
			s.buildForceIndexString(),
			where,
			s.buildGroupByString(),
			s.buildOrderByString(),
			s.buildLimitString(),
		})

		s.sql = strings.Join(parts, " ")
		s.args = []any{}

		s.getCriteriaValues(s.where)
	}

	return s.sql, s.args, nil
}

func (s *Session) BuildUpdateSQL() (string, []any, error) {
	if s.sql == "" {
		where := s.buildWhereString()
		if where == "" {
			return "", nil, ErrWhereEmpty
		}

		from := s.buildFromString(true)
		if from == "" {
			return "", nil, ErrFromEmpty
		}

		parts := utils.SliceStringFilter([]string{
			"UPDATE",
			from,
			s.buildSetString(),
			where,
			s.buildOrderByString(),
			s.buildLimitString(),
		})

		s.sql = strings.Join(parts, " ")

		s.getCriteriaValues(s.where)
	}

	return s.sql, s.args, nil
}

func (s *Session) BuildDeleteSQL() (string, []any, error) {
	if s.sql == "" {
		where := s.buildWhereString()
		if where == "" {
			return "", nil, ErrWhereEmpty
		}

		from := s.buildFromString(false)
		if from == "" {
			return "", nil, ErrFromEmpty
		}

		parts := utils.SliceStringFilter([]string{
			"DELETE",
			from,
			where,
			s.buildOrderByString(),
			s.buildLimitString(),
		})

		s.sql = strings.Join(parts, " ")
		s.args = []any{}

		s.getCriteriaValues(s.where)
	}

	return s.sql, s.args, nil
}

func (s *Session) BuildInsertSQL() (string, []any, error) {
	if s.sql == "" {
		from := s.buildFromString(true)
		if from == "" {
			return "", nil, ErrFromEmpty
		}

		values := s.buildValuesString()
		if values == "" {
			return "", nil, ErrNotSetInsertField
		}

		s.sql = strings.Join(utils.SliceStringFilter([]string{
			"INSERT",
			s.buildOptionString(),
			"INTO",
			from,
			values,
			s.buildSetString(true),
		}), " ")
	}

	return s.sql, s.args, nil
}

func (s *Session) buildSetString(insert ...bool) string {
	if len(s.set) == 0 {
		return ""
	}

	arr := []string{}

	if s.args == nil {
		s.args = []any{}
	}

	for k, v := range s.set {
		if rs, ok := v.(rawStore); ok {
			arr = append(arr, "`"+k+"`"+" = "+rs.value)
			continue
		}

		arr = append(arr, "`"+k+"`"+" = ?")
		s.args = append(s.args, v)
	}

	if len(insert) > 0 && insert[0] {
		return "ON DUPLICATE KEY UPDATE " + strings.Join(arr, ", ")
	}

	return "SET " + strings.Join(arr, ", ")
}

func (s *Session) buildValuesString() string {
	if s.fields == nil || len(s.fields) == 0 {
		return ""
	}

	if s.error != nil {
		return ""
	}

	str := "(`"
	str += strings.Join(s.fields, "`,`")
	str += "`) VALUES ("

	l := len(s.fields)

	val := make([]string, l)
	for i := 0; i < l; i++ {
		val[i] = "?"
	}

	str += strings.Join(val, ",")
	str += ")"

	return str
}

func (s *Session) buildSelectString() string {
	str := "SELECT "

	if opt := s.buildOptionString(); opt != "" {
		str += opt + " "
	}

	if len(s.columns) > 0 {
		columns := []string{}
		for _, v := range s.columns {
			if !strings.Contains(v, "(") && !strings.Contains(v, " ") {
				v = "`" + v + "`"
			}

			columns = append(columns, v)
		}

		return str + strings.Join(columns, ",")
	}

	return str + "*"
}

func (s *Session) buildFromString(tableOnly bool) string {
	str := ""

	if !tableOnly {
		str += "FROM "
	}

	return str + "`" + s.table.Name + "`"
}

func (s *Session) buildForceIndexString() string {
	if len(s.forceIndex) > 0 {
		return "FORCE INDEX(`" + s.forceIndex + "`)"
	}

	return ""
}

func (s *Session) buildWhereString() string {
	str := s.buildCriteriaString(s.where)

	if str != "" {
		return "WHERE " + str
	}

	return ""
}

func (s *Session) buildOrderByString() string {
	if len(s.orderBy) > 0 {
		return "ORDER BY " + strings.Join(s.orderBy, ", ")
	}

	return ""
}

func (s *Session) buildGroupByString() string {
	if len(s.grainpBy) > 0 {
		return "GROUP BY " + strings.Join(s.grainpBy, ", ")
	}

	return ""
}

func (s *Session) buildLimitString() string {
	str := ""

	if s.offset > 0 {
		str = fmt.Sprintf("LIMIT %d, %d", s.offset, s.limit)
	} else if s.limit > 0 {
		str = fmt.Sprintf("LIMIT %d", s.limit)
	}

	return str
}

func (s *Session) buildOptionString() string {
	if len(s.options) > 0 {
		return strings.Join(s.options, ", ")
	}

	return ""
}

func (s *Session) buildCriteriaString(store []conditionStore) string {
	if len(store) == 0 {
		return ""
	}

	statement, useConnector := "", false

	for _, item := range store {
		if item.Bracket != "" {
			isBracketOpen := item.Bracket == bracketOpen

			if isBracketOpen && useConnector {
				statement += " " + item.Connector + " "
			}

			useConnector = !isBracketOpen
			statement += item.Bracket

			continue
		}

		if useConnector {
			statement += " " + item.Connector + " "
		}

		useConnector = true

		if v, ok := item.Value.(rawStore); ok {
			statement += v.value
			continue
		}

		value := "?"

		if item.Operator == operateIn || item.Operator == operateNotIn {
			vals := strings.Repeat("?,", len(item.Value.([]any))-1) + "?"
			value = bracketOpen + vals + bracketClose
		} else if item.Operator == operateIs || item.Operator == operateIsNot {
			value = "NULL"
		}

		if strings.Contains(item.Column, "(") || strings.Contains(item.Column, " ") {
			statement += item.Column
		} else {
			statement += "`" + item.Column + "`"
		}

		statement += " " + item.Operator + " " + value
	}

	return statement
}

func (s *Session) getCriteriaValues(store []conditionStore) {
	if len(store) == 0 {
		return
	}

	for _, item := range store {
		if item.Bracket != "" {
			continue
		}

		if item.Operator == operateIn || item.Operator == operateNotIn {
			s.args = append(s.args, item.Value.([]any)...)

			continue
		}

		if item.Operator == operateIs || item.Operator == operateIsNot {
			continue
		}

		if _, ok := item.Value.(rawStore); ok {
			continue
		}

		s.args = append(s.args, item.Value)
	}
}
