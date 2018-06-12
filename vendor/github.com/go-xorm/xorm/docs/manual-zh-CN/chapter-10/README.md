## 事务处理
当使用事务处理时，需要创建Session对象。在进行事物处理时，可以混用ORM方法和RAW方法，如下代码所示：

```Go
session := engine.NewSession()
defer session.Close()
// add Begin() before any action
err := session.Begin()
user1 := Userinfo{Username: "xiaoxiao", Departname: "dev", Alias: "lunny", Created: time.Now()}
_, err = session.Insert(&user1)
if err != nil {
    session.Rollback()
    return
}
user2 := Userinfo{Username: "yyy"}
_, err = session.Where("id = ?", 2).Update(&user2)
if err != nil {
    session.Rollback()
    return
}

_, err = session.Exec("delete from userinfo where username = ?", user2.Username)
if err != nil {
    session.Rollback()
    return
}

// add Commit() after all actions
err = session.Commit()
if err != nil {
    return
}
```

* 注意如果您使用的是mysql，数据库引擎为innodb事务才有效，myisam引擎是不支持事务的。
