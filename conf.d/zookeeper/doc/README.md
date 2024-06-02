## 改造

cprobe 对 zookeeper 的监控是把 [zookeeper_exporter] 集成了进来，然后进行了一些改造，改造的点主要有：

- 统一化日志打印库，和 cprobe 主程序使用同一个日志库，方便日志的统一化
- 把命令行参数、环境变量参数、配置文件参数统一化，支持了配置文件切分管理
- 干掉了原来的 HTTP Server，毕竟 metrics 直接通过 remote write 方式发走了，不需要通过 `/metrics` 接口暴露了
### 参考文档
- [zookeeper指标采集服务1](https://github.com/carlpett/zookeeper_exporter)
- [zookeeper指标采集服务2](https://github.com/dln/zookeeper_exporter)
- [zookeeper指标采集服务](https://github.com/carlpett/zookeeper_exporter/blob/master/zookeeper.go)
- [zookeeper告警配置](https://zookeeper.apache.org/doc/current/zookeeperMonitor.html#Metrics)
- [指标列表](https://docs.datadoghq.com/integrations/zk/?tab=host)

## 集群监控

zookeeper 的集群监控，就是把集群里的每个组件（leader、follower、observer）当做一个普通的 redis 实例来对待。所以，只要把集群里的每个组件的 target 地址都配置到 cprobe 的抓取列表里即可。当然了，不同的集群，最好使用标签做区分，建议附加一个 cluster_name 的标签，比如：

```yaml
global:
  scrape_interval: 15s
  external_labels:
    cplugin: 'zookeeper'

scrape_configs:
- job_name: 'zookeeper'
  static_configs:
  - targets:
      - '127.0.0.1:2180'
      - '127.0.0.1:2181'
      - '127.0.0.1:2182'
    labels:
      cluster_name: "zk-cluster-01"
  scrape_rule_files:
  - 'rule.toml'
```

## 仪表盘

- 无（自行配置）

## 告警规则

- Prometheus 告警规则请参考 [这里](./alert/prom_alert_01.yaml)。

## 声明

cprobe 是一个缝合怪，类似 grafana-agent，相当于集成了众多 exporter 为一个二进制。Zookeeper 的监控集成的是 [zookeeper_exporter](https://github.com/carlpett/zookeeper_exporter)。更多使用方法请参考原项目的文档。

另外，本插件并没有其他文档，如果上面的信息不足以帮到你，你可能需要自行阅读源码了。当然，并非所有人都有能力阅读源码，所以欢迎大家提 PR 一起完善这个文档，这才是开源的正确协作模式。

