## 缓存

xorm内置了一致性缓存支持，不过默认并没有开启。要开启缓存，需要在engine创建完后进行配置，如：
启用一个全局的内存缓存

```Go
cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000)
engine.SetDefaultCacher(cacher)
```

上述代码采用了LRU算法的一个缓存，缓存方式是存放到内存中，缓存struct的记录数为1000条，缓存针对的范围是所有具有主键的表，没有主键的表中的数据将不会被缓存。
如果只想针对部分表，则：

```Go
cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000)
engine.MapCacher(&user, cacher)
```

如果要禁用某个表的缓存，则：

```Go
engine.MapCacher(&user, nil)
```

设置完之后，其它代码基本上就不需要改动了，缓存系统已经在后台运行。

当前实现了内存存储的CacheStore接口MemoryStore，如果需要采用其它设备存储，可以实现CacheStore接口。

不过需要特别注意不适用缓存或者需要手动编码的地方：

1. 当使用了`Distinct`,`Having`,`GroupBy`方法将不会使用缓存

2. 在`Get`或者`Find`时使用了`Cols`,`Omit`方法，则在开启缓存后此方法无效，系统仍旧会取出这个表中的所有字段。

3. 在使用Exec方法执行了方法之后，可能会导致缓存与数据库不一致的地方。因此如果启用缓存，尽量避免使用Exec。如果必须使用，则需要在使用了Exec之后调用ClearCache手动做缓存清除的工作。比如：

```Go
engine.Exec("update user set name = ? where id = ?", "xlw", 1)
engine.ClearCache(new(User))
```

缓存的实现原理如下图所示：

![cache design](https://raw.github.com/go-xorm/xorm/master/docs/cache_design.png)