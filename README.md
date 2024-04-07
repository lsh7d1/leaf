## LEAF
号段模式发号器 

### 结构
|字段|类型|
|---|---|
| `biz_tag` | varchar(128) |
| `max_id` | bigint(20) |
| `step` | int(11) |
| `description` | varchar(256) |
| `update_time` | timestamp | 
| `created_time` | timestamp |
| `auto_clean` | tinyint(1) |

### Benchmark

当发号器step逼近1e6时
本机通过rpc调用，qps大约在1e6上下
```bash
goos: linux
goarch: amd64
pkg: leaf
cpu: AMD Ryzen 7 5800H with Radeon Graphics
BenchmarkMainInClient-16            8838            515490 ns/op            5168 B/op        103 allocs/op
PASS
ok      leaf    5.633s
```

### 每次扩容的操作会有一次慢SQL

```sql
2024/03/07 17:30:35 /root/leaf/dao/segment.go:41 SLOW SQL >= 200ms
[255.575ms] [rows:1] UPDATE `leaf_alloc` SET `max_id`=max_id + step WHERE `leaf_alloc`.`biz_tag` = 'bbbb'

2024/03/07 17:30:36 /root/leaf/dao/segment.go:70 SLOW SQL >= 200ms
[295.397ms] [rows:1] UPDATE `leaf_alloc` SET `max_id`=`leaf_alloc`.`max_id`+800 WHERE `leaf_alloc`.`biz_tag` = 'bbbb'

2024/03/07 17:30:36 /root/leaf/dao/segment.go:77 SLOW SQL >= 200ms
[264.477ms] [rows:1] SELECT * FROM `leaf_alloc` WHERE `leaf_alloc`.`biz_tag` = 'bbbb' ORDER BY `leaf_alloc`.`biz_tag` LIMIT 1

2024/03/07 17:30:44 /root/leaf/dao/segment.go:70 SLOW SQL >= 200ms
[283.596ms] [rows:1] UPDATE `leaf_alloc` SET `max_id`=`leaf_alloc`.`max_id`+6400 WHERE `leaf_alloc`.`biz_tag` = 'bbbb'

2024/03/07 17:30:45 /root/leaf/dao/segment.go:77 SLOW SQL >= 200ms
[282.894ms] [rows:1] SELECT * FROM `leaf_alloc` WHERE `leaf_alloc`.`biz_tag` = 'bbbb' ORDER BY `leaf_alloc`.`biz_tag` LIMIT 1

2024/03/07 17:30:47 /root/leaf/dao/segment.go:70 SLOW SQL >= 200ms
[254.794ms] [rows:1] UPDATE `leaf_alloc` SET `max_id`=`leaf_alloc`.`max_id`+12800 WHERE `leaf_alloc`.`biz_tag` = 'bbbb'

2024/03/07 17:31:14 /root/leaf/dao/segment.go:70 SLOW SQL >= 200ms
[422.032ms] [rows:1] UPDATE `leaf_alloc` SET `max_id`=`leaf_alloc`.`max_id`+51200 WHERE `leaf_alloc`.`biz_tag` = 'bbbb'

2024/03/07 17:31:15 /root/leaf/dao/segment.go:77 SLOW SQL >= 200ms
[284.877ms] [rows:1] SELECT * FROM `leaf_alloc` WHERE `leaf_alloc`.`biz_tag` = 'bbbb' ORDER BY `leaf_alloc`.`biz_tag` LIMIT 1

2024/03/07 17:32:24 /root/leaf/dao/segment.go:70 SLOW SQL >= 200ms
[277.370ms] [rows:1] UPDATE `leaf_alloc` SET `max_id`=`leaf_alloc`.`max_id`+204800 WHERE `leaf_alloc`.`biz_tag` = 'bbbb'

2024/03/07 17:35:17 /root/leaf/dao/segment.go:70 SLOW SQL >= 200ms
[301.234ms] [rows:1] UPDATE `leaf_alloc` SET `max_id`=`leaf_alloc`.`max_id`+819200 WHERE `leaf_alloc`.`biz_tag` = 'bbbb'

2024/03/07 17:35:34 /root/leaf/dao/segment.go:70 SLOW SQL >= 200ms
[263.514ms] [rows:1] UPDATE `leaf_alloc` SET `max_id`=`leaf_alloc`.`max_id`+819200 WHERE `leaf_alloc`.`biz_tag` = 'bbbb'
```