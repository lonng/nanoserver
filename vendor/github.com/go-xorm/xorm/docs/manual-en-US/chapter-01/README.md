## Create ORM Engine 

When using xorm, you can create multiple orm engines, an engine means a databse. So you can：

```Go
import (
    _ "github.com/go-sql-driver/mysql"
    "github.com/go-xorm/xorm"
)
engine, err := xorm.NewEngine("mysql", "root:123@/test?charset=utf8")
defer engine.Close()
```

or

```Go
import (
    _ "github.com/mattn/go-sqlite3"
    "github.com/go-xorm/xorm"
    )
engine, err = xorm.NewEngine("sqlite3", "./test.db")
defer engine.Close()
```

You can create many engines for different databases.Generally, you just need create only one engine. Engine supports run on go routines.

xorm supports four drivers now:

* Mysql: [github.com/Go-SQL-Driver/MySQL](https://github.com/Go-SQL-Driver/MySQL)

* MyMysql: [github.com/ziutek/mymysql](https://github.com/ziutek/mymysql)

* SQLite: [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)

* Postgres: [github.com/lib/pq](https://github.com/lib/pq)

* MsSql: [github.com/lunny/godbc](https://githubcom/lunny/godbc)

NewEngine's parameters are the same as `sql.Open`. So you should read the drivers' document for parameters' usage.

After engine created, you can do some settings.

1.Logs

* `engine.ShowSQL = true`, Show SQL statement on standard output;
* `engine.ShowDebug = true`, Show debug infomation on standard output;
* `engine.ShowError = true`, Show error infomation on standard output;
* `engine.ShowWarn = true`, Show warnning information on standard output;

2.If want to record infomation with another method: use `engine.Logger` as `io.Writer`:

```Go
f, err := os.Create("sql.log")
    if err != nil {
        println(err.Error())
        return
    }
engine.Logger = xorm.NewSimpleLogger(f)
```

3.Engine provide DB connection pool settings.

* Use `engine.SetMaxIdleConns()` to set idle connections.
* Use `engine.SetMaxOpenConns()` to set Max connections. This methods support only Go 1.2+.
