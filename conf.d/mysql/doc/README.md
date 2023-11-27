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

## 改造

由于不同的 exporter 打印日志的方式各异，配置文件的格式各异，命令行的参数各异，有些 exporter 是一对一的设计，即一个 exporter 采集一个实例，或者即便支持一对多，不同目标实例也只能使用完全相同的 exporter 配置，最终还是决定把 mysqld_exporter 的代码直接拷贝过来，然后进行改造。改造的点主要有：

- 统一化日志打印库，和 cprobe 主程序使用同一个日志库，方便日志的统一化
- 把命令行参数、环境变量参数、配置文件参数统一化
- 修改各个 Scraper 抽象，把原本直接引用命令行参数的地方改成结构体属性，避免并发问题
- 增加了自定义 SQL 语句的功能，可以通过配置文件指定自定义的 SQL 语句，然后把结果作为指标输出，这对于业务数据的监控尤其有用
- 增加了 mysql_instance_up 指标，用于标识数据库实例的连通性

作为普通用户上面的前 3 点不太理解也没关系，这 3 点说明主要是面向开发者的。自定义 SQL 的功能，这里做一下说明：每一个自定义 SQL 就是一个 `[[queries]]` 配置段，多个自定义 SQL 就是多个 `[[queries]]` 配置段。每个 `[[queries]]` 配置段包含以下几个属性：

- `mesurement`：指标名称前缀
- `metric_fields`：SQL 会查到多个字段，这里指定哪些字段作为指标输出，对应的字段的字段名作为指标后缀，字段值作为指标值
- `label_fields`：SQL 会查到多个字段，这里指定哪些字段作为标签输出
- `field_to_append`：SQL 会查到多个字段，这里指定哪个字段作为指标名称中缀
- `timeout`：SQL 执行超时时间
- `request`：SQL 语句

下面是一个例子：

```toml
[[queries]]
mesurement = "lock_wait"
metric_fields = [ "total" ]
timeout = "3s"
request = '''
SELECT count(*) as total FROM information_schema.innodb_trx WHERE trx_state='LOCK WAIT'
'''
```

自定义 SQL 功能，通常用于监控业务数据，当然，如果现在内置的性能指标不够用，也可以通过这个扩展机制来自定义 SQL 采集更多性能指标。

## 仪表盘

TODO

## 告警规则

TODO

## 声明

cprobe 是一个缝合怪，类似 grafana-agent，相当于集成了众多 exporter 为一个二进制。MySQL 的监控集成的是 [mysqld_exporter](https://github.com/prometheus/mysqld_exporter)。更多使用方法请参考原项目的文档。

另外，本插件并没有其他文档，如果上面的信息不足以帮到你，你可能需要自行阅读源码了。当然，并非所有人都有能力阅读源码，所以欢迎大家提 PR 一起完善这个文档，这才是开源的正确协作模式。

