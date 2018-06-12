## 插入数据

插入数据使用Insert方法，Insert方法的参数可以是一个或多个Struct的指针，一个或多个Struct的Slice的指针。

如果传入的是Slice并且当数据库支持批量插入时，Insert会使用批量插入的方式进行插入。

* 插入一条数据，此时可以用Insert或者InsertOne

```Go
user := new(User)
user.Name = "myname"
affected, err := engine.Insert(user)
// INSERT INTO user (name) values (?)
```

在插入单条数据成功后，如果该结构体有自增字段，则自增字段会被自动赋值为数据库中的id

```Go
fmt.Println(user.Id)
```

* 插入同一个表的多条数据，此时如果数据库支持批量插入，那么会进行批量插入，但是这样每条记录就无法被自动赋予id值。如果数据库不支持批量插入，那么就会一条一条插入。

```Go
users := make([]User, 0)
users[0].Name = "name0"
...
affected, err := engine.Insert(&users)
```

* 使用指针Slice插入多条记录，同上

```Go
users := make([]*User, 0)
users[0] = new(User)
users[0].Name = "name0"
...
affected, err := engine.Insert(&users)
```

* 插入多条记录并且不使用批量插入，此时实际生成多条插入语句，每条记录均会自动赋予Id值。

```Go
users := make([]*User, 0)
users[0] = new(User)
users[0].Name = "name0"
...
affected, err := engine.Insert(users...)
```

* 插入不同表的一条记录

```Go
user := new(User)
user.Name = "myname"
question := new(Question)
question.Content = "whywhywhwy?"
affected, err := engine.Insert(user, question)
```

* 插入不同表的多条记录

```Go
users := make([]User, 0)
users[0].Name = "name0"
...
questions := make([]Question, 0)
questions[0].Content = "whywhywhwy?"
affected, err := engine.Insert(&users, &questions)
```

* 插入不同表的一条或多条记录
```Go
user := new(User)
user.Name = "myname"
...
questions := make([]Question, 0)
questions[0].Content = "whywhywhwy?"
affected, err := engine.Insert(user, &questions)
```

这里需要注意以下几点：
* 这里虽然支持同时插入，但这些插入并没有事务关系。因此有可能在中间插入出错后，后面的插入将不会继续。
* 批量插入会自动生成`Insert into table values (),(),()`的语句，因此各个数据库对SQL语句有长度限制，因此这样的语句有一个最大的记录数，根据经验测算在150条左右。大于150条后，生成的sql语句将太长可能导致执行失败。因此在插入大量数据时，目前需要自行分割成每150条插入一次。