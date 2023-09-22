package orm

import (
	"reflect"
	"strings"
)

type model struct {
	Name          string
	Fields        map[int]string
	PrimaryKeys   []int
	UniqueKeys    map[string][]int
	AutoIncrement int
}

func (o *Orm) getModelInfo(table any, isType ...bool) (*model, error) {
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

	sName := t.Name()

	if !v.IsValid() {
		v = reflect.New(t)
	}

	tName := ""
	if fn, ok := v.Interface().(TableName); ok {
		tName = fn.TableName()
	} else {
		tName = strings.ToUpper(sName[:1]) + sName[1:]
	}

	if s, ok := o.models.Load(tName); ok {
		return s.(*model), nil
	}

	sc, err := o.parseTableInfo(t, tName)
	if err != nil {
		return nil, err
	}

	return sc, nil
}

func (o *Orm) parseTableInfo(t reflect.Type, tName string) (*model, error) {
	newVal := &model{
		Name:          tName,
		Fields:        map[int]string{},
		PrimaryKeys:   []int{},
		UniqueKeys:    map[string][]int{},
		AutoIncrement: -1,
	}

	uniqueKeys := map[any][]int{}

	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("db")
		if tag == "-" {
			continue
		}

		newVal.Fields[i] = strings.ToUpper(t.Field(i).Name[:1]) + t.Field(i).Name[1:]

		for _, v := range strings.Fields(tag) {
			if v == keywordAutoIncrement {
				newVal.AutoIncrement = i
				continue
			}

			if v == keywordPrimaryKey {
				newVal.PrimaryKeys = append(newVal.PrimaryKeys, i)
				continue
			}

			if len(v) > 9 && v[:9] == keywordUniqueKey {
				if u := strings.Replace(v, keywordUniqueKey, "", -1); u == "" {
					uniqueKeys[i] = append(uniqueKeys[i], i)
				} else {
					u := u[1:]
					uniqueKeys[u] = append(uniqueKeys[u], i)
				}

				continue
			}

			newVal.Fields[i] = v
		}
	}

	for k, v := range uniqueKeys {
		key, ok := k.(string)
		if !ok {
			key = newVal.Fields[k.(int)]
		}

		newVal.UniqueKeys[key] = v
	}

	if len(newVal.PrimaryKeys) == 0 {
		return nil, ErrNoPrimaryKey
	}

	o.models.Store(tName, newVal)

	return newVal, nil
}
