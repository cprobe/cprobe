## 改造

cprobe 对 kafka 的监控是把 [kafka_exporter](https://github.com/danielqsj/kafka_exporter) 集成了进来，然后进行了一些改造，改造的点主要有：

- 统一化日志打印库，和 cprobe 主程序使用同一个日志库，方便日志的统一化
- 把命令行参数、环境变量参数、配置文件参数统一化，支持了配置文件切分管理
- 干掉了原来的 HTTP Server，毕竟 metrics 直接通过 remote write 方式发走了，不需要通过 `/metrics` 接口暴露了

## 集群监控

如果监控单实例：
```yaml
global:
  scrape_interval: 15s
  external_labels:
    cplugin: 'kafka'

scrape_configs:
  - job_name: 'kafka'
    static_configs:
      - targets:
          - '172.21.0.162:9092'
    scrape_rule_files:
      - 'rule.toml'
```

如果监控集群:
```yaml
global:
  scrape_interval: 15s
  external_labels:
    cplugin: 'kafka'

scrape_configs:
  - job_name: 'kafka'
    static_configs:
      - targets:
          - '172.21.0.162:9092,172.21.0.163:9092,172.21.0.164:9092'
    scrape_rule_files:
      - 'rule.toml'
```

如果监控多实例:
```yaml
global:
  scrape_interval: 15s
  external_labels:
    cplugin: 'kafka'

scrape_configs:
  - job_name: 'kafka'
    static_configs:
      - targets:
          - '172.21.0.162:9092,172.21.0.163:9092,172.21.0.164:9092'
          - '172.21.1.162:9092,172.21.1.163:9092,172.21.1.164:9092'
    scrape_rule_files:
      - 'rule.toml'
```

## 仪表盘

- 呈现kafka-exporter监控数据,用 [这个仪表盘](./dashboard/kafka-exporter.json)

## 告警规则

TODO

## 声明

cprobe 是一个缝合怪，类似 grafana-agent，相当于集成了众多 exporter 为一个二进制。Redis 的监控集成的是 [kafka-exporter](https://github.com/danielqsj/kafka_exporter)。更多使用方法请参考原项目的文档。

另外，本插件并没有其他文档，如果上面的信息不足以帮到你，你可能需要自行阅读源码了。当然，并非所有人都有能力阅读源码，所以欢迎大家提 PR 一起完善这个文档，这才是开源的正确协作模式。

