## 改造

核心是改造了 postgres_exporter 的 collector flag，之前是用命令行传参的方式来控制启用哪些 collector，现在改成了配置文件，如下：

```toml
enabled_collectors = [
    "database",
    "stat_database",
    "locks",
    "replication_slot",
    "replication",
    "stat_bgwriter",
    "stat_user_tables",
    "statio_user_tables",
    "wal",
    # "long_running_transactions",
    # "database_wraparound",
    # "postmaster",
    # "process_idle",
    # "stat_activity_autovacuum",
    # "stat_statements",
    # "stat_wal_receiver",
    # "statio_user_indexes",
    # "xlog_location",
]
```

要启用哪个 collector，就打开注释即可。

## 仪表盘

- [Grafana 仪表盘](./dash/grafana_postgres_01.json)

## 告警规则

- [Prometheus 告警规则](./alert/prom_alert_01.yaml)

## 声明

cprobe 是一个缝合怪，类似 grafana-agent，相当于集成了众多 exporter 为一个二进制。Postgres 的监控集成的是 [postgres_exporter](https://github.com/prometheus-community/postgres_exporter)。更多使用方法请参考原项目的文档。

另外，本插件并没有其他文档，如果上面的信息不足以帮到你，你可能需要自行阅读源码了。当然，并非所有人都有能力阅读源码，所以欢迎大家提 PR 一起完善这个文档，这才是开源的正确协作模式。

