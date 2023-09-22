# 快速开始

* 第一步

```Go
db, err := orm.New(&orm.Config{
    Name: "default",
    Addr: "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8",
})
```

* 定义表结构体

```Go
type User struct {	
    Id int64        `db:"id primaryKey autoIncrement"` // primary key auto_increment
    Name string     `db:"name uniqueKey"`              // unique key
    NickName string `db:"nick_name"`                   // 指定字段名
    Ignore  int     `db:"-"`                           // 忽略字段
    Balance int     `db:"balance"`
    Created int64   `db:"created"`
    Updated int64   `db:"updated"`
}
```

* 查询

```Go
// SELECT * FROM user WHERE id = ? LIMIT 1
find, err := db.Where("id", id).Get(&user)

// SELECT name FROM user WHERE id = ? LIMIT 1
find, err := db.Columns("name").Where("id", id).Get(&user)

// SELECT name FROM user WHERE id = ? LIMIT 1 (append to map)
res := map[string]any{}
find, err := db.Columns("name").Where("id", id).Get(&user, &res)

// SELECT * FROM user WHERE id < ?
users := []User{}
find, err := db.Where("id", id, "<").Find(&users)  
// or
find, err := db.Where("id", "<", id).Find(&users)  
// or
find, err := db.Where("id < ?", id).Find(&users)  

// SELECT * FROM user WHERE id IN (?, ?, ?)
users := []User{}
find, err := db.Where("id", []int{1,2,3}, "IN").Find(&users)
// or 
find, err := db.Where("id", "IN", []int{1,2,3}).Find(&users)
// or 
find, err := db.Where("id IN (?)", []int{1,2,3}).Find(&users)

// SELECT * FROM user WHERE id = ? OR id = ?
users := []User{}
find, err := db.Where("id", 1).OrWhere("id", 2).Find(&users)  

// SELECT * FROM user WHERE (id = ? OR id = ?) AND name LIKE ?
users := []User{}
find, err := db.Bracket(func (s *lemon.Session) {
	s.Where("id", id_1).OrWhere("id", id_2)
}).Where("name", name, "like").Find(&users)
```

* 插入

```Go
insertId, err := db.Values(&user).Insert()

insertId, err := db.Insert(&user)

insertId, err := db.Values(&map[string]any{
    "name": "v_name",
    "nickName": "v_nick_name",
}).Insert()

insertId, err := db.Insert(&user, &map[string]any{
    "name": "v_name",
    "nickName": "v_nick_name",
})
```

* 更新

```Go
// UPDATE user SET name = ? Where id = ?
affected, err := db.Where("id", 1).Set("name", name).Update(&user)

// UPDATE user SET name = ? Where id = ?
affected, err := db.Where("id", id).Update(&user, &map[string]any{"name": "xxx"})

// UPDATE user SET name = ? Where id = ?
affected, err := db.Where("id", id).Set(&map[string]any{"name": "xxx"}).Update(&user)

// UPDATE user SET balance = balance + 1 Where id = ?
affected, err := db.Where("id", id).SetRaw("balance", "balance + 1").Update(&user)

// UPDATE user SET ... Where id = ?
affected, err := db.Where("id", id).Update(&user)

// UPDATE user SET name = ? Where id = ?
user := User{}
find, err := db.Where("id", id).Get(user)
user.Name = "other name"
affected, err := db.Save(&user)

// UPDATE user SET name = ?, balance = ? Where id = ?
affected, err := db.Save(&user, &map[string]any{
    "name": "other name"
    "balance": 100
})

```

* 删除

```Go
// DELETE FROM user Where id = ?
affected, err := db.Where("id", 1).Delete(&user)
// or 
affected, err := db.Delete(&User{Id: 1})
```
