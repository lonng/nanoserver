## 4.Insert data

Inserting records use Insert method. 

* Insert one record
```Go
user := new(User)
user.Name = "myname"
affected, err := engine.Insert(user)
```

After inseted, `user.Id` will be filled with primary key column value.
```Go
fmt.Println(user.Id)
```

* Insert multiple records by Slice on one table
```Go
users := make([]User, 0)
users[0].Name = "name0"
...
affected, err := engine.Insert(&users)
```

* Insert multiple records by Slice of pointer on one table
```Go
users := make([]*User, 0)
users[0] = new(User)
users[0].Name = "name0"
...
affected, err := engine.Insert(&users)
```

* Insert one record on two table.
```Go
user := new(User)
user.Name = "myname"
question := new(Question)
question.Content = "whywhywhwy?"
affected, err := engine.Insert(user, question)
```

* Insert multiple records on multiple tables.
```Go
users := make([]User, 0)
users[0].Name = "name0"
...
questions := make([]Question, 0)
questions[0].Content = "whywhywhwy?"
affected, err := engine.Insert(&users, &questions)
```

* Insert one or multple records on multiple tables.
```Go
user := new(User)
user.Name = "myname"
...
questions := make([]Question, 0)
questions[0].Content = "whywhywhwy?"
affected, err := engine.Insert(user, &questions)
```

Notice: If you want to use transaction on inserting, you should use session.Begin() before calling Insert.
