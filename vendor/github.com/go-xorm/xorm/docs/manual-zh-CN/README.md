# xorm

xorm是一个简单而强大的Go语言ORM库. 通过它可以使数据库操作非常简便。xorm的目标并不是让你完全不去学习SQL，我们认为SQL并不会为ORM所替代，但是ORM将可以解决绝大部分的简单SQL需求。xorm支持两种风格的混用。

## 特性

* 支持Struct和数据库表之间的灵活映射，并支持自动同步表结构

* 事务支持

* 同时支持原始SQL语句和ORM操作的混合执行

* 使用连写来简化调用

* 支持使用Id, In, Where, Limit, Join, Having, Table, Sql, Cols等函数和结构体等方式作为条件

* 支持级联加载Struct 

* 支持LRU内存缓存 和 Redis缓存

* 支持反转，即根据数据库自动生成xorm的结构体

* 事件支持

* 支持记录版本（即乐观锁）

## 驱动支持

xorm当前支持的驱动和数据库如下：

* Mysql: [github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)

* MyMysql: [github.com/ziutek/mymysql/godrv](https://github.com/ziutek/mymysql/godrv)

* SQLite: [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)

* Postgres: [github.com/lib/pq](https://github.com/lib/pq)

* MsSql: [github.com/denisenkom/go-mssqldb](https://github.com/denisenkom/go-mssqldb)

* MsSql: [github.com/lunny/godbc](https://github.com/lunny/godbc)

## 安装

推荐使用 [gopm](https://github.com/gpmgo/gopm) 进行安装： 

	gopm get github.com/go-xorm/xorm
	
或者您也可以使用go工具进行安装：

	go get github.com/go-xorm/xorm

## 文档

* [操作指南](http://xorm.io/docs)

* [GoWalker代码文档](http://gowalker.org/github.com/go-xorm/xorm)

* [Godoc代码文档](http://godoc.org/github.com/go-xorm/xorm)

## 讨论

请加入QQ群：280360085 进行讨论。

## 贡献

如果您也想为Xorm贡献您的力量，请查看 [CONTRIBUTING](https://github.com/go-xorm/xorm/blob/master/CONTRIBUTING.md)

## LICENSE

BSD License
[http://creativecommons.org/licenses/BSD/](http://creativecommons.org/licenses/BSD/)

