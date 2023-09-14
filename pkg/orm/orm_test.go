package orm

import (
	"log"
	"testing"
)

type TestModel struct {
	Model
	Id   int64  `db:"id primaryKey autoIncrement"`
	Name string `db:"name"`
}

func (t *TestModel) TableName() string {
	return "test"
}

var testIns *Orm

func TestMain(m *testing.M) {
	var err error
	testIns, err = New(&Config{
		Name:          "test",
		Type:          "mysql",
		Addr:          "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&interpolateParams=true",
		MaxIdleConns:  10,
		MaxOpenConns:  100,
		MaxLifeTime:   600,
		SlowThreshold: 1000,
		PoolThreshold: 80,
	})
	if err != nil {
		log.Fatalln(err)
	}

	m.Run()
}

func TestCreate(t *testing.T) {
	id, err := testIns.Insert(&TestModel{
		Name: "test create1",
	})
	t.Log(id, err)

	id, err = testIns.Values(&TestModel{
		Name: "test create2.1",
	}).Values(&TestModel{
		Name: "test create2.2",
	}).Insert()
	t.Log(id, err)

	id, err = testIns.Values([]map[string]any{
		{"name": "test create3"},
		{"name": "test create4"},
	}).Insert(&TestModel{})
	t.Log(id, err)
}

func TestUpdate(t *testing.T) {
	tm := TestModel{}
	_, _ = testIns.Where("id", 1).Get(&tm)
	tm.Name = "dadadada"
	res, err := testIns.Update(&tm)
	t.Log(res, err)

	res, err = testIns.Where("id", 2).Set("name", "lalalada").Update(&TestModel{})
	t.Log(res, err)

	res, err = testIns.Table(&TestModel{}).Where("id", 3).Set("name", "hahahada").Update()
	t.Log(res, err)
}

func TestSave(t *testing.T) {}

func TestDelete(t *testing.T) {
	res, err := testIns.Where("id", 10).Delete(&TestModel{})
	t.Log(res, err)

	res, err = testIns.Table(&TestModel{}).Where("id", 11).Delete()
	t.Log(res, err)
}

func TestGet(t *testing.T) {
	tm := TestModel{}
	res, err := testIns.Where("id", 1).Get(&tm)
	t.Log(res, err)
	t.Log(tm.Id, tm.Name)
}

func TestGetMap(t *testing.T) {
	nm1 := map[string]any{}
	res, err := testIns.Where("id", 1).GetMap(&TestModel{}, nm1)
	t.Log(res, err, nm1)
}

func TestFind(t *testing.T) {
	tms := []TestModel{}
	res, err := testIns.Where("id", []int{1, 2}, "IN").Find(&tms)
	t.Log(res, err, tms)
}

func TestFindMap(t *testing.T) {
	nm := []map[string]any{}
	res, err := testIns.Where("id", []int{1, 2}, "IN").FindMap(&TestModel{}, &nm)
	t.Log(res, err, nm)
}
