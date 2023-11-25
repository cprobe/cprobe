## 授权

要监控 MySQL，实际就是连上去执行一些命令，比如 `show global status`，`show global variables` 等等。所以，需要一个用户，这个用户需要有足够的权限。下面是创建用户并授权的 SQL 语句：

```sql
CREATE USER 'cprobe'@'cprobe-server-ip' IDENTIFIED BY 'cProbePa55' WITH MAX_USER_CONNECTIONS 3;
GRANT PROCESS, REPLICATION CLIENT, SELECT ON *.* TO 'cprobe'@'cprobe-server-ip';
```

- 上面的 `cprobe` 是用户名，一般监控账号都是单独一个账号，所以创建一个单独的 `cprobe` 账号是比较好的选择。
- `cprobe-server-ip` 是 cprobe 进程所在的服务器的 IP 地址，改成你的环境的实际 IP 地址即可，如果不想对 IP 做限制，可以改成 `%`。
- `cProbePa55` 是密码，改成你自己喜欢的密码即可。
- 另外建议，所有的数据库在初始化的时候，都应该创建一个统一的账号密码用于监控，可以大幅降低运维成本。
- 创建账号时，最好限制一下最大连接数，避免账号被滥用导致数据库压力过大。不过，这个限制并非所有的数据库版本都支持。

## 声明

cprobe 是一个缝合怪，类似 grafana-agent，相当于集成了众多 exporter 为一个二进制。MySQL 的监控集成的是 [mysqld_exporter](https://github.com/prometheus/mysqld_exporter)。更多使用方法请参考原项目的文档。

另外，本插件并没有其他文档，如果上面的信息不足以帮到你，你可能需要自行阅读源码了。当然，并非所有人都有能力阅读源码，所以欢迎大家提 PR 一起完善这个文档，这才是开源的正常协作模式。

