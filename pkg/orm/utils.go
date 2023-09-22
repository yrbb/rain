package orm

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/yrbb/rain/pkg/utils"
)

func convertAssign(dest, src any) error {
	switch s := src.(type) {
	case string:
		switch d := dest.(type) {
		case *string:
			*d = s
		case *[]byte:
			*d = []byte(s)
		case *time.Time:
			*d, _ = utils.StrToTime(s)
		default:
			return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type %T", src, dest)
		}

		return nil

	case []byte:
		switch d := dest.(type) {
		case *string:
			*d = string(s)
		case *any:
			*d = bytesClone(s)
		case *[]byte:
			*d = bytesClone(s)
		case *int:
			*d = utils.ToInt(s)
		case *int64:
			*d = utils.ToInt64(s)
		case *time.Time:
			*d, _ = utils.StrToTime(string(s))
		case *float64:
			*d = utils.ToFloat64(s)
		default:
			return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type %T", src, dest)
		}

		return nil

	case time.Time:
		switch d := dest.(type) {
		case *time.Time:
			*d = s
		default:
			return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type %T", src, dest)
		}

		return nil

	case nil:
		switch d := dest.(type) {
		case *any:
			*d = nil
		case *[]byte:
			*d = nil
		default:
			return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type %T", src, dest)
		}

		return nil
	}

	switch d := dest.(type) {
	case *string:
		*d = utils.ToString(src)
	case *[]byte:
		*d = src.([]byte)
	case *bool:
		*d = src.(bool)
	default:
		dv := reflect.Indirect(reflect.ValueOf(dest))

		switch dv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			var s string

			v := reflect.ValueOf(src)
			switch v.Kind() {
			// 解决 php 传过来数据类型问题
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				s = strconv.FormatInt(int64(v.Uint()), 10)
			// 解决 json 解析时把 int 转为 float 问题
			case reflect.Float32, reflect.Float64:
				s = strconv.FormatInt(int64(v.Float()), 10)
			case reflect.Bool:
				s = "0"
			default:
				s = strconv.FormatInt(v.Int(), 10)
			}

			i64, err := strconv.ParseInt(s, 10, dv.Type().Bits())
			if err != nil {
				return strconvErr(src, s, dv.Kind(), err)
			}

			dv.SetInt(i64)

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			s := strconv.FormatUint(reflect.ValueOf(src).Uint(), 10)

			u64, err := strconv.ParseUint(s, 10, dv.Type().Bits())
			if err != nil {
				return strconvErr(src, s, dv.Kind(), err)
			}

			dv.SetUint(u64)
		case reflect.Float32, reflect.Float64:
			var s string

			rv := reflect.ValueOf(src)

			switch rv.Kind() {
			case reflect.Float64:
				s = strconv.FormatFloat(rv.Float(), 'g', -1, 64)
			case reflect.Float32:
				s = strconv.FormatFloat(rv.Float(), 'g', -1, 32)
			}

			f64, err := strconv.ParseFloat(s, dv.Type().Bits())
			if err != nil {
				return strconvErr(src, s, dv.Kind(), err)
			}

			dv.SetFloat(f64)
		default:
			return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type %T", src, dest)
		}
	}

	return nil
}

func strconvErr(src any, s string, kind reflect.Kind, err error) error {
	if ne, ok := err.(*strconv.NumError); ok {
		err = ne.Err
	}

	return fmt.Errorf("converting driver.Value type %T (%q) to a %s: %v", src, s, kind, err)
}

func bytesClone(b []byte) []byte {
	if b == nil {
		return nil
	}

	c := make([]byte, len(b))
	copy(c, b)

	return c
}

func sliceCount(a any) int {
	v := reflect.ValueOf(a)
	if v.Kind() != reflect.Slice {
		return 0
	}

	return v.Len()
}

func convertSlice(a any) []any {
	v := reflect.ValueOf(a)
	if v.Kind() != reflect.Slice {
		return nil
	}

	result := make([]any, v.Len())
	for i := 0; i < v.Len(); i++ {
		elemValue := v.Index(i)
		switch elemValue.Kind() {
		case reflect.String:
			result[i] = elemValue.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			result[i] = elemValue.Int()
		case reflect.Float32, reflect.Float64:
			result[i] = elemValue.Float()
		default:
			result[i] = elemValue.Interface()
		}
	}

	return result
}

func isZero(x any) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}
