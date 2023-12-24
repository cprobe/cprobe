## 授权

```
# Authenticate as the admin user.
use admin
db.auth("admin", "<YOUR_MONGODB_ADMIN_PASSWORD>")

# Create the user for the Datadog Agent.
db.createUser({
  "user": "cprobe",
  "pwd": "<UNIQUEPASSWORD>",
  "roles": [
    { role: "read", db: "admin" },
    { role: "clusterMonitor", db: "admin" },
    { role: "read", db: "local" }
  ]
})
```

## 改动

- 去掉了 `collect-all` 的选项，默认把所有 collector 都打开了，通过配置文件统一控制
- 去掉了 global connection pool，在 cprobe 场景下每次就用短连接就好
- 日志库统一化处理，原本的命令行参数全部改成配置文件 rule.toml 里的配置项

## 配置

main.yaml 配置举例：

```yaml
scrape_configs:
- job_name: 'standalone'
  static_configs:
  - targets:
    - 10.99.1.110:27017
  scrape_rule_files:
  - 'rule.toml'
```

targets 下面配置的就是原本 mongodb_exporter 里的 `mongodb.uri` 参数指定的连接地址。在 cprobe 下，可以很方便指定多个 target。

## 仪表盘

- [Grafana Dashboard](./dash/grafana_mongodb_01.json)

## 告警规则

```
# 连接 mongodb 实例失败
mongodb_cprobe_up == 0

# 可用连接小于 1000
mongodb_ss_connections{conn_type="available"} < 1000

# 延迟超过 100 毫秒
avg by (job, instance, type) (rate(mongodb_mongod_op_latencies_latency_total[5m]) / (rate(mongodb_mongod_op_latencies_ops_total[5m]) > 0)) / 1000 > 100

# 有些集群成员的健康状态不正常
mongodb_members_health < 1

# secondary 落后 primary 太多
mongodb_mongod_replset_member_replication_lag{state="SECONDARY"} > 30

# 读请求队列堆积
mongodb_mongod_global_lock_current_queue{type="reader"} > 100

# 写请求队列堆积
mongodb_mongod_global_lock_current_queue{type="writer"} > 100

# page fault 太多
rate(mongodb_extra_info_page_faults_total[1m]) > 30

# 未设置超时的游标数量
mongodb_mongod_metrics_cursor_open{state="noTimeout"} > 0

# 出现 Message asserts
rate(mongodb_asserts_total{type="msg"}[1m]) > 0

# 出现 Regular asserts
rate(mongodb_asserts_total{type="warning"}[1m]) > 0
```

## 声明

cprobe 是一个缝合怪，类似 grafana-agent，相当于集成了众多 exporter 为一个二进制。MongoDB 的监控集成的是 [mongodb_exporter](https://github.com/percona/mongodb_exporter)。更多使用方法请参考原项目的文档。

另外，本插件并没有其他文档，如果上面的信息不足以帮到你，你可能需要自行阅读源码了。当然，并非所有人都有能力阅读源码，所以欢迎大家提 PR 一起完善这个文档，这才是开源的正确协作模式。
