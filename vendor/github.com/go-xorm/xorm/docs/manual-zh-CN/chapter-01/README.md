## 创建Orm引擎

在xorm里面，可以同时存在多个Orm引擎，一个Orm引擎称为Engine。Engine通过调用`xorm.NewEngine`生成，如：

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

一般情况下如果只操作一个数据库，只需要创建一个Engine即可。Engine是GoRutine安全的。

对于有大量数据并且需要分区的应用，也可以根据规则来创建多个Engine，比如：

```Go
var err error
for i:=0;i<5;i++ {
    engines[i], err = xorm.NewEngine("sqlite3", fmt.Sprintf("./test%d.db", i))
}
```

NewEngine传入的参数和`sql.Open`传入的参数完全相同，因此，在使用某个驱动前，请查看此驱动中关于传入参数的说明文档。以下为各个驱动的连接符对应的文档链接：

* [sqlite3](http://godoc.org/github.com/mattn/go-sqlite3#SQLiteDriver.Open)

* [mysql dsn](https://github.com/go-sql-driver/mysql#dsn-data-source-name)

* [mymysql](http://godoc.org/github.com/ziutek/mymysql/godrv#Driver.Open)

* [postgres](http://godoc.org/github.com/lib/pq)

在engine创建完成后可以进行一些设置，如：

1.调试，警告以及错误等显示设置，默认如下均为`false`

* `engine.ShowSQL = true`，则会在控制台打印出生成的SQL语句；
* `engine.ShowDebug = true`，则会在控制台打印调试信息；
* `engine.ShowError = true`，则会在控制台打印错误信息；
* `engine.ShowWarn = true`，则会在控制台打印警告信息；

2.如果希望将信息不仅打印到控制台，而是保存为文件，那么可以通过类似如下的代码实现，`NewSimpleLogger(w io.Writer)`接收一个io.Writer接口来将数据写入到对应的设施中。

```Go
f, err := os.Create("sql.log")
    if err != nil {
        println(err.Error())
        return
    }
engine.Logger = xorm.NewSimpleLogger(f)
```

3.engine内部支持连接池接口和对应的函数。

* 如果需要设置连接池的空闲数大小，可以使用`engine.SetMaxIdleConns()`来实现。
* 如果需要设置最大打开连接数，则可以使用`engine.SetMaxOpenConns()`来实现。
