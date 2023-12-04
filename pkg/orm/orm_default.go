package orm

import "database/sql"

func (o *Orm) Get(obj any) (find bool, err error) {
	return o.NewSession().Get(obj)
}

func (o *Orm) GetMap(obj any) (m map[string]any, err error) {
	return o.NewSession().GetMap(obj)
}

func (o *Orm) Find(obj []any) (find bool, err error) {
	return o.NewSession().Find(obj)
}

func (o *Orm) FindMap(obj any) (m []map[string]any, err error) {
	return o.NewSession().FindMap(obj)
}

func (o *Orm) Table(obj any) *Session {
	return o.NewSession().Table(obj)
}

func (o *Orm) Columns(columns ...string) *Session {
	return o.NewSession().Columns(columns...)
}

func (o *Orm) Where(value ...any) *Session {
	return o.NewSession().Where(value...)
}

func (o *Orm) Set(column any, value ...any) *Session {
	return o.NewSession().Set(column, value...)
}

func (o *Orm) SetRaw(column string, value string) *Session {
	return o.NewSession().SetRaw(column, value)
}

func (o *Orm) Values(value any) *Session {
	return o.NewSession().Values(value)
}

func (o *Orm) Insert(obj ...any) (int64, error) {
	return o.NewSession().Insert(obj...)
}

func (o *Orm) Update(obj ...any) (int64, error) {
	return o.NewSession().Update(obj...)
}

func (o *Orm) Delete(obj ...any) (int64, error) {
	return o.NewSession().Delete(obj...)
}

func (o *Orm) Begin() (*Session, error) {
	s := o.NewSession()

	return s, s.Begin()
}

func (o *Orm) Query(sql string, args []any) (*sql.Rows, error) {
	return o.NewSession().Query(sql, args)
}

func (o *Orm) QueryMap(sql string, args []any) ([]map[string]any, error) {
	return o.NewSession().QueryMap(sql, args)
}

func (o *Orm) QueryStruct(sql string, args []any, obj any) error {
	return o.NewSession().QueryStruct(sql, args, obj)
}
