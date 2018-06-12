## 事件
xorm支持两种方式的事件，一种是在Struct中的特定方法来作为事件的方法，一种是在执行语句的过程中执行事件。

在Struct中作为成员方法的事件如下：

* BeforeInsert()

* BeforeUpdate()

* BeforeDelete()

* `func BeforeSet(name string, cell xorm.Cell)` 在 Get 或 Find 方法中，当数据已经从数据库查询出来，而在设置到结构体之前调用。

* AfterInsert()

* AfterUpdate()

* AfterDelete()

在语句执行过程中的事件方法为：

* Before(beforeFunc interface{})

* After(afterFunc interface{})

其中beforeFunc和afterFunc的原型为func(bean interface{}).
