package orm

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"testing"

	"github.com/yrbb/rain/pkg/logger"
)

type TestModel struct {
	Id   int64  `db:"id primaryKey autoIncrement"`
	Key  int64  `db:"-"`
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

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	logger.SetLevel("debug")

	m.Run()
}

func TestQueryMap(t *testing.T) {
	res, err := testIns.NewSession().QueryMap(
		"select id, name, ft, sum(num) as sn from test where id<? group by id",
		[]any{10},
	)
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range res {
		fmt.Printf("%T, %T, %T, %T\n", v["id"], v["name"], v["ft"], v["sn"])
		t.Log(v["id"], v["name"], v["ft"], v["sn"])
	}
}

func TestQueryStruct(t *testing.T) {
	res := []struct {
		Id   int64   `json:"id"`
		Name string  `json:"name"`
		Ft   float64 `json:"ft"`
		Sn   int64   `json:"sn"`
	}{}
	err := testIns.NewSession().QueryStruct(
		"select id, name, ft, sum(num) as sn from test where id<? group by id",
		[]any{10},
		&res,
	)
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range res {
		t.Log(v.Id, v.Name, v.Ft, v.Sn)
	}
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

	res, err = testIns.Where("id", 2).Set("name", "lalaladda").Update(&TestModel{})
	t.Log(res, err)

	res, err = testIns.Table(&TestModel{}).Where("id", 3).Set("name", "hahashada").Update()
	t.Log(res, err)

	res, err = testIns.Update(&TestModel{
		Id:   1,
		Name: "lalalalala",
	})
	t.Log(res, err)
}

func TestDelete(t *testing.T) {
	res, err := testIns.Where("id", 12).Delete(&TestModel{})
	t.Log(res, err)

	res, err = testIns.Table(&TestModel{}).Where("id", 11).Delete()
	t.Log(res, err)

	res, err = testIns.Delete(&TestModel{Id: 13})
	t.Log(res, err)
}

func TestGet(t *testing.T) {
	tm := TestModel{Id: 1}
	res, err := testIns.Get(&tm)
	t.Log(res, err, tm)

	tm1 := TestModel{}
	res, err = testIns.Where("id", 1).Get(&tm1)
	t.Log(res, err, tm)

	tm2 := TestModel{}
	res, err = testIns.Where("id = ?", 2).Get(&tm2)
	t.Log(res, err, tm2)
}

func TestGetMap(t *testing.T) {
	res, err := testIns.Where("id", 1).GetMap(&TestModel{})
	t.Log(res, err)
}

func TestFind(t *testing.T) {
	tms := []TestModel{}
	res, err := testIns.Where("id", []int{1, 2}, "IN").Find(&tms)
	t.Log(res, err, tms)

	tms2 := []TestModel{}
	res, err = testIns.Where("id not in (?) and id = 4", []int{1, 2, 3}).Find(&tms2)
	t.Log(res, err, tms2)
}

func TestFindMap(t *testing.T) {
	res, err := testIns.Where("id", []int{1, 3}, "IN").FindMap(&TestModel{})
	t.Log(res, err)
}
