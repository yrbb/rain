package utils

import (
	"fmt"
	"strconv"
	"strings"
)

func ToInt(i any) (v int) {
	switch t := i.(type) {
	case int:
		v = t
	case int8:
		v = int(t)
	case int16:
		v = int(t)
	case int32:
		v = int(t)
	case int64:
		v = int(t)
	case uint:
		v = int(t)
	case uint8:
		v = int(t)
	case uint16:
		v = int(t)
	case uint32:
		v = int(t)
	case uint64:
		v = int(t)
	case float64:
		v = int(t)
	case string:
		vv, _ := strconv.ParseInt(t, 10, 64)
		v = int(vv)
	case []byte:
		vv, _ := strconv.ParseInt(string(t), 10, 64)
		v = int(vv)
	}

	return v
}

func ToInt64(i any) (v int64) {
	switch t := i.(type) {
	case int:
		v = int64(t)
	case int8:
		v = int64(t)
	case int16:
		v = int64(t)
	case int32:
		v = int64(t)
	case int64:
		v = t
	case uint:
		v = int64(t)
	case uint8:
		v = int64(t)
	case uint16:
		v = int64(t)
	case uint32:
		v = int64(t)
	case uint64:
		v = int64(t)
	case float64:
		v = int64(t)
	case string:
		v, _ = strconv.ParseInt(t, 10, 64)
	case []byte:
		v, _ = strconv.ParseInt(string(t), 10, 64)
	}

	return v
}

func ToString(i any) (v string) {
	switch t := i.(type) {
	case string:
		v = t
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		v = fmt.Sprintf("%d", t)
	case float32, float64:
		v = fmt.Sprintf("%v", t)
	case []byte:
		v = string(t)
	}

	return v
}

func ToFloat64(i any) (v float64) {
	switch t := i.(type) {
	case float32:
		v = float64(t)
	case float64:
		v = t
	case string:
		v, _ = strconv.ParseFloat(t, 64)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		v = float64(ToInt64(i))
	case []byte:
		v, _ = strconv.ParseFloat(string(t), 64)
	}

	return v
}

func ToBool(i any) bool {
	switch t := i.(type) {
	case string:
		return strings.EqualFold(t, "true")
	case int, int8, int16, int32, int64:
		return ToInt64(i) > 0
	case bool:
		return t
	}

	return false
}
