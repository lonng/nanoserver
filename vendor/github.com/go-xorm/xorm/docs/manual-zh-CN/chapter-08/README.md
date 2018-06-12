## 执行SQL查询

也可以直接执行一个SQL查询，即Select命令。在Postgres中支持原始SQL语句中使用 ` 和 ? 符号。

```Go
sql := "select * from userinfo"
results, err := engine.Query(sql)
```

当调用`Query`时，第一个返回值`results`为`[]map[string][]byte`的形式。
