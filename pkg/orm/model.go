package orm

import (
	"reflect"
	"strings"
	"sync"

	"github.com/yrbb/rain/pkg/utils"
)

type ITableName interface {
	TableName() string
}

type IModel interface {
	ColIdx() []int
	Before() any
	Original() *reflect.Value
}

type Model struct {
	colIdx   []int
	original *reflect.Value
}

func (d *Model) ColIdx() []int {
	return d.colIdx
}

func (d *Model) Before() any {
	if d.original == nil {
		return nil
	}

	return d.original.Elem().Addr().Interface()
}

func (d *Model) Original() *reflect.Value {
	return d.original
}

type ModelInfo struct {
	Name          string
	Fields        []string
	PrimaryKeys   []*string
	UniqueKeys    map[string][]*string
	AutoIncrement *string
}

type modelParser struct {
	models sync.Map
}

func (o *modelParser) GetModelInfo(table any, isType ...bool) (*ModelInfo, error) {
	var (
		v reflect.Value
		t reflect.Type
	)

	if len(isType) > 0 && isType[0] {
		t = table.(reflect.Type)
	} else {
		v = reflect.ValueOf(table)

		if v.Kind() == reflect.Ptr {
			t = v.Elem().Type()
		} else {
			t = v.Type()
		}
	}

	// 结构体名
	sName := t.Name()

	if !v.IsValid() {
		v = reflect.New(t)
	}

	// 是否继承 Model
	if _, ok := v.Interface().(IModel); !ok {
		return nil, ErrUnknownStruct
	}

	// 第一个字段是否 Model
	if _, ok := v.Elem().Field(0).Interface().(Model); !ok {
		return nil, ErrUnknownStruct
	}

	//  表名
	tName := ""

	// 有 tableName 方法，则从 tableName 方法取得表名
	if fn, ok := v.Interface().(ITableName); ok {
		tName = fn.TableName()
	} else {
		// 否则认为结构体名首字母小写为表名
		tName = strings.ToUpper(sName[:1]) + sName[1:]
	}

	// 是否已经缓存过
	if s, ok := o.models.Load(tName); ok {
		return s.(*ModelInfo), nil
	}

	// 如果是指针则取得具体元素
	// if v.Kind() == reflect.Ptr {
	// 	v = v.Elem()
	// }

	// 缓存表信息
	sc, err := o.cacheTableInfo(t, tName)
	if err != nil {
		return nil, err
	}

	return sc, nil
}

// 缓存表信息
func (o *modelParser) cacheTableInfo(t reflect.Type, tName string) (*ModelInfo, error) {
	// 表字段数
	fNum := t.NumField()

	// 实例化一个 TableInfo
	newVal := &ModelInfo{
		Name:        tName,
		Fields:      make([]string, fNum-1),
		PrimaryKeys: []*string{},
		UniqueKeys:  map[string][]*string{},
	}

	endIndex := -1

	for i := 1; i < fNum; i++ {
		newVal.Fields[i-1] = strings.ToUpper(t.Field(i).Name[:1]) + t.Field(i).Name[1:]

		tag := t.Field(i).Tag.Get("db")
		if tag == "" {
			continue
		}

		// 指针
		fPtr := &newVal.Fields[i-1]

		tags := strings.Fields(tag)

		// 后面都忽略 ps: 在任意位置忽略字段改起来太麻烦，等有空吧~
		if utils.SliceIn(keywordIgnoreField, tags) {
			endIndex = i
			break
		}

		for _, v := range tags {
			if v == keywordAutoIncrement {
				newVal.AutoIncrement = fPtr
				continue
			}

			if v == keywordPrimaryKey {
				newVal.PrimaryKeys = append(newVal.PrimaryKeys, fPtr)
				continue
			}

			// 唯一字段
			if v[:9] == keywordUniqueKey {
				if u := strings.Replace(v, keywordUniqueKey, "", -1); u == "" {
					newVal.UniqueKeys[*fPtr] = append(newVal.UniqueKeys[*fPtr], fPtr)
				} else {
					u := u[1:]
					newVal.UniqueKeys[u] = append(newVal.UniqueKeys[u], fPtr)
				}

				continue
			}

			newVal.Fields[i-1] = v
		}
	}

	if len(newVal.PrimaryKeys) == 0 {
		return nil, ErrNoPrimaryKey
	}

	if endIndex > -1 {
		newVal.Fields = newVal.Fields[:endIndex-1]
	}

	o.models.Store(tName, newVal)

	return newVal, nil
}
