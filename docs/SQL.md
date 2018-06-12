# 按照时间统计局数

```
select uid as `ID`, name as `昵称`, count(*) as `局数` from rank where `record_at` > 1505480400 and uid > 0 group by uid order by `局数` desc;
```