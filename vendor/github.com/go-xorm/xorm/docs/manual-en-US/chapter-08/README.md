## 8.Execute SQL query

Of course, SQL execution is also provided.

If select then use Query

```Go
sql := "select * from userinfo"
results, err := engine.Query(sql)
```
