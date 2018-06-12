## 删除数据

删除数据`Delete`方法，参数为struct的指针并且成为查询条件。

```Go
user := new(User)
affected, err := engine.Id(id).Delete(user)
```

`Delete`的返回值第一个参数为删除的记录数，第二个参数为错误。

注意：当删除时，如果user中包含有bool,float64或者float32类型，有可能会使删除失败。具体请查看 <a href="#160">FAQ</a>
