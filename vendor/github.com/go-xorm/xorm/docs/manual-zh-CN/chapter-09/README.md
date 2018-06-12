## 执行SQL命令

也可以直接执行一个SQL命令，即执行Insert， Update， Delete 等操作。此时不管数据库是何种类型，都可以使用 ` 和 ? 符号。

```Go
sql = "update `userinfo` set username=? where id=?"
res, err := engine.Exec(sql, "xiaolun", 1) 
```
