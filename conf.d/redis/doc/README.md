## 改造

cprobe 对 redis 的监控是把 [redis_exporter](https://github.com/oliver006/redis_exporter) 集成了进来，然后进行了一些改造，改造的点主要有：

- 统一化日志打印库，和 cprobe 主程序使用同一个日志库，方便日志的统一化
- 把命令行参数、环境变量参数、配置文件参数统一化，支持了配置文件切分管理
- 干掉了原来的 HTTP Server，毕竟 metrics 直接通过 remote write 方式发走了，不需要通过 `/metrics` 接口暴露了

## 集群监控

redis 的集群监控，就是把集群里的每个组件（master、slave、sentinel）当做一个普通的 redis 实例来对待。所以，只要把集群里的每个组件的 target 地址都配置到 cprobe 的抓取列表里即可。当然了，不同的集群，最好使用标签做区分，建议附加一个 cluster_name 的标签，比如：

```yaml
global:
  scrape_interval: 15s
  external_labels:
    cplugin: 'redis'

scrape_configs:
- job_name: 'redis'
  static_configs:
  - targets:
    - '127.0.0.1:6479'
    - '127.0.0.1:6579'
    - '127.0.0.1:6679'
    - '127.0.0.1:26479'
    - '127.0.0.1:26579'
    - '127.0.0.1:26679'
    labels:
      cluster_name: "redis-cluster-01"
  scrape_rule_files:
  - 'rule.toml'
```

## 仪表盘

- 没有使用 redis 集群或者只有一个 redis 集群，用 [这个仪表盘](./dash/grafana_redis_01.json)
- 如果有多个 redis 集群，要为每套集群分别附加一个 cluster_name 的标签，用 [这个仪表盘](./dash/grafana_redis_02.json)

## 告警规则

TODO

## 声明

cprobe 是一个缝合怪，类似 grafana-agent，相当于集成了众多 exporter 为一个二进制。Redis 的监控集成的是 [redis_exporter](https://github.com/oliver006/redis_exporter)。更多使用方法请参考原项目的文档。

另外，本插件并没有其他文档，如果上面的信息不足以帮到你，你可能需要自行阅读源码了。当然，并非所有人都有能力阅读源码，所以欢迎大家提 PR 一起完善这个文档，这才是开源的正确协作模式。

