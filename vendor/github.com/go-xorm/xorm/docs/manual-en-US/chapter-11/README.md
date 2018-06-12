## Built-in LRU memory cache provider

1. Global Cache
Xorm implements cache support. Defaultly, it's disabled. If enable it, use below code.

```Go
cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000)
engine.SetDefaultCacher(cacher)
```

If disable some tables' cache, then:

```Go
engine.MapCacher(&user, nil)
```

2. Table's Cache
If only some tables need cache, then:

```Go
cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000)
engine.MapCacher(&user, cacher)
```

Caution:

1. When use Cols methods on cache enabled, the system still return all the columns.

2. When using Exec method, you should clear cache：

```Go
engine.Exec("update user set name = ? where id = ?", "xlw", 1)
engine.ClearCache(new(User))
```

Cache implement theory below:

![cache design](https://raw.github.com/go-xorm/xorm/master/docs/cache_design.png)
