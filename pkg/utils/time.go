package utils

import (
	"fmt"
	"strings"
	"time"
)

func Date(format string, i ...any) string {
	patterns := []string{
		"Y", "2006", // 4 位数字完整表示的年份
		"m", "01", // 数字表示的月份，有前导零
		"d", "02", // 月份中的第几天，有前导零的 2 位数字
		"H", "15", // 小时，24 小时格式，有前导零
		"i", "04", // 有前导零的分钟数
		"s", "05", // 秒数，有前导零
		"v", "000000", // 毫秒
	}

	format = strings.NewReplacer(patterns...).Replace(format)

	t := time.Now()

	if len(i) > 0 {
		switch v := i[0].(type) {
		case int, int16, int32, int64, uint, uint16, uint32, uint64:
			d := ToInt64(v)
			if d < 0 {
				return ""
			}

			if d > 9999999999 {
				t = time.Unix(0, d*int64(time.Millisecond))
			} else {
				t = time.Unix(d, 0)
			}
		case time.Time:
			if v.IsZero() {
				return ""
			}

			t = v
		}
	}

	return t.Format(format)
}

func StrToTime(value string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04:05.000000",
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
		"2006-01-02",
		"2006/01/02",
		"20060102",
		"20060102150405",
	}

	var (
		t time.Time
		e error
	)

	for _, layout := range layouts {
		t, e = time.ParseInLocation(layout, value, time.Local)
		if e == nil {
			return t, nil
		}
	}

	return t, e
}

func Timestamp(v ...any) int64 {
	b := 10
	if len(v) > 0 {
		b = v[0].(int)
	}

	var n time.Time
	if len(v) > 1 {
		n = v[1].(time.Time)

		if n.IsZero() {
			return 0
		}
	} else {
		n = time.Now()
	}

	switch b {
	case 10:
		return n.UnixNano() / int64(time.Second)
	case 13:
		return n.UnixNano() / int64(time.Millisecond)
	case 16:
		return n.UnixNano() / int64(time.Microsecond)
	default:
		return n.UnixNano() / int64(time.Millisecond)
	}
}

func ExecTime(fn func() string, print ...bool) (content string) {
	st := time.Now()
	str := fn()
	dt := time.Since(st)

	content = fmt.Sprintf("%s, 耗时: [%s]", str, dt)

	if len(print) > 0 && !print[0] {
		content = fmt.Sprintf("[%s] %s", Date("Y-m-d H:i:s"), content)
		fmt.Println(content)
	}

	return content
}

func AvgTime(items []time.Duration) time.Duration {
	var sum time.Duration
	for _, item := range items {
		sum += item
	}

	return time.Duration(int64(sum) / int64(len(items)))
}
