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

TODO

## 声明

cprobe 是一个缝合怪，类似 grafana-agent，相当于集成了众多 exporter 为一个二进制。MongoDB 的监控集成的是 [mongodb_exporter](https://github.com/percona/mongodb_exporter)。更多使用方法请参考原项目的文档。

另外，本插件并没有其他文档，如果上面的信息不足以帮到你，你可能需要自行阅读源码了。当然，并非所有人都有能力阅读源码，所以欢迎大家提 PR 一起完善这个文档，这才是开源的正确协作模式。
